package glocust

import (
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"sync"
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
	waitFinishTime      = 5 * time.Second
)

// Task is like locust's task.
// when boomer receive start message, it will spawn several
// goroutines to run Task.Fn .

var runnerReady = false

type runner struct {
	newLocust   func() Locust
	numClients  int32
	hatchRate   int
	stopChannel chan bool
	exitChannel chan bool
	state       string
	client      client
	nodeID      string
	wg          sync.WaitGroup
}

func (r *runner) safeRun(fn func()) *bool {
	hasException := false
	defer func() {
		// don't panic
		err := recover()
		if err != nil {
			debug.PrintStack()
			RequestFailure("unknown", "panic", 0.0, fmt.Sprintf("%v", err))
			hasException = true
		}
	}()
	fn()
	return &hasException
}

func (r *runner) wait() {
	<-r.exitChannel
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

			go func() {
				r.wg.Add(1)
				defer func() {
					r.wg.Done()
					atomic.AddInt32(&r.numClients, -1)
				}()

				locust := r.newLocust()

				locust.OnStart()

				tasks := locust.WeightsTasks()
				// weightTasks := tasks
				for {
					select {
					case <-quit:
						return
					default:

						err := r.safeRun(tasks[rand.Intn(len(tasks))].Fn)
						if *err && (locust.CatchExceptions() == false) {
							return
						}
						sleepTime := rand.Intn(locust.Max()-locust.Min()) + locust.Min()
						if sleepTime != 0 {
							time.Sleep(time.Duration(sleepTime) * time.Millisecond)
						}
					}
				}
			}()
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
	if r.client != nil {
		toMaster <- newMessage("hatch_complete", data, r.nodeID)
	}
	r.state = stateRunning
}

func (r *runner) waitTaskFinish() {
	c := make(chan struct{})
	go func() {
		defer close(c)
		r.wg.Wait()
	}()

	select {
	case <-c:
		log.Println("all task finish") // completed normally
		break
	case <-time.After(waitFinishTime):
		log.Println("wait task finish timeout")
	}

	r.exitChannel <- true
}

func (r *runner) onQuiting() {
	if r.client != nil {
		toMaster <- newMessage("quit", nil, r.nodeID)
	}
}

func (r *runner) stop() {

	if r.state == stateRunning || r.state == stateHatching {
		close(r.stopChannel)
		r.state = stateStopped
		log.Println("Recv stop message, all the goroutines will stop")
	}

}

func (r *runner) getReady() {
	runnerReady = true
	r.state = stateInit

	if r.client == nil {
		return
	}
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
				r.stop()
				r.waitTaskFinish()
			}
		}
	}()

	// tell master, I'm ready
	toMaster <- newMessage("client_ready", nil, r.nodeID)

}

var grunner *runner

func newRunner(newLocust func() Locust, slaveMode bool) *runner {
	if slaveMode {
		client := newClient()
		grunner = &runner{
			newLocust:   newLocust,
			client:      client,
			nodeID:      getNodeID(),
			exitChannel: make(chan bool),
		}
		return grunner
	}

	grunner = &runner{
		newLocust:   newLocust,
		nodeID:      getNodeID(),
		exitChannel: make(chan bool),
	}
	return grunner
}

func reportStats(data map[string]interface{}) {
	if grunner != nil && grunner.client != nil {
		data["user_count"] = grunner.numClients
		toMaster <- newMessage("stats", data, grunner.nodeID)
	}
}
