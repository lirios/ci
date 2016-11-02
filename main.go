package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	. "github.com/lirios/ci/service"
	"gopkg.in/gcfg.v1"
)

var routes = []struct {
	route   string
	handler func(context, http.ResponseWriter, *http.Request) (int, interface{})
	method  string
}{
	{"/jobs", listJobs, "GET"},
	{"/jobs", addJob, "POST"},
	{"/jobs/{job}", getJob, "GET"},
	{"/jobs/{job}", deleteJob, "DELETE"},
	{"/jobs/{job}/tasks", addTaskToJob, "POST"},
	{"/jobs/{job}/tasks/{task}", removeTaskFromJob, "DELETE"},
	{"/jobs/{job}/triggers/", addTriggerToJob, "POST"},
	{"/jobs/{job}/triggers/{trigger}", removeTriggerFromJob, "DELETE"},

	{"/tasks", listTasks, "GET"},
	{"/tasks", addTask, "POST"},
	{"/tasks/{task}", getTask, "GET"},
	{"/tasks/{task}", updateTask, "PUT"},
	{"/tasks/{task}", deleteTask, "DELETE"},
	{"/tasks/{task}/jobs", listJobsForTask, "GET"},

	{"/runs", listRuns, "GET"},
	{"/runs", addRun, "POST"},
	{"/runs/{run}", getRun, "GET"},

	{"/triggers", listTriggers, "GET"},
	{"/triggers", addTrigger, "POST"},
	{"/triggers/{trigger}", getTrigger, "GET"},
	{"/triggers/{trigger}", updateTrigger, "PUT"},
	{"/triggers/{trigger}", deleteTrigger, "DELETE"},
	{"/triggers/{trigger}/jobs", listJobsForTrigger, "GET"},
}

type ctx struct {
	settings    *Settings
	hub         *Hub
	executor    *Executor
	jobList     *JobList
	taskList    *TaskList
	triggerList *TriggerList
	runList     *RunList
}

func (t ctx) Settings() *Settings {
	return t.settings
}

func (t ctx) Hub() *Hub {
	return t.hub
}

func (t ctx) Executor() *Executor {
	return t.executor
}

func (t ctx) JobList() *JobList {
	return t.jobList
}

func (t ctx) TaskList() *TaskList {
	return t.taskList
}

func (t ctx) TriggerList() *TriggerList {
	return t.triggerList
}

func (t ctx) RunList() *RunList {
	return t.runList
}

type context interface {
	Settings() *Settings
	Hub() *Hub
	Executor() *Executor
	JobList() *JobList
	TaskList() *TaskList
	TriggerList() *TriggerList
	RunList() *RunList
}

type appHandler struct {
	*ctx
	handler func(context, http.ResponseWriter, *http.Request) (int, interface{})
}

func (t appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, data := t.handler(t.ctx, w, r)
	marshal(data, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
}

func main() {
	wd, _ := os.Getwd()
	log.Println("Working directory", wd)

	// Load settings
	var settingsPath string = "./config.ini"
	if len(os.Args) > 1 {
		settingsPath = os.Args[1:][0]
	}
	var settings Settings
	err := gcfg.ReadFileInto(&settings, settingsPath)
	if err != nil {
		panic(err)
	}

	// Create directories
	os.MkdirAll(settings.Server.DbRootPath, os.ModePerm)
	os.MkdirAll(filepath.Join(settings.Server.OutputPath, "files", "logs"), os.ModePerm)

	jobList := NewJobList(settings.Server.DbRootPath)
	taskList := NewTaskList(settings.Server.DbRootPath)
	triggerList := NewTriggerList(settings.Server.DbRootPath)
	runList := NewRunList(settings.Server.DbRootPath, jobList)

	jobList.Load()
	taskList.Load()
	triggerList.Load()
	runList.Load()

	hub := NewHub(runList)
	go hub.HubLoop()

	executor := NewExecutor(&settings, jobList, taskList, runList)

	appContext := &ctx{&settings, hub, executor, jobList, taskList, triggerList, runList}

	r := mux.NewRouter()

	// non REST routes
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir("web/")))
	r.PathPrefix("/files/").Handler(http.FileServer(http.Dir(settings.Server.OutputPath)))
	r.HandleFunc("/", app).Methods("GET")
	r.Handle("/ws", appHandler{appContext, wsHandler}).Methods("GET")

	for _, detail := range routes {
		r.Handle(detail.route, appHandler{appContext, detail.handler}).Methods(detail.method)
	}

	log.Println("Running on " + settings.Server.Port)
	http.ListenAndServe(settings.Server.Port, r)
}
