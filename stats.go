package glocust

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

type requestStats struct {
	entries   map[string]*statsEntry
	errors    map[string]*statsError
	total     *statsEntry
	startTime int64
	csvFile   *os.File
	csvWriter *csv.Writer
}

func newRequestStats() *requestStats {
	entries := make(map[string]*statsEntry)
	errors := make(map[string]*statsError)

	requestStats := &requestStats{
		entries: entries,
		errors:  errors,
	}

	requestStats.total = &statsEntry{
		name:   "Total",
		method: "",
	}
	requestStats.total.reset()

	return requestStats
}

func (s *requestStats) logRequest(method, name string, responseTime int64, contentLength int64) {
	s.total.log(responseTime, contentLength)
	s.get(name, method).log(responseTime, contentLength)
	if s.csvWriter != nil {
		s.csvWriter.Write([]string{
			strconv.FormatInt(time.Now().Unix(), 10),
			"request",
			"success", method, name, "1",
			strconv.FormatInt(responseTime, 10),
			strconv.FormatInt(contentLength, 10)})
	}
}

func (s *requestStats) logError(method, name, err string) {
	s.total.logError(err)
	s.get(name, method).logError(err)

	if s.csvWriter != nil {
		s.csvWriter.Write([]string{
			strconv.FormatInt(time.Now().Unix(), 10),
			"request",
			"failure", method, name, "1",
			"0", err, "-"})
	}
	// store error in errors map
	key := MD5(method, name, err)
	entry, ok := s.errors[key]
	if !ok {
		entry = &statsError{
			name:   name,
			method: method,
			error:  err,
		}
		s.errors[key] = entry
	}
	entry.occured()
}

func (s *requestStats) get(name string, method string) (entry *statsEntry) {
	entry, ok := s.entries[name+method]
	if !ok {
		newEntry := &statsEntry{
			name:          name,
			method:        method,
			numReqsPerSec: make(map[int64]int64),
			responseTimes: make(map[int64]int64),
		}
		newEntry.reset()
		s.entries[name+method] = newEntry
		return newEntry
	}
	return entry
}

func (s *requestStats) clearAll() {
	s.total = &statsEntry{
		name:   "Total",
		method: "",
	}
	s.total.reset()

	s.entries = make(map[string]*statsEntry)
	s.errors = make(map[string]*statsError)
	s.startTime = time.Now().Unix()
}

func (s *requestStats) serializeStats() []interface{} {
	entries := make([]interface{}, 0, len(s.entries))
	for _, v := range s.entries {
		if !(v.numRequests == 0 && v.numFailures == 0) {
			entries = append(entries, v.getStrippedReport())
		}
	}
	return entries
}

func (s *requestStats) serializeErrors() map[string]map[string]interface{} {
	errors := make(map[string]map[string]interface{})
	for k, v := range s.errors {
		errors[k] = v.toMap()
	}
	return errors
}

func (s *requestStats) createResultFile(name *string) error {
	f, err := os.OpenFile(*name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("create result file %s fail\n", name)
		return err
	}
	s.csvFile = f
	s.csvWriter = csv.NewWriter(f)
	s.csvWriter.Write([]string{"time", "type", "result", "subtype",
		"name", "count", "respondtime", "exception", "parent"})

	return nil
}

func (s *requestStats) closeResultFile() {
	if s.csvWriter != nil {
		s.csvWriter.Flush()
	}
	if s.csvFile != nil {
		s.csvFile.Close()
	}
}

type statsEntry struct {
	name                 string
	method               string
	numRequests          int64
	numFailures          int64
	totalResponseTime    int64
	minResponseTime      int64
	maxResponseTime      int64
	numReqsPerSec        map[int64]int64
	responseTimes        map[int64]int64
	totalContentLength   int64
	startTime            int64
	lastRequestTimestamp int64
}

func (s *statsEntry) reset() {
	s.startTime = time.Now().Unix()
	s.numRequests = 0
	s.numFailures = 0
	s.totalResponseTime = 0
	s.responseTimes = make(map[int64]int64)
	s.minResponseTime = 0
	s.maxResponseTime = 0
	s.lastRequestTimestamp = time.Now().Unix()
	s.numReqsPerSec = make(map[int64]int64)
	s.totalContentLength = 0
}

func (s *statsEntry) log(responseTime int64, contentLength int64) {
	s.numRequests++

	s.logTimeOfRequest()
	s.logResponseTime(responseTime)

	s.totalContentLength += contentLength
}

func (s *statsEntry) logTimeOfRequest() {
	now := time.Now().Unix()

	_, ok := s.numReqsPerSec[now]
	if !ok {
		s.numReqsPerSec[now] = 1
	} else {
		s.numReqsPerSec[now]++
	}

	s.lastRequestTimestamp = now
}

func (s *statsEntry) logResponseTime(responseTime int64) {
	s.totalResponseTime += responseTime

	if s.minResponseTime == 0 {
		s.minResponseTime = responseTime
	}

	if responseTime < s.minResponseTime {
		s.minResponseTime = responseTime
	}

	if responseTime > s.maxResponseTime {
		s.maxResponseTime = responseTime
	}

	roundedResponseTime := int64(0)

	// to avoid to much data that has to be transferred to the master node when
	// running in distributed mode, we save the response time rounded in a dict
	// so that 147 becomes 150, 3432 becomes 3400 and 58760 becomes 59000
	// see also locust's stats.py
	if responseTime < 100 {
		roundedResponseTime = responseTime
	} else if responseTime < 1000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -1))
	} else if responseTime < 10000 {
		roundedResponseTime = int64(round(float64(responseTime), .5, -2))
	} else {
		roundedResponseTime = int64(round(float64(responseTime), .5, -3))
	}

	_, ok := s.responseTimes[roundedResponseTime]
	if !ok {
		s.responseTimes[roundedResponseTime] = 1
	} else {
		s.responseTimes[roundedResponseTime]++
	}
}

func (s *statsEntry) logError(err string) {
	s.numFailures++
}

func (s *statsEntry) serialize() map[string]interface{} {
	result := make(map[string]interface{})
	result["name"] = s.name
	result["method"] = s.method
	result["last_request_timestamp"] = s.lastRequestTimestamp
	result["start_time"] = s.startTime
	result["num_requests"] = s.numRequests
	result["num_failures"] = s.numFailures
	result["total_response_time"] = s.totalResponseTime
	result["max_response_time"] = s.maxResponseTime
	result["min_response_time"] = s.minResponseTime
	result["total_content_length"] = s.totalContentLength
	result["response_times"] = s.responseTimes
	result["num_reqs_per_sec"] = s.numReqsPerSec
	return result
}

func (s *statsEntry) getStrippedReport() map[string]interface{} {
	report := s.serialize()
	s.reset()
	return report
}

var buffer bytes.Buffer

func printStats(stats *requestStats) {
	buffer.Reset()
	fmt.Fprintf(&buffer, "%15s %7s %7s %7s %7s %7s  | %7s %7s\n", "Name", "# reqs", "# fails", "Avg", "Min", "Max", "Median", "req/s")
	fmt.Fprintf(&buffer, "----------------------------------------------------------------------------------------------------------\n")

	var keys []string
	for k := range stats.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	now := time.Now().Unix()
	for _, k := range keys {
		s := stats.entries[k]
		qps := int64(0)
		if now-s.lastRequestTimestamp < 3 {
			qps = s.numReqsPerSec[s.lastRequestTimestamp]
		}

		sum := int64(0)
		for _, v := range s.responseTimes {
			sum += v
		}
		avg := int64(0)
		if len(s.responseTimes) > 0 {
			avg = sum / int64(len(s.responseTimes))
		}

		fmt.Fprintf(&buffer, "%15s %7d %7d %7d %7d %7d  | %7d %7d\n", s.name, s.numRequests,
			s.numFailures, avg, s.minResponseTime, s.maxResponseTime, 0, qps)

	}
	fmt.Fprintf(&buffer, "----------------------------------------------------------------------------------------------------------\n")
	total := stats.total
	tqps := int64(0)
	if now-total.lastRequestTimestamp < 3 {
		tqps = total.numReqsPerSec[total.lastRequestTimestamp]
	}
	fmt.Fprintf(&buffer, "%15s %7d %7d %7d %7d %7d  | %7d %7d\n\n", total.name, total.numRequests,
		total.numFailures, 0, total.minResponseTime, total.maxResponseTime, 0, tqps)
	print(buffer.String())

}

type statsError struct {
	name       string
	method     string
	error      string
	occurences int64
}

func (err *statsError) occured() {
	err.occurences++
}

func (err *statsError) toMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["method"] = err.method
	m["name"] = err.name
	m["error"] = err.error
	m["occurences"] = err.occurences

	return m
}

func collectReportData() map[string]interface{} {
	data := make(map[string]interface{})

	data["stats"] = stats.serializeStats()
	data["stats_total"] = stats.total.getStrippedReport()
	data["errors"] = stats.serializeErrors()

	stats.errors = make(map[string]*statsError)

	return data
}

type requestSuccess struct {
	requestType    string
	name           string
	responseTime   int64
	responseLength int64
}

type requestFailure struct {
	requestType  string
	name         string
	responseTime int64
	error        string
}

var stats = newRequestStats()
var requestSuccessChannel = make(chan *requestSuccess, 1000)
var requestFailureChannel = make(chan *requestFailure, 1000)
var clearStatsChannel = make(chan bool)

func init() {
	stats.entries = make(map[string]*statsEntry)
	stats.errors = make(map[string]*statsError)
	go func() {
		var ticker = time.NewTicker(slaveReportInterval)
		for {
			select {
			case m := <-requestSuccessChannel:
				// println("dd")
				stats.logRequest(m.requestType, m.name, m.responseTime, m.responseLength)
			case n := <-requestFailureChannel:
				stats.logError(n.requestType, n.name, n.error)
			case <-clearStatsChannel:
				stats.clearAll()
			case <-ticker.C:
				if *options.slave != true && *options.onlySummary != true && runnerReady == true {
					printStats(stats)
				}
				data := collectReportData()

				reportStats(data)
			}
		}
	}()
}
