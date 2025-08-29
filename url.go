package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failure parsing baseurl: %w", err)
	}

	reader := strings.NewReader(htmlBody)
	tree, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failure parsing htmlbody: %w", err)
	}
	urls := []string{}

	var treeTraverse func(*html.Node)
	treeTraverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			for _, a := range node.Attr {
				if a.Key == "href" {
					parsed, err := url.Parse(a.Val)
					if err != nil {
						fmt.Printf("failure parsing url: %v\n", err)
						continue
					}
					resolved := base.ResolveReference(parsed)
					if resolved.Scheme == "http" || resolved.Scheme == "https" {
						urls = append(urls, resolved.String())
					}
				}
			}
		}

		for n := node.FirstChild; n != nil; n = n.NextSibling {
			treeTraverse(n)
		}
	}
	treeTraverse(tree)

	return urls, nil
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

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	parsedBase, err := url.Parse(rawBaseURL)
	if err != nil {
		fmt.Printf("Error crawlPage: couldn't parse URL '%s': %v\n", rawBaseURL, err)
		return
	}
	parsedCurrent, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error crawlPage: couldn't parse URL '%s': %v\n", rawCurrentURL, err)
		return
	}
	if parsedBase.Hostname() != parsedCurrent.Hostname() {
		return
	}

	normalisedCurrent, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - normalizeURL: %v\n", err)
		return
	}

	if _, ok := pages[normalisedCurrent]; ok {
		pages[normalisedCurrent]++
		return
	}
	pages[normalisedCurrent] = 1

	fmt.Printf("Crawling %s...\n", rawCurrentURL)
	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error - getHTML: %v\n", err)
		return
	}
	urls, err := getURLsFromHTML(html, rawBaseURL)
	if err != nil {
		fmt.Printf("Error - getURLsFromHTML: %v\n", err)
		return
	}

	for _, link := range urls {
		crawlPage(rawBaseURL, link, pages)
	}
}
