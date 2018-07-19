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

	if len(*options.csvFilebase) > 0 {
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

	r.getReady()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	log.Println("press Ctrl+c to quit")

	if *options.runTime != 0 {
		go func() {
			time.Sleep(time.Duration(*options.runTime) * time.Second)
			log.Println("Time limit reached. Stopping Locust.")
			r.stop()
			r.waitTaskFinish()
		}()
	}
	go func() {
		<-c
		log.Println("Got SIGTERM signal")
		r.stop()
		r.waitTaskFinish()
	}()

	r.wait()

	r.onQuiting()
	// wait for quit message is sent to master
	<-disconnectedFromMaster
	log.Println("shut down")
}

func runLocal(newLocust func() Locust) {

	r := newRunner(newLocust, false)

	r.getReady()
	r.startHatching(*options.numClients, *options.hatchRate)
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)

	if *options.runTime != 0 {
		go func() {
			time.Sleep(time.Duration(*options.runTime) * time.Second)
			log.Println("Time limit reached. Stopping Locust.")
			r.stop()
			r.waitTaskFinish()
		}()
	}
	go func() {
		<-c
		log.Println("Got SIGTERM signal")
		r.stop()
		r.waitTaskFinish()
	}()

	r.wait()
	log.Println("shut down")
}
