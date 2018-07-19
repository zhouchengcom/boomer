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

// func do(i interface{}) {
// 	switch v := i.(type) {
// 	case int, int8, int16, int32, int64:
// 		fmt.Printf("Twice %v is %v\n", v, v*2)
// 		// workers = int(clients.(t))
// 	case uint, uint8, uint16, uint32, uint64:
// 		// workers = int(t)
// 		println(v)
// 	}
// }

func main() {
	glocust.Run(newLocust)
	// println(time.Now().Unix())

}
