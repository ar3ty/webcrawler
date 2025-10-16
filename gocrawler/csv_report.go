package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
)

func (cfg *config) writeCSVReport(filename string) error {
	if len(cfg.pages) == 0 {
		return fmt.Errorf("no data to write to CSV")
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	err = writer.Write([]string{"page_url", "h1", "first_paragraph", "outgoing_link_urls", "image_urls"})
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(cfg.pages))
	for k := range cfg.pages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		current := cfg.pages[key]
		writer.Write([]string{current.URL,
			current.H1,
			current.FirstParagraph,
			strings.Join(current.OutgoingLinks, ";"),
			strings.Join(current.ImageURLs, ";")})
	}
	return nil
}
