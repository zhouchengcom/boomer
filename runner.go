package glocust

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const (
	stateInit     = "ready"
	stateHatching = "hatching"
	stateRunning  = "running"
	stateStopped  = "stopped"
)

const (
	slaveReportInterval = 3 * time.Second
)

// Task is like locust's task.
// when boomer receive start message, it will spawn several
// goroutines to run Task.Fn .

type runner struct {
	newLocust   func() Locust
	numClients  int32
	hatchRate   int
	stopChannel chan bool
	state       string
	client      client
	nodeID      string
}

func (r *runner) safeRun(fn func()) *bool {
	hasException := false
	defer func() {
		// don't panic
		err := recover()
		if err != nil {
			debug.PrintStack()
			Events.Publish("request_failure", "unknown", "panic", 0.0, fmt.Sprintf("%v", err))
			hasException = true
		}
	}()
	fn()
	return &hasException
}

func (r *runner) spawnGoRoutines(spawnCount int, quit chan bool) {

	log.Println("Hatching and swarming", spawnCount, "clients at the rate", r.hatchRate, "clients/s...")

	for i := 1; i <= spawnCount; i++ {
		select {
		case <-quit:
			// quit hatching goroutine
			return
		default:
			if i%r.hatchRate == 0 {
				time.Sleep(1 * time.Second)
			}
			atomic.AddInt32(&r.numClients, 1)
			locust := r.newLocust()
			go func(locust Locust) {
				locust.OnStart()

				tasks := locust.Tasks()
				weightTasks := tasks
				for {
					select {
					case <-quit:
						return
					default:
						err := r.safeRun(tasks[0].Fn)
						if *err && (locust.CatchExceptions() == false) {
							return
						}
					}
				}
			}(locust)
		}
	}
	r.hatchComplete()
}

func (r *runner) startHatching(spawnCount int, hatchRate int) {

	if r.state != stateRunning && r.state != stateHatching {
		clearStatsChannel <- true
		r.stopChannel = make(chan bool)
	}

	if r.state == stateRunning || r.state == stateHatching {
		// stop previous goroutines without blocking
		// those goroutines will exit when r.safeRun returns
		close(r.stopChannel)
	}

	r.stopChannel = make(chan bool)
	r.state = stateHatching

	r.hatchRate = hatchRate
	r.numClients = 0
	go r.spawnGoRoutines(spawnCount, r.stopChannel)
}

func (r *runner) hatchComplete() {

	data := make(map[string]interface{})
	data["count"] = r.numClients
	toMaster <- newMessage("hatch_complete", data, r.nodeID)

	r.state = stateRunning
}

func (r *runner) onQuiting() {
	toMaster <- newMessage("quit", nil, r.nodeID)
}

func (r *runner) stop() {

	if r.state == stateRunning || r.state == stateHatching {
		close(r.stopChannel)
		r.state = stateStopped
		log.Println("Recv stop message from master, all the goroutines are stopped")
	}

}

func (r *runner) getReady() {

	r.state = stateInit

	// read message from master
	go func() {
		for {
			msg := <-fromMaster
			switch msg.Type {
			case "hatch":
				toMaster <- newMessage("hatching", nil, r.nodeID)
				rate, _ := msg.Data["hatch_rate"]
				clients, _ := msg.Data["num_clients"]
				hatchRate := int(rate.(float64))
				workers := 0
				if _, ok := clients.(uint64); ok {
					workers = int(clients.(uint64))
				} else {
					workers = int(clients.(int64))
				}
				if workers == 0 || hatchRate == 0 {
					log.Printf("Invalid hatch message from master, num_clients is %d, hatch_rate is %d\n",
						workers, hatchRate)
				} else {
					r.startHatching(workers, hatchRate)
				}
			case "stop":
				r.stop()
				toMaster <- newMessage("client_stopped", nil, r.nodeID)
				toMaster <- newMessage("client_ready", nil, r.nodeID)
			case "quit":
				log.Println("Got quit message from master, shutting down...")
				os.Exit(0)
			}
		}
	}()

	// tell master, I'm ready
	toMaster <- newMessage("client_ready", nil, r.nodeID)

	// report to master
	go func() {
		for {
			select {
			case data := <-messageToRunner:
				data["user_count"] = r.numClients
				toMaster <- newMessage("stats", data, r.nodeID)
			}
		}
	}()

}
