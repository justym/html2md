package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractContent(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		wantContains []string
		wantExcludes []string
	}{
		{
			name:         "empty string",
			html:         "",
			wantContains: []string{},
			wantExcludes: []string{},
		},
		{
			name:         "no body tag returns original",
			html:         "<h1>Title</h1><p>Content</p>",
			wantContains: []string{"Title", "Content"},
			wantExcludes: []string{},
		},
		{
			name: "extracts article content",
			html: `<html><body>
				<nav><a href="#">Menu</a></nav>
				<article><p>Main article content here.</p></article>
				<footer>Copyright 2024</footer>
			</body></html>`,
			wantContains: []string{"Main article content"},
			wantExcludes: []string{"Menu", "Copyright"},
		},
		{
			name: "extracts main content",
			html: `<html><body>
				<header><h1>Site Title</h1></header>
				<main><p>This is the main content.</p></main>
				<aside>Sidebar content</aside>
			</body></html>`,
			wantContains: []string{"main content"},
			wantExcludes: []string{"Site Title", "Sidebar"},
		},
		{
			name: "prefers content class over sidebar",
			html: `<html><body>
				<div class="sidebar"><p>Related links here</p></div>
				<div class="content"><p>Article text with many sentences.</p></div>
			</body></html>`,
			wantContains: []string{"Article text"},
			wantExcludes: []string{"Related links"},
		},
		{
			name: "removes script and style",
			html: `<html><body>
				<script>alert('hello')</script>
				<style>.foo { color: red; }</style>
				<article><p>Clean content here.</p></article>
			</body></html>`,
			wantContains: []string{"Clean content"},
			wantExcludes: []string{"alert", "color: red"},
		},
		{
			name: "removes hidden elements",
			html: `<html><body>
				<div hidden><p>Hidden content</p></div>
				<article><p>Visible content here.</p></article>
			</body></html>`,
			wantContains: []string{"Visible content"},
			wantExcludes: []string{"Hidden content"},
		},
		{
			name: "handles multiple paragraphs",
			html: `<html><body>
				<nav><a href="#">Link 1</a><a href="#">Link 2</a></nav>
				<div class="post">
					<p>First paragraph with some text.</p>
					<p>Second paragraph with more text.</p>
					<p>Third paragraph with even more text.</p>
				</div>
			</body></html>`,
			wantContains: []string{"First paragraph", "Second paragraph", "Third paragraph"},
			wantExcludes: []string{"Link 1"},
		},
		{
			name: "prefers high text density",
			html: `<html><body>
				<div class="nav"><a href="#">L1</a><a href="#">L2</a><a href="#">L3</a></div>
				<div class="article">
					<p>This is a long paragraph with lots of actual content text that should score higher because it has better text density and more actual words than links.</p>
				</div>
			</body></html>`,
			wantContains: []string{"long paragraph"},
			wantExcludes: []string{},
		},
		{
			name: "handles nested structure",
			html: `<html><body>
				<div id="wrapper">
					<header><nav>Navigation</nav></header>
					<main>
						<article>
							<h1>Article Title</h1>
							<p>Article body text.</p>
						</article>
					</main>
					<footer>Footer text</footer>
				</div>
			</body></html>`,
			wantContains: []string{"Article Title", "Article body"},
			wantExcludes: []string{"Footer text"},
		},
		{
			name: "fallback to body when no good candidate",
			html: `<html><body>
				<p>Just some text without containers.</p>
			</body></html>`,
			wantContains: []string{"Just some text"},
			wantExcludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractContent(tt.html)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("ExtractContent() should contain %q, got:\n%s", want, got)
				}
			}

			for _, exclude := range tt.wantExcludes {
				if strings.Contains(got, exclude) {
					t.Errorf("ExtractContent() should NOT contain %q, got:\n%s", exclude, got)
				}
			}
		})
	}
}

func TestScoreNode(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "article tag gets high score",
			html:    `<article><p>Content</p></article>`,
			wantMin: 20,
		},
		{
			name:    "nav tag gets low score",
			html:    `<nav><a href="#">Link</a></nav>`,
			wantMax: 0,
		},
		{
			name:    "content class gets bonus",
			html:    `<div class="content"><p>Text</p></div>`,
			wantMin: 25,
		},
		{
			name:    "sidebar class gets penalty",
			html:    `<div class="sidebar"><p>Text</p></div>`,
			wantMax: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseFirstElement(tt.html)
			if node == nil {
				t.Fatal("failed to parse HTML")
			}

			score := scoreNode(node)

			if tt.wantMin > 0 && score < tt.wantMin {
				t.Errorf("scoreNode() = %v, want >= %v", score, tt.wantMin)
			}
			if tt.wantMax > 0 && score > tt.wantMax {
				t.Errorf("scoreNode() = %v, want <= %v", score, tt.wantMax)
			}
		})
	}
}

func TestGetTextContent(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "simple text",
			html: `<p>Hello World</p>`,
			want: "Hello World",
		},
		{
			name: "nested elements",
			html: `<div><p>First</p><p>Second</p></div>`,
			want: "FirstSecond",
		},
		{
			name: "mixed content",
			html: `<p>Text <strong>bold</strong> more</p>`,
			want: "Text bold more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseFirstElement(tt.html)
			if node == nil {
				t.Fatal("failed to parse HTML")
			}

			got := getTextContent(node)
			if got != tt.want {
				t.Errorf("getTextContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemoveUnwantedElements(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		wantContains string
		wantExcludes string
	}{
		{
			name:         "removes script",
			html:         `<div><script>alert(1)</script><p>Keep</p></div>`,
			wantContains: "Keep",
			wantExcludes: "alert",
		},
		{
			name:         "removes style",
			html:         `<div><style>.foo{}</style><p>Keep</p></div>`,
			wantContains: "Keep",
			wantExcludes: ".foo",
		},
		{
			name:         "removes hidden",
			html:         `<div><div hidden>Hidden</div><p>Visible</p></div>`,
			wantContains: "Visible",
			wantExcludes: "Hidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseDocument(tt.html)
			if node == nil {
				t.Fatal("failed to parse HTML")
			}

			removeUnwantedElements(node)
			result := renderNode(node)

			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("result should contain %q, got:\n%s", tt.wantContains, result)
			}
			if strings.Contains(result, tt.wantExcludes) {
				t.Errorf("result should NOT contain %q, got:\n%s", tt.wantExcludes, result)
			}
		})
	}
}

// Helper functions for tests

func parseFirstElement(htmlStr string) *html.Node {
	doc := parseDocument(htmlStr)
	if doc == nil {
		return nil
	}
	return findFirstElement(doc)
}

func parseDocument(htmlStr string) *html.Node {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil
	}
	return doc
}

func findFirstElement(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data != "html" && n.Data != "head" && n.Data != "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c); found != nil {
			return found
		}
	}
	return nil
}
