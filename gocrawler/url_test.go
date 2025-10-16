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

func TestGetH1FromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputBody string
		expected  string
	}{
		{
			name: "header",
			inputBody: `
		<html>
  			<body>
    			<h1>Welcome to Boot.dev</h1>
    			<main>
      				<p>Learn to code by building real projects.</p>
      				<p>This is the second paragraph.</p>
    			</main>
  			</body>
		</html>
		`,
			expected: "Welcome to Boot.dev",
		},
		{
			name:      "header again",
			inputBody: "<html><body><h1>Test Title</h1></body></html>",
			expected:  "Test Title",
		},
		{
			name:      "missing header",
			inputBody: `<html><body></body></html>`,
			expected:  "",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := getH1FromHTML(tc.inputBody)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}

func TestGetFirstParagraphFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputBody string
		expected  string
	}{
		{
			name: "paragraph",
			inputBody: `<html><body>
		<p>Outside paragraph.</p>
		<main>
			<p>Main paragraph.</p>
		</main>
	</body></html>`,
			expected: "Main paragraph.",
		},
		{
			name:      "no paragraph",
			inputBody: "<html><body><h1>No paragraphs here</h1></body></html>",
			expected:  "",
		},
		{
			name: "two paragraphs",
			inputBody: `<html><body>
		<p>First paragraph outside main.</p>
		<p>Second paragraph outside main.</p>
	</body></html>`,
			expected: "First paragraph outside main.",
		},
		{
			name:      "paragraph again",
			inputBody: "<html><body><p>This is the first paragraph.</p></body></html>",
			expected:  "This is the first paragraph.",
		},
		{
			name:      "paragraph empty",
			inputBody: "<html><body><p></p></body></html>",
			expected:  "",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := getFirstParagraphFromHTML(tc.inputBody)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}

func TestGetURLsFromHTML(t *testing.T) {
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

func TestGetImagesFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
		wantErr   bool
	}{
		{
			name:      "absolute",
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src="https://blog.boot.dev/logo.png" alt="Logo"></body></html>`,
			expected:  []string{"https://blog.boot.dev/logo.png"},
			wantErr:   false,
		},
		{
			name:      "multiple",
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src="/logo.png" alt="Logo"><img src="https://cdn.boot.dev/banner.jpg"></body></html>`,
			expected:  []string{"https://blog.boot.dev/logo.png", "https://cdn.boot.dev/banner.jpg"},
			wantErr:   false,
		},
		{
			name:      "relative",
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src="/logo.png" alt="Logo"></body></html>`,
			expected:  []string{"https://blog.boot.dev/logo.png"},
			wantErr:   false,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseURL, err := url.Parse(tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: couldn't parse input URL: %v", i, tc.name, err)
				return
			}
			actual, err := getImagesFromHTML(tc.inputBody, baseURL)
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

func TestExtractPageData(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  PageData
		wantErr   bool
	}{
		{
			name:     "basic",
			inputURL: "https://blog.boot.dev",
			inputBody: `<html><body>
        <h1>Test Title</h1>
        <p>This is the first paragraph.</p>
        <a href="/link1">Link 1</a>
        <img src="/image1.jpg" alt="Image 1">
    </body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "Test Title",
				FirstParagraph: "This is the first paragraph.",
				OutgoingLinks:  []string{"https://blog.boot.dev/link1"},
				ImageURLs:      []string{"https://blog.boot.dev/image1.jpg"},
			},
			wantErr: false,
		},
		{
			name:     "only href",
			inputURL: "https://blog.boot.dev",
			inputBody: `<html><body>
            <h1>Test Title</h1>
            <p>This is the first paragraph.</p>
            <a href="/link1">Link 1</a>
        </body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "Test Title",
				FirstParagraph: "This is the first paragraph.",
				OutgoingLinks:  []string{"https://blog.boot.dev/link1"},
				ImageURLs:      []string{},
			},
			wantErr: false,
		},
		{
			name:      "missing_elements",
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><div>No h1, p, links, or images</div></body></html>`,
			expected: PageData{
				URL:            "https://blog.boot.dev",
				H1:             "",
				FirstParagraph: "",
				OutgoingLinks:  []string{},
				ImageURLs:      []string{},
			},
			wantErr: false,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := extractPageData(tc.inputBody, tc.inputURL)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
