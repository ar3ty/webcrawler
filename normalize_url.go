package main

import (
	"errors"
	"fmt"
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
