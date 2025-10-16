package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
)

func main() {
	args := os.Args
	if len(args) != 4 {
		log.Fatal("usage: ./crawler URL maxConcurrency maxPages")
	}

	baseURL := args[1]
	maxConcurrency, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal("invalid type of maxConcurrency arg")
	}
	maxPages, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatal("invalid type of maxPages arg")
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal("invalid url provided")
	}

	cfg := configure(parsedBase, maxConcurrency, maxPages)

	fmt.Printf("Starting crawl of: %s...\n", baseURL)

	cfg.wg.Add(1)
	go cfg.crawlPage(baseURL)

	cfg.wg.Wait()
	err = cfg.writeCSVReport("report.csv")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("Report written to %s\n", "report.csv")
}
