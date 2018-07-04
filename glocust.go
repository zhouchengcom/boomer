package glocust

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// Run accepts a slice of Task and connects
// to a locust master.
func Run(newLocust func() Locust) {

	kingpin.Parse()
	if *options.slave == false {
		runLocal(newLocust)

	} else {
		runDistributed(newLocust)
	}
}

func runDistributed(newLocust func() Locust) {
	var r *runner
	client := newClient()
	r = &runner{
		newLocust: newLocust,
		client:    client,
		nodeID:    getNodeID(),
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

func runLocal(newLocust func() Locust) {

	var r *runner
	r = &runner{
		newLocust: newLocust,
		nodeID:    getNodeID(),
	}

	Events.Subscribe("boomer:quit", r.onQuiting)

	r.getReady()
	r.startHatching(1, 1)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	<-c
	Events.Publish("boomer:quit")

	// wait for quit message is sent to master
	log.Println("shut down")
}
