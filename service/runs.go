package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Result struct {
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Task        Task      `json:"task"`
	LogPath     string    `json:"logpath"`
	LogFileName string    `json:"logfilename"`
	Error       string    `json:"error"`
}

type Run struct {
	UUID    string    `json:"uuid"`
	Job     Job       `json:"job"`
	Tasks   []Task    `json:"tasks"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Results []*Result `json:"results"`
	Status  string    `json:"status"`
}

func (r Run) ID() string {
	return r.UUID
}

type RunList struct {
	list
	notifier *Notifier
	jobList  *JobList
}

func NewRunList(rootPath string, notifier *Notifier, jobList *JobList) *RunList {
	return &RunList{
		list{elements: []elementer{}, fileName: filepath.Join(rootPath, runsFile)},
		notifier,
		jobList,
	}
}

func (l *RunList) Load() {
	bytes := readFile(l.fileName)
	var runs []Run
	err := json.Unmarshal([]byte(string(bytes)), &runs)
	if err != nil {
		panic(err)
	}
	l.elements = nil
	for _, run := range runs {
		l.elements = append(l.elements, run)
	}
}

func (j *RunList) Len() int {
	j.RLock()
	defer j.RUnlock()

	return len(j.elements)
}

func (l *RunList) Less(i, j int) bool {
	l.RLock()
	defer l.RUnlock()

	return l.elements[i].(Run).Start.Before(l.elements[j].(Run).Start)
}

func (l *RunList) Swap(i, j int) {
	l.RLock()
	defer l.RUnlock()

	l.elements[i], l.elements[j] = l.elements[j], l.elements[i]
}

func (l *RunList) GetRecent(offset, length int) []elementer {
	runs := l.elements
	if offset != -1 {
		if offset >= len(runs) {
			return nil
		}
		if length != -1 && offset+length < len(runs) {
			runs = runs[offset : offset+length]
		} else {
			runs = runs[offset:]
		}
	} else {
		if length != -1 {
			runs = runs[:length]
		}
	}
	return runs
}

func (j *RunList) AddRun(UUID string, logRootPath string, job Job, tasks []Task) error {
	run := Run{UUID: UUID, Job: job, Tasks: tasks, Start: time.Now(), Status: "New"}
	// check to make sure that UUID doesn't already exist
	var found bool = false
	for _, j := range j.elements {
		if run.UUID == j.(Run).UUID {
			found = true
		}
	}
	if found {
		return errors.New("Run with that name found in list")
	}
	j.Lock()
	defer j.Unlock()

	// add the run to the list and execute
	j.elements = append(j.elements, run)
	logPath := filepath.Join(logRootPath, run.ID())
	os.MkdirAll(logPath, os.ModePerm)
	go j.execute(logPath, &run)
	j.save()
	return nil
}

func (l *RunList) execute(logPath string, r *Run) {
	r.Status = "Running"
	for _, task := range r.Tasks {
		result := &Result{Start: time.Now(), LogPath: logPath, LogFileName: task.ID() + ".log", Task: task}
		r.Results = append(r.Results, result)
		l.Update(*r)
		shell, commandArg := getShell()
		cmd := exec.Command(shell, commandArg, task.Script)

		cmd.Env = append(cmd.Env, "LIRICI_UUID="+r.UUID)
		cmd.Env = append(cmd.Env, "LIRICI_JOB_NAME="+r.Job.ID())
		cmd.Env = append(cmd.Env, "LIRICI_TASK_NAME="+task.Name)
		cmd.Env = append(cmd.Env, "LIRICI_OUTPUT_DIR="+l.notifier.settings.Server.OutputPath)

		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			reportRunError(l, r, result, err)
			return
		}
		errPipe, err := cmd.StderrPipe()
		if err != nil {
			outPipe.Close()
			reportRunError(l, r, result, err)
			return
		}
		var outputWg sync.WaitGroup
		outputWg.Add(1)
		go result.muxIntoOutput(outPipe, errPipe, &outputWg)

		if err2 := cmd.Start(); err2 != nil {
			reportRunError(l, r, result, err2)
			return
		}
		outputWg.Wait()
		if err2 := cmd.Wait(); err2 != nil {
			reportRunError(l, r, result, err2)
			return
		}
		result.End = time.Now()
		if err != nil {
			reportRunError(l, r, result, err)
			return
		}
		l.Update(*r)
	}
	r.End = time.Now()
	r.Status = "Done"
	job, err := l.jobList.Get(r.Job.Name)
	if err != nil {
		return
	}
	j := job.(Job)
	j.Status = "Ok"
	l.jobList.Update(job)
	l.Update(*r)
	l.notifier.Queue <- r
}

func (result *Result) muxIntoOutput(stdout io.ReadCloser, stderr io.ReadCloser, done *sync.WaitGroup) {
	defer done.Done()
	file, err := os.Create(filepath.Join(result.LogPath, result.LogFileName))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	outLines := consumeLines(stdout)
	errLines := consumeLines(stderr)
	for outLines != nil || errLines != nil {
		select {
		case line, ok := <-outLines:
			if ok {
				file.WriteString(line + "\n")
			} else {
				outLines = nil
			}
		case line, ok := <-errLines:
			if ok {
				file.WriteString(line + "\n")
			} else {
				errLines = nil
			}
		}
	}
}

func consumeLines(reader io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer reader.Close()
		defer close(lines)
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			lines <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Println("reading output: ", err)
		}
	}()
	return lines
}

func reportRunError(l *RunList, r *Run, result *Result, err error) {
	log.Println("Reporting error", err)
	result.Error = err.Error()
	r.Status = "Failed"
	r.End = time.Now()
	l.Update(*r)
	job, err := l.jobList.Get(r.Job.Name)
	if err != nil {
		return
	}
	j := job.(Job)
	j.Status = "Failing"
	l.jobList.Update(job)
	l.notifier.Queue <- r
	return
}

func getShell() (string, string) {
	var shell = os.Getenv("SHELL")
	if shell == "" {
		if _, err := os.Stat("/bin/bash"); err == nil {
			shell = "/bin/bash"
		}
	}
	if shell != "" {
		return shell, "-c"
	}
	return "cmd", "/C"
}
