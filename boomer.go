package glocust

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Run accepts a slice of Task and connects
// to a locust master.
func Run(tasks ...*Task) {

	if !flag.Parsed() {
		flag.Parse()
	}

	var r *runner
	client := newClient()
	r = &runner{
		tasks:  tasks,
		client: client,
		nodeID: getNodeID(),
	}

	Events.Subscribe("boomer:quit", r.onQuiting)

	r.getReady()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	<-c
	Events.Publish("boomer:quit")

	// wait for quit message is sent to master
	<-disconnectedFromMaster
	log.Println("shut down")

}
