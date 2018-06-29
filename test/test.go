package main

import (
	"fmt"
	// "glocust"
)

var opthons struct {
}

type BB struct {
	aa   int
	name string
}

func (p *BB) Name() string {
	println(p.name)
	return ""
}

// func (p *BB) Call() {
// 	println("Sdfsdfsdfsdf")
// }

type CC struct {
	name string
}

type QQ struct {
	BB
	CC
}

func (p *CC) School() {
	fmt.Println(p.name)
}

func (p *CC) Name() string {
	fmt.Println(p.name)
	return ""
}

func testFunc(f func() string) *bool {
	result := false
	defer func() {

		result = true

	}()

	f()
	println("resut")
	return &result
}

func main() {
	// glocust.Now()

	a := BB{1, "222"}
	// println()
	// println(BB)
	c := testFunc(a.Name)
	println("main return")
	println(*c)

}
