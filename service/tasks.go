package service

import (
	"encoding/json"
	"path/filepath"
)

type Task struct {
	Name   string `json:"name"`
	Script string `json:"script"`
}

func (t Task) ID() string {
	return t.Name
}

type TaskList struct {
	list
}

func NewTaskList(rootPath string) *TaskList {
	return &TaskList{
		list{elements: []elementer{}, fileName: filepath.Join(rootPath, tasksFile)},
	}
}

func (l *TaskList) Load() {
	bytes := readFile(l.fileName)
	var tasks []Task
	err := json.Unmarshal([]byte(string(bytes)), &tasks)
	if err != nil {
		panic(err)
	}
	l.elements = []elementer{}
	for _, task := range tasks {
		l.elements = append(l.elements, task)
	}
}
