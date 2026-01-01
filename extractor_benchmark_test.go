package main

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// Test data of varying sizes
var (
	smallHTML  = `<html><body><article><p>Small content here.</p></article></body></html>`
	mediumHTML = generateMediumHTML()
	largeHTML  = generateLargeHTML()
)

// generateMediumHTML creates a ~1KB HTML document.
func generateMediumHTML() string {
	var sb strings.Builder
	sb.WriteString(`<html><body>
		<nav><a href="#">Home</a><a href="#">About</a><a href="#">Contact</a></nav>
		<article class="content">
			<h1>Article Title</h1>`)
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf(`
			<p>This is paragraph %d with some content. It contains multiple sentences to simulate real article text, including commas, periods, and other punctuation marks.</p>`, i))
	}
	sb.WriteString(`
		</article>
		<aside class="sidebar"><p>Sidebar content</p></aside>
		<footer>Copyright 2026</footer>
	</body></html>`)
	return sb.String()
}

// generateLargeHTML creates a ~10KB HTML document.
func generateLargeHTML() string {
	var sb strings.Builder
	sb.WriteString(`<html><body>
		<header>
			<nav>`)
	for i := 0; i < 20; i++ {
		sb.WriteString(fmt.Sprintf(`<a href="/page%d">Link %d</a>`, i, i))
	}
	sb.WriteString(`</nav>
		</header>
		<main>
			<article class="post">
				<h1>Main Article Title</h1>`)
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf(`
				<p>This is paragraph number %d in the article. It contains substantial text content with various punctuation marks, including commas, semicolons; and other characters. The goal is to simulate a real-world article with meaningful content density.</p>`, i))
	}
	sb.WriteString(`
			</article>
		</main>
		<aside class="sidebar">`)
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf(`<div class="widget"><h3>Widget %d</h3><p>Widget content</p></div>`, i))
	}
	sb.WriteString(`</aside>
		<footer>
			<nav>`)
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf(`<a href="/footer%d">Footer Link %d</a>`, i, i))
	}
	sb.WriteString(`</nav>
			<p>Copyright 2026. All rights reserved.</p>
		</footer>
	</body></html>`)
	return sb.String()
}

// BenchmarkExtractContent_Small benchmarks ExtractContent with small HTML.
func BenchmarkExtractContent_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExtractContent(smallHTML)
	}
}

// BenchmarkExtractContent_Medium benchmarks ExtractContent with medium HTML.
func BenchmarkExtractContent_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExtractContent(mediumHTML)
	}
}

// BenchmarkExtractContent_Large benchmarks ExtractContent with large HTML.
func BenchmarkExtractContent_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExtractContent(largeHTML)
	}
}

// BenchmarkScoreNode benchmarks the scoreNode function.
func BenchmarkScoreNode(b *testing.B) {
	doc, _ := html.Parse(strings.NewReader(mediumHTML))
	body := findElement(doc, "body")
	article := findElement(body, "article")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scoreNode(article)
	}
}

// BenchmarkGetTextContent benchmarks the getTextContent function.
func BenchmarkGetTextContent(b *testing.B) {
	doc, _ := html.Parse(strings.NewReader(mediumHTML))
	body := findElement(doc, "body")
	article := findElement(body, "article")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getTextContent(article)
	}
}

// BenchmarkRemoveUnwantedElements benchmarks the removeUnwantedElements function.
func BenchmarkRemoveUnwantedElements(b *testing.B) {
	htmlWithScripts := `<html><body>
		<script>alert('test')</script>
		<style>.foo{}</style>
		<article><p>Content</p></article>
		<div hidden>Hidden</div>
	</body></html>`

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc, _ := html.Parse(strings.NewReader(htmlWithScripts))
		b.StartTimer()
		removeUnwantedElements(doc)
	}
}

// BenchmarkGetLinkTextLength benchmarks the getLinkTextLength function.
func BenchmarkGetLinkTextLength(b *testing.B) {
	htmlWithLinks := `<div>
		<p>Text before <a href="#">link one</a> and <a href="#">link two</a> after.</p>
		<p>More text with <a href="#">another link</a> here.</p>
	</div>`
	doc, _ := html.Parse(strings.NewReader(htmlWithLinks))
	div := findElement(doc, "div")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getLinkTextLength(div)
	}
}

// BenchmarkCountElements benchmarks the countElements function.
func BenchmarkCountElements(b *testing.B) {
	doc, _ := html.Parse(strings.NewReader(largeHTML))
	body := findElement(doc, "body")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countElements(body, "p")
	}
}

// BenchmarkFindBestCandidate benchmarks the findBestCandidate function.
func BenchmarkFindBestCandidate(b *testing.B) {
	doc, _ := html.Parse(strings.NewReader(largeHTML))
	removeUnwantedElements(doc)
	body := findElement(doc, "body")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findBestCandidate(body)
	}
}
