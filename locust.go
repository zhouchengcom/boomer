package glocust

type Locust interface {
	Min() int
	Max() int
	Tasks() []Task
	WeightsTasks() []Task
	OnStart()
	CatchExceptions() bool
}

type Task struct {
	Weight int
	Fn     func()
}

type LocustMate struct {
	MinWait         int
	MaxWait         int
	TaskSet         []Task
	Catchexceptions bool
}

func (l *LocustMate) Min() int {
	return l.MinWait
}

func (l *LocustMate) Max() int {
	return l.MaxWait
}

func (l *LocustMate) Tasks() []Task {
	return l.TaskSet
}

func (l *LocustMate) OnStart() {

}

func (l *LocustMate) AddTask(weight int, fn func()) {
	task := Task{Weight: weight, Fn: fn}
	l.TaskSet = append(l.TaskSet, task)
}

func (l *LocustMate) CatchExceptions() bool {
	return l.Catchexceptions
}

func (l *LocustMate) WeightsTasks() []Task {
	var tasks []Task
	for _, v := range l.TaskSet {
		for i := 0; i < v.Weight; i++ {
			tasks = append(tasks, v)
		}
	}
	return tasks
}
