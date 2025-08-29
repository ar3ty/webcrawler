package main

import (
	"fmt"
	"net/url"
	"sort"
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

func (cfg *config) addPageVisit(normalizedUrl string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, ok := cfg.pages[normalizedUrl]; ok {
		cfg.pages[normalizedUrl]++
		return false
	}
	cfg.pages[normalizedUrl] = 1
	return true
}

func (cfg *config) getLengthPages() int {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return len(cfg.pages)
}

func configure(parsedBase *url.URL, maxConcurrency, maxPages int) *config {
	return &config{
		pages:              map[string]int{},
		baseURL:            parsedBase,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}
}

type Record struct {
	URL   string
	Count int
}

type ByCountAlphabetically []Record

func (a ByCountAlphabetically) Len() int      { return len(a) }
func (a ByCountAlphabetically) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCountAlphabetically) Less(i, j int) bool {
	if a[i].Count == a[j].Count {
		return a[i].URL < a[j].URL
	}
	return a[i].Count > a[j].Count
}

func (cfg *config) printReport() {
	fmt.Println("=============================")
	fmt.Printf("REPORT for %s\n", cfg.baseURL.String())
	fmt.Println("=============================")

	records := []Record{}
	for k, v := range cfg.pages {
		parsedKey, err := url.Parse(k)
		if err != nil {
			fmt.Printf("Error parsing key: %s\n", k)
		}
		records = append(records, Record{
			URL:   cfg.baseURL.ResolveReference(parsedKey).String(),
			Count: v,
		})
	}

	sort.Sort(ByCountAlphabetically(records))

	for _, record := range records {
		fmt.Printf("Found %d internal links to %s\n", record.Count, record.URL)
	}
}
