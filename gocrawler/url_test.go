package main

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
		wantErr  bool
	}{
		{
			name:     "remove scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
			wantErr:  false,
		},
		{
			name:     "remove before",
			inputURL: "https://blog.boot.dev",
			expected: "blog.boot.dev",
			wantErr:  false,
		},
		{
			name:     "remove after",
			inputURL: "blog.boot.dev/",
			expected: "blog.boot.dev",
			wantErr:  false,
		},
		{
			name:     "lowercase capital letters",
			inputURL: "https://BLOG.boot.dev/PATH",
			expected: "blog.boot.dev/path",
			wantErr:  false,
		},
		{
			name:     "handle invalid URL",
			inputURL: `:\\invalidURL`,
			expected: "",
			wantErr:  true,
		},
		{
			name:     "space scheme",
			inputURL: " ",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty scheme",
			inputURL: "",
			expected: "",
			wantErr:  true,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeURL(tc.inputURL)
			if err != nil && !tc.wantErr {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if actual != tc.expected {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
		wantErr   bool
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<a href="/path/one">
					<span>Boot.dev</span>
				</a>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
			wantErr:  false,
		},
		{
			name:     "absolute and relative URLs, different base",
			inputURL: "https://boot.dev",
			inputBody: `
		<html>
			<body>
				<a href="/path/one">
					<span>Boot.dev</span>
				</a>
				<a href="http://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://boot.dev/path/one", "http://other.com/path/one"},
			wantErr:  false,
		},
		{
			name:     "missing href",
			inputURL: "http://boot.dev",
			inputBody: `
		<html>
			<body>
				<a href="/path/relative">
					<span>Boot.dev</span>
				</a>
				<a href="http://other.com/path/one">
					<span>Boot.dev</span>
				</a>
				<a>
					<span>sometext</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"http://boot.dev/path/relative", "http://other.com/path/one"},
			wantErr:  false,
		},
		{
			name:     "no href",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a>
			<span>Boot.dev></span>
		</a>
	</body>
</html>
`,
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "invalid href URL",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href=":\\invalidURL">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{},
			wantErr:  false,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseURL, err := url.Parse(tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: couldn't parse input URL: %v", i, tc.name, err)
				return
			}
			actual, err := getURLsFromHTML(tc.inputBody, baseURL)
			if err != nil && !tc.wantErr {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
