package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL            string
	H1             string
	FirstParagraph string
	OutgoingLinks  []string
	ImageURLs      []string
}

func getH1FromHTML(html string) string {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return ""
	}
	headers := doc.Find("h1").First().Text()
	return strings.TrimSpace(headers)
}

func getFirstParagraphFromHTML(html string) string {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return ""
	}
	main := doc.Find("main")
	var p string
	if main.Length() > 0 {
		p = main.Find("p").First().Text()
	} else {
		p = doc.Find("p").First().Text()
	}

	return strings.TrimSpace(p)
}

func normalizeURL(toParse string) (string, error) {
	if strings.Fields(toParse) == nil || toParse == " " {
		return "", errors.New("empty url")
	}
	parsedURL, err := url.Parse(toParse)
	if err != nil {
		return "", fmt.Errorf("failure parsing url: %w", err)
	}

	normalized := parsedURL.Host + parsedURL.Path
	normalized = strings.TrimSuffix(normalized, "/")
	normalized = strings.ToLower(normalized)

	return normalized, nil
}

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	reader := strings.NewReader(htmlBody)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	urls := []string{}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if !ok || strings.TrimSpace(link) == "" {
			return
		}
		parsedLink, err := url.Parse(link)
		if err != nil {
			return
		}
		resolved := baseURL.ResolveReference(parsedLink)
		urls = append(urls, resolved.String())
	})
	return urls, nil
}

func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	reader := strings.NewReader(htmlBody)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	urls := []string{}
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		link, ok := s.Attr("src")
		if !ok || strings.TrimSpace(link) == "" {
			return
		}
		parsedLink, err := url.Parse(link)
		if err != nil {
			return
		}
		resolved := baseURL.ResolveReference(parsedLink)
		urls = append(urls, resolved.String())
	})
	return urls, nil
}

func extractPageData(html, pageURL string) PageData {
	var pd PageData
	parsedBase, err := url.Parse(pageURL)
	if err != nil {
		return pd
	}
	links, err := getURLsFromHTML(html, parsedBase)
	if err != nil {
		return pd
	}
	images, err := getImagesFromHTML(html, parsedBase)
	if err != nil {
		return pd
	}
	return PageData{
		URL:            pageURL,
		H1:             getH1FromHTML(html),
		FirstParagraph: getFirstParagraphFromHTML(html),
		OutgoingLinks:  links,
		ImageURLs:      images,
	}
}

func getHTML(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("couldn't make request: %w", err)
	}

	req.Header.Set("User-Agent", "crawler")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("couldn't get response: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("response failed with code %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return "", fmt.Errorf("invalid content type")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read body: %w", err)
	}

	return string(body), nil
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()
	cfg.concurrencyControl <- struct{}{}

	if cfg.getLengthPages() >= cfg.maxPages {
		return
	}

	parsedCurrent, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error crawlPage: couldn't parse URL '%s': %v\n", rawCurrentURL, err)
		return
	}
	if cfg.baseURL.Hostname() != parsedCurrent.Hostname() {
		return
	}

	normalisedCurrent, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - normalizeURL: %v\n", err)
		return
	}

	isFirst := cfg.addPageVisit(normalisedCurrent)
	if !isFirst {
		return
	}

	fmt.Printf("Crawling %s...\n", rawCurrentURL)
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - getHTML: %v\n", err)
		return
	}

	pageData := extractPageData(html, rawCurrentURL)
	cfg.setPageData(normalisedCurrent, pageData)

	for _, link := range pageData.OutgoingLinks {
		cfg.wg.Add(1)
		go cfg.crawlPage(link)
	}
}
