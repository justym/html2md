package main

import (
	"strings"
	"testing"
)

func BenchmarkConvert(b *testing.B) {
	type args struct {
		html string
	}
	benchmarks := []struct {
		name string
		args args
	}{
		{
			name: "シンプルなHTML",
			args: args{html: "<h1>Title</h1><p>Hello world</p>"},
		},
		{
			name: "インライン要素を含むHTML",
			args: args{html: `<h1>Title</h1><p>This is <strong>bold</strong> and <em>italic</em>.</p>`},
		},
		{
			name: "複合要素を含むHTML",
			args: args{html: `
		<h1>Main Title</h1>
		<p>This is a <strong>bold</strong> and <em>italic</em> paragraph with a <a href="https://example.com">link</a>.</p>
		<h2>Section</h2>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
			<li>Item 3</li>
		</ul>
		<blockquote>This is a quote</blockquote>
		<pre><code>func main() {
	fmt.Println("Hello")
}</code></pre>
		<table>
			<tr><th>Header 1</th><th>Header 2</th></tr>
			<tr><td>Cell 1</td><td>Cell 2</td></tr>
		</table>
	`},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				Convert(bm.args.html)
			}
		})
	}
}

func BenchmarkConvert_LargeDocument(b *testing.B) {
	base := `<h2>Section</h2><p>Paragraph with <strong>bold</strong> and <em>italic</em> text.</p>`
	var sb strings.Builder
	sb.WriteString("<h1>Document Title</h1>")
	for range 100 {
		sb.WriteString(base)
	}
	input := sb.String()

	b.ResetTimer()
	for b.Loop() {
		Convert(input)
	}
}

func BenchmarkConvertInternal(b *testing.B) {
	type args struct {
		html string
	}
	benchmarks := []struct {
		name string
		fn   func(string) string
		args args
	}{
		{
			name: "convertHeadings",
			fn:   convertHeadings,
			args: args{html: "<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3>"},
		},
		{
			name: "convertLinks",
			fn:   convertLinks,
			args: args{html: `<a href="https://example.com">Link 1</a> and <a href="https://test.com">Link 2</a>`},
		},
		{
			name: "convertTables",
			fn:   convertTables,
			args: args{html: `<table>
		<tr><th>H1</th><th>H2</th><th>H3</th></tr>
		<tr><td>A1</td><td>A2</td><td>A3</td></tr>
		<tr><td>B1</td><td>B2</td><td>B3</td></tr>
	</table>`},
		},
		{
			name: "convertBold",
			fn:   convertBold,
			args: args{html: "<strong>text</strong> and <b>more</b>"},
		},
		{
			name: "convertLists",
			fn:   convertLists,
			args: args{html: "<ul><li>A</li><li>B</li></ul><ol><li>1</li><li>2</li></ol>"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				bm.fn(bm.args.html)
			}
		})
	}
}
