package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("no website provided")
	}
	if len(args) > 2 {
		log.Fatal("too many arguments provided")
	}

	baseURL := args[1]
	fmt.Printf("Starting crawl of: %s...\n", baseURL)

	pages := map[string]int{}
	crawlPage(baseURL, baseURL, pages)
	for k, v := range pages {
		fmt.Printf("Page %s, count: %d\n", k, v)
	}
}
