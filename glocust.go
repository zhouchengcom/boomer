package glocust

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// Run accepts a slice of Task and connects
// to a locust master.
func Run(newLocust func() Locust) {

	kingpin.Parse()

	if options.csvFilebase != nil {
		stats.createResultFile(options.csvFilebase)
	}
	defer stats.closeResultFile()

	if *options.slave == false {
		runLocal(newLocust)

	} else {
		runDistributed(newLocust)
	}
}

func runDistributed(newLocust func() Locust) {
	r := newRunner(newLocust, true)
	Events.Subscribe("boomer:quit", r.onQuiting)

	r.getReady()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	if *options.runTime != 0 {
		go func() {
			time.Sleep(time.Duration(*options.runTime) * time.Second)
			log.Println("Time limit reached. Stopping Locust.")
			Events.Publish("boomer:quit")
			r.stop()
		}()
	}

	r.wait()

	Events.Publish("boomer:quit")

	// wait for quit message is sent to master
	<-disconnectedFromMaster
	log.Println("shut down")
}

func runLocal(newLocust func() Locust) {

	r := newRunner(newLocust, false)
	println(r)
	Events.Subscribe("boomer:quit", r.onQuiting)

	r.getReady()
	r.startHatching(*options.numClients, *options.hatchRate)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	if *options.runTime != 0 {
		go func() {
			time.Sleep(time.Duration(*options.runTime) * time.Second)
			log.Println("Time limit reached. Stopping Locust.")
			Events.Publish("boomer:quit")
			r.stop()
			r.waitTaskFinish()
		}()
	}
	go func() {
		<-c
		log.Println("Got SIGTERM signal")
		Events.Publish("boomer:quit")
		r.stop()
		r.waitTaskFinish()
	}()

	r.wait()

	log.Println("shut down")
}
