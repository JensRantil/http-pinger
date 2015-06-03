package main

import (
	"flag"
	carbon "github.com/marpaia/graphite-golang"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var httpUrl = flag.String("url", "http://localhost/", "The URL to issue the GET to.")
var carbonHost = flag.String("carbon-host", "localhost", "Carbon host/IP.")
var carbonPort = flag.Int("carbon-port", 2003, "Carbon port.")
var httpTimeout = flag.Duration("http-timeout", 30*time.Second, "HTTP request socket timeout.")
var testInterval = flag.Duration("http-interval", 10*time.Second, "HTTP test interval.")
var carbonInterval = flag.Duration("carbon-interval", 60*time.Second, "Interval to write to Carbon.")
var carbonNameSpace = flag.String("carbon-namespace", "http-pinger", "Where the Carbon data should be stored in Graphite.")

var quitChan = make(chan os.Signal, 1)

type testResult struct {
	latency time.Duration
	err     error
}

func runTest() *testResult {
	log.Println("Making test HTTP request...")

	client := http.Client{
		Timeout: *httpTimeout,
	}

	result := new(testResult)

	startTime := time.Now()
	resp, err := client.Get(*httpUrl)
	if resp != nil && resp.Body != nil {
		resp.Body.Close() // Not interested in the response.
	}
	result.latency = time.Now().Sub(startTime)
	result.err = err

	log.Println("Done making test HTTP request.")
	return result
}

type ByDuration []time.Duration

func (a ByDuration) Len() int           { return len(a) }
func (a ByDuration) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDuration) Less(i, j int) bool { return a[i] < a[j] }

func round(a float64) float64 {
	return math.Floor(a + 0.5)
}

func percentile(data []time.Duration, perc int) time.Duration {
	i := int(round(float64(len(data)-1) * float64(perc) / 100.0))
	return data[i]
}

func metricName(s string) string {
	return strings.Join([]string{*carbonNameSpace, s}, ".")
}

func milliseconds(t time.Duration) int64 {
	return t.Nanoseconds() / 1000000
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func graphiteSubmissionLoop(results <-chan *testResult, _carbon *carbon.Graphite) {
	timeouts := 0
	count := 0
	errors := 0
	latencies := make([]time.Duration, 0)

	minMetricName := metricName("min")
	maxMetricName := metricName("max")
	meanMetricName := metricName("median")
	lowerLowMetricName := metricName("5p")
	lowMetricName := metricName("25p")
	highMetricName := metricName("75p")
	upperHighMetricName := metricName("95p")
	timeoutsMetricName := metricName("timeouts")
	countMetricName := metricName("count")
	errorsMetricName := metricName("errors")

	testTicker := time.NewTicker(*carbonInterval)
	for {
		select {
		case <-testTicker.C:

			if len(latencies) == 0 {
				log.Println("No sample to send to Graphite.")
				continue
			}

			// Submit
			log.Println("Summarizing", len(latencies), "samples and sending to Carbon...")
			sort.Sort(ByDuration(latencies))

			now := time.Now().Unix()
			err := _carbon.SendMetrics([]carbon.Metric{
				carbon.NewMetric(minMetricName, itoa(milliseconds(latencies[0])), now),
				carbon.NewMetric(maxMetricName, itoa(milliseconds(percentile(latencies, 100))), now),
				carbon.NewMetric(meanMetricName, itoa(milliseconds(percentile(latencies, 50))), now),
				carbon.NewMetric(lowerLowMetricName, itoa(milliseconds(percentile(latencies, 5))), now),
				carbon.NewMetric(lowMetricName, itoa(milliseconds(percentile(latencies, 25))), now),
				carbon.NewMetric(upperHighMetricName, itoa(milliseconds(percentile(latencies, 95))), now),
				carbon.NewMetric(highMetricName, itoa(milliseconds(percentile(latencies, 75))), now),
				carbon.NewMetric(timeoutsMetricName, strconv.Itoa(timeouts), now),
				carbon.NewMetric(countMetricName, strconv.Itoa(count), now),
				carbon.NewMetric(errorsMetricName, strconv.Itoa(errors), now),
			})
			if err != nil {
				log.Println("Could not write to Graphite:", err)
			}
			latencies = latencies[:0]

		case result := <-results:

			count++
			if result.err == http.ErrHandlerTimeout {
				timeouts++
			}
			if result.err != nil {
				errors++
			}
			latencies = append(latencies, result.latency)

		case <-quitChan:

			log.Println("Stopping graphite submission...")
			break

		}
	}
}

func main() {
	flag.Parse()

	// Initialization

	signal.Notify(quitChan, syscall.SIGQUIT)

	*carbonNameSpace = strings.Trim(*carbonNameSpace, ".")

	_carbon, err := carbon.NewGraphite(*carbonHost, *carbonPort)
	if err != nil {
		log.Fatalln("Could not connect to Carbon:", err)
	}

	// Starting the actual testing

	log.Println("Starting...")

	results := make(chan *testResult)
	go graphiteSubmissionLoop(results, _carbon)

	testTicker := time.NewTicker(*testInterval)
	for {
		select {
		case <-testTicker.C:
			results <- runTest()
		case <-quitChan:
			log.Println("Stopping test loop...")
			break
		}
	}
}
