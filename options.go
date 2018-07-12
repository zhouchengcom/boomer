package glocust

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var options struct {
	slave       *bool
	masterHost  *string
	masterPort  *int
	numClients  *int
	hatchRate   *int
	runTime     *int
	csvFilebase *string
	onlySummary *bool
}

func init() {
	options.csvFilebase = kingpin.Flag(
		"csv",
		"Store current request stats to files in CSV format.").
		String()

	options.slave = kingpin.Flag(
		"slave",
		"Set locust to run in distributed mode with this process as slave").
		Default("false").
		Bool()

	options.masterHost = kingpin.Flag(
		"master-host",
		"Host or IP address of locust master for distributed load testing. Only used when running with --slave. Defaults to 127.0.0.1.").
		Default("127.0.0.1").
		String()

	options.numClients = kingpin.Flag(
		"clients",
		"Number of concurrent Locust users. Only used together with --no-web").
		Default("1").
		Int()

	options.masterPort = kingpin.Flag(
		"master-port",
		"The port to connect to that is used by the locust master for distributed load testing. Only used when running with --slave. Defaults to 5557. Note that slaves will also connect to the master node on this port + 1.").
		Default("5557").
		Int()

	options.hatchRate = kingpin.Flag(
		"hatch-rate",
		"The rate per second in which clients are spawned. Only used together with --no-web").
		Default("1").
		Int()

	options.runTime = kingpin.Flag(
		"run-time",
		"Stop after the specified amount of time,. Only used together with --no-web").
		Default("0").
		Int()

	options.onlySummary = kingpin.Flag(
		"only-summary",
		"Only print the summary stats").
		Default("false").
		Bool()
}
