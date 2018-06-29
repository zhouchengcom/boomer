package glocust

type Locust interface {
	Min() int
	Max() int
	Tasks() []Task
	OnStart()
	CatchExceptions() bool
}

type Task struct {
	Weight int
	Fn     func()
	Name   string
}
