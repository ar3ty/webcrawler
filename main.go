package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

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

	cfg := config{
		pages:              map[string]int{},
		baseURL:            parsedBase,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	fmt.Printf("Starting crawl of: %s...\n", baseURL)

	cfg.wg.Add(1)
	go cfg.crawlPage(baseURL)

	cfg.wg.Wait()
	for k, v := range cfg.pages {
		fmt.Printf("Page %s, count: %d\n", k, v)
	}
}
