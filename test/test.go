package main

import (
	"glocust"
	"time"
)

// "glocust"

type Tasks struct {
	glocust.LocustMate
	index int
}

func newLocust() glocust.Locust {
	t := Tasks{}
	t.AddTask(1, t.foo)
	return &t
}

func (t *Tasks) foo() {
	start := glocust.Now()
	time.Sleep(100 * time.Millisecond)
	elapsed := glocust.Now() - start
	t.index++

	glocust.Events.Publish("request_success", "http", "foo", elapsed, int64(10))
}

func main() {
	log
	// glocust.Run(newLocust)
}
