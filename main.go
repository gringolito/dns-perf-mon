package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/sevlyar/go-daemon"
)

var domainsFile string = "domains.txt"
var resultsFile string = "dns-query-times.csv"

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
		log.Fatal("Failed to daemonize: ", err)
		return context, err
	}
	if child != nil {
		os.Exit(0)
	}

	return context, nil
}

func main() {
	context, err := daemonize()
	if err != nil {
		log.Fatal("Aborting execution!")
		os.Exit(1)
	}
	defer context.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("dns-perf-mon daemon started")

	runDNSMonitor()
}

func runDNSMonitor() {
	domains, err := getLookupDomains(domainsFile)
	if err != nil {
		log.Fatalf("Failed to load lookup domains from file: %s", domainsFile)
		os.Exit(1)
	}

	for range time.Tick(time.Minute) {
		performDNSLookup(domains)
	}
}

func getLookupDomains(inputFile string) ([]string, error) {
	file, err := os.OpenFile(inputFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return nil, err
	}
	defer file.Close()

	domains := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domains = append(domains, scanner.Text())
	}

	return domains, nil
}

func performDNSLookup(domains []string) {
	domain := getRandomItem(domains)

	startTime := time.Now()

	_, err := net.LookupIP(domain)
	if err != nil {
		log.Fatalf("Could not get IPs: %v\n", err)
		return
	}

	elapsedTime := time.Since(startTime)
	log.Printf("DNS lookup to '%s' took %s\n", domain, elapsedTime)

	err = saveResults(resultsFile, startTime, domain, elapsedTime)
	if err != nil {
		log.Fatalf("Failed to save result to file: %s", resultsFile)
	}
}

func getRandomItem(itens []string) string {
	rand.Seed(time.Now().Unix())
	return itens[rand.Intn(len(itens))]
}

func saveResults(outputFile string, date time.Time, domain string, elapsedTime time.Duration) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	milliseconds := fmt.Sprintf("%d", elapsedTime.Milliseconds())
	results := []string{date.Format("2006-01-02 15:04:05"), domain, milliseconds}
	err = writer.Write(results)
	if err != nil {
		log.Fatalf("write file error: %v", err)
		return err
	}

	return nil
}
