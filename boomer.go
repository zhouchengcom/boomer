package glocust

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Run accepts a slice of Task and connects
// to a locust master.
func Run(tasks ...*Task) {

	if !flag.Parsed() {
		flag.Parse()
	}

	if *runTasks != "" {
		// Run tasks without connecting to the master.
		taskNames := strings.Split(*runTasks, ",")
		for _, task := range tasks {
			if task.Name == "" {
				continue
			} else {
				for _, name := range taskNames {
					if name == task.Name {
						log.Println("Running " + task.Name)
						task.Fn()
					}
				}
			}
		}
		return
	}

	if maxRPS > 0 {
		log.Println("Max RPS that boomer may generate is limited to", maxRPS)
		maxRPSEnabled = true
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

var runTasks *string
var maxRPS int64
var maxRPSThreshold int64
var maxRPSEnabled = false
var maxRPSControlChannel = make(chan bool)

var slave = false

var clients int
var hatchRate float64
var runTime *int

func init() {
	flag.Int64Var(&maxRPS, "max-rps", 0, "Max RPS that boomer can generate.")
	flag.IntVar(&clients, "clients", 1, "Number of concurrent Locust users. Only used together with --no-web")
	flag.Float64Var(&hatchRate, "hatch-rate", 1, "The rate per second in which clients are spawned. Only used together with --no-web")
	runTime = flag.Int("run-time", 0, "Stop after the specified amount of time, e.g. (300s, 20m, 3h, 1h30m, etc.). Only used together with --no-web")
}
