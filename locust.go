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
}

type LocustMate struct {
	min             int
	max             int
	tasks           []Task
	catchExceptions bool
}

func (l *LocustMate) Min() int {
	return l.min
}

func (l *LocustMate) Max() int {
	return l.max
}

func (l *LocustMate) Tasks() []Task {
	return l.tasks
}

func (l *LocustMate) OnStart() {

}

func (l *LocustMate) AddTask(weight int, fn func()) {
	task := Task{Weight: weight, Fn: fn}
	l.tasks = append(l.tasks, task)
}

func (l *LocustMate) CatchExceptions() bool {
	return l.catchExceptions
}
