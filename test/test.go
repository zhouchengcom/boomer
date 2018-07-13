package main

import (
	"glocust"
	"time"
)

// "glocust"

// "glocust"

type Tasks struct {
	glocust.LocustMate
	index int
}

func newLocust() glocust.Locust {
	t := Tasks{}
	t.MinWait = 100
	t.MaxWait = 101
	t.AddTask(1, t.foo)
	t.AddTask(2, t.boo)
	return &t
}

func (t *Tasks) foo() {
	start := glocust.Now()
	time.Sleep(100 * time.Millisecond)
	elapsed := glocust.Now() - start
	t.index++
	// println(t.index)
	glocust.RequestSuccess("http", "foo", elapsed, int64(10))
}

func (t *Tasks) boo() {
	start := glocust.Now()
	time.Sleep(100 * time.Millisecond)
	elapsed := glocust.Now() - start
	t.index++
	// println(t.index)
	glocust.RequestSuccess("http", "boo", elapsed, int64(10))
}

func main() {
	glocust.Run(newLocust)
}
