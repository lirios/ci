package service

import (
	"fmt"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
	cronService "gopkg.in/robfig/cron.v2"
)

var triggers map[string]struct{}

type Executor struct {
	cron     *cronService.Cron
	settings *Settings
	notifier *Notifier
	jobList  *JobList
	taskList *TaskList
	runList  *RunList
	entries  map[string]cronService.EntryID
}

func NewExecutor(settings *Settings, notifier *Notifier, jobList *JobList, taskList *TaskList, runList *RunList) *Executor {
	cron := cronService.New()
	cron.Start()
	return &Executor{
		cron,
		settings,
		notifier,
		jobList,
		taskList,
		runList,
		make(map[string]cronService.EntryID),
	}
}

func (e Executor) ArmTrigger(t Trigger) {
	entryID, err := e.cron.AddFunc(t.Schedule, func() { e.findAndRun(t) })
	if err == nil {
		e.entries[t.Name] = entryID
	} else {
		fmt.Printf("Error arming trigger %s: %v", t.Name, err)
	}
}

func (e Executor) DisarmTrigger(name string) {
	e.cron.Remove(e.entries[name])
	delete(e.entries, name)
	println("Trigger has been removed")
}

// Walks through each job, seeing if the trigger who's turn it is to execute is attached. Executes those jobs.
func (e Executor) findAndRun(t Trigger) {
	jobs := e.jobList.GetJobsWithTrigger(t.ID())
	for _, job := range jobs {
		println("Executing job " + job.Name)
		e.runnit(job)
	}
}

// Gathers the tasks attached to the given job and executes them.
func (e Executor) runnit(j Job) {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	var tasks []Task
	for _, taskName := range j.Tasks {
		task, err2 := e.taskList.Get(taskName)
		if err2 != nil {
			panic(err2)
		}
		t := task.(Task)
		tasks = append(tasks, t)
	}
	err = e.runList.AddRun(id.String(), filepath.Join(e.settings.Server.OutputPath, "files", "logs"), j, tasks)
	if err != nil {
		panic(err)
	}
}
