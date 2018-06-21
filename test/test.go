package main

import (
	"flag"
)

var ss struct {
	bb =1
}

func main() {
	if flag.Parsed() {
		flag.Parse()
	}

}
