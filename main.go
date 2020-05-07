package main

import (
	"flag"
	"time"

	"github.com/gringolito/dns-perf-mon/dnsperf"
)

func main() {

	var domainsFile string
	flag.StringVar(&domainsFile, "domains", "domains.txt", "Domains list input file")

	var resultsFile string
	flag.StringVar(&resultsFile, "output", "dns-lookup-times.csv", "Output CSV file")

	var lookupInterval time.Duration
	flag.DurationVar(&lookupInterval, "interval", time.Minute, "Interval between lookups")

	flag.Parse()

	dnsperf.RunMonitor(domainsFile, resultsFile, lookupInterval)
}
