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

func (t *Tasks) Min() int {
	return 0
}

func (t *Tasks) Max() int {
	return 0
}

func (t *Tasks) Tasks() []glocust.Task {
	return nil
}

func newLocust() glocust.Locust {
	t := Tasks{}
	return &t
}

func (t *Tasks) OnStart() {

}

func (t *Tasks) CatchExceptions() bool {
	return true
}

func (t *Tasks) foo() {
	start := glocust.Now()
	time.Sleep(100 * time.Millisecond)
	elapsed := glocust.Now() - start
	t.index++

	glocust.Events.Publish("request_success", "http", "foo", elapsed, int64(10))
}

func main() {
	// glocust.Now()
	glocust.Run(newLocust)
}
