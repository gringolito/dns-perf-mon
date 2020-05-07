package dnsperf

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/sevlyar/go-daemon"
)

var csvHeader []string = []string{"DATE", "DOMAIN", "LOOKUPTIME"}

type ctx struct {
	domainsFile    string
	resultsFile    string
	lookupInterval time.Duration
	context        *daemon.Context
	domains        []string
}

// DNSLookup type definition
type DNSLookup struct {
	requestTime time.Time
	domain      string
	ips         []net.IP
	elapsedTime time.Duration
}

// RunMonitor runs a daemonized dnsperf monitor
func RunMonitor(domainsFile string, resultsFile string, lookupInterval time.Duration) {
	context, err := daemonize()
	if err != nil {
		log.Panic("Aborting execution!")
	}
	defer context.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("dns-perf-mon daemon started")

	c := ctx{
		domainsFile:    domainsFile,
		resultsFile:    resultsFile,
		lookupInterval: lookupInterval,
		context:        context,
	}

	c.runDNSMonitor()
}

func daemonize() (*daemon.Context, error) {
	context := &daemon.Context{
		PidFileName: "dns-perf-mon.pid",
		PidFilePerm: 0644,
		LogFileName: "dns-perf-mon.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{},
	}

	child, err := context.Reborn()
	if err != nil {
		log.Print("Failed to daemonize: ", err)
		return context, err
	}
	if child != nil {
		os.Exit(0)
	}

	return context, nil
}

func (c *ctx) runDNSMonitor() {
	err := c.loadLookupDomains()
	if err != nil {
		log.Panic("Failed to load lookup domains from file: ", c.domainsFile)
	}

	for range time.Tick(c.lookupInterval) {
		c.performDNSLookup()
	}
}

func (c *ctx) loadLookupDomains() error {
	file, err := os.OpenFile(c.domainsFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("open file error: %v\n", err)
		return err
	}
	defer file.Close()

	c.domains = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c.domains = append(c.domains, scanner.Text())
	}

	return nil
}

func (c ctx) performDNSLookup() {
	domain := getRandomItem(c.domains)

	startTime := time.Now()

	ips, err := net.LookupIP(domain)
	if err != nil {
		log.Printf("[FAIL] Could not get IPs: %v\n", err)
		return
	}

	elapsedTime := time.Since(startTime)
	log.Printf("DNS lookup to '%s' took %s\n", domain, elapsedTime)

	results := DNSLookup{
		requestTime: startTime,
		domain:      domain,
		ips:         ips,
		elapsedTime: elapsedTime,
	}
	err = c.saveResults(results)
	if err != nil {
		log.Print("Failed to save result to file: ", c.resultsFile)
	}
}

func getRandomItem(itens []string) string {
	rand.Seed(time.Now().Unix())
	return itens[rand.Intn(len(itens))]
}

func (c ctx) saveResults(results DNSLookup) error {
	file, err := createOrOpenCSVFile(c.resultsFile)
	if err != nil {
		log.Printf("open file error: %v\n", err)
		return err
	}
	defer file.Close()

	milliseconds := fmt.Sprintf("%d", results.elapsedTime.Milliseconds())
	data := []string{results.requestTime.Format("2006-01-02 15:04:05"), results.domain, milliseconds}
	err = writeToCSV(file, data)
	if err != nil {
		log.Printf("write file error: %v\n", err)
		return err
	}

	return nil
}

func writeToCSV(file io.Writer, data []string) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	err := writer.Write(data)
	if err != nil {
		log.Printf("write file error: %v\n", err)
		return err
	}

	return nil
}

func createOrOpenCSVFile(filename string) (file *os.File, err error) {
	file = nil

	if _, err = os.Stat(filename); err == nil {
		return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	} else if os.IsNotExist(err) {
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return
		}

		err = writeToCSV(file, csvHeader)
		if err != nil {
			log.Print("Can't write CSV header on file: ", filename)
		}

	} else {
		// Schrodinger: file may or may not exist. See err for details.
	}

	return
}
