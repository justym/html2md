package main

import (
	"testing"
)

func TestConvertHeadings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"h1", "<h1>Title</h1>", "# Title"},
		{"h2", "<h2>Subtitle</h2>", "## Subtitle"},
		{"h3", "<h3>Section</h3>", "### Section"},
		{"h4", "<h4>Subsection</h4>", "#### Subsection"},
		{"h5", "<h5>Minor</h5>", "##### Minor"},
		{"h6", "<h6>Smallest</h6>", "###### Smallest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertParagraphs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single paragraph", "<p>Hello world</p>", "Hello world"},
		{"multiple paragraphs", "<p>First</p><p>Second</p>", "First\n\nSecond"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertBold(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"strong", "<strong>bold</strong>", "**bold**"},
		{"b tag", "<b>bold</b>", "**bold**"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertItalic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"em", "<em>italic</em>", "*italic*"},
		{"i tag", "<i>italic</i>", "*italic*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple link", `<a href="https://example.com">Example</a>`, "[Example](https://example.com)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertImages(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with alt", `<img src="image.png" alt="An image">`, "![An image](image.png)"},
		{"without alt", `<img src="image.png">`, "![](image.png)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"inline code", "<code>fmt.Println()</code>", "`fmt.Println()`"},
		{"code with entities", "<code>&lt;div&gt;</code>", "`<div>`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertCodeBlocks(t *testing.T) {
	input := "<pre><code>func main() {}</code></pre>"
	expected := "```\nfunc main() {}\n```"

	result := Convert(input)
	if result != expected {
		t.Errorf("Convert(%q) = %q, want %q", input, result, expected)
	}
}

func TestConvertLists(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"unordered list",
			"<ul><li>Item 1</li><li>Item 2</li></ul>",
			"- Item 1\n- Item 2",
		},
		{
			"ordered list",
			"<ol><li>First</li><li>Second</li></ol>",
			"1. First\n2. Second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertBlockquote(t *testing.T) {
	input := "<blockquote>This is a quote</blockquote>"
	expected := "> This is a quote"

	result := Convert(input)
	if result != expected {
		t.Errorf("Convert(%q) = %q, want %q", input, result, expected)
	}
}

func TestConvertTable(t *testing.T) {
	input := `<table>
		<tr><th>Header 1</th><th>Header 2</th></tr>
		<tr><td>Cell 1</td><td>Cell 2</td></tr>
	</table>`
	expected := `| Header 1 | Header 2 |
| --- | --- |
| Cell 1 | Cell 2 |`

	result := Convert(input)
	if result != expected {
		t.Errorf("Convert(%q) = %q, want %q", input, result, expected)
	}
}

func TestConvertHorizontalRule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"hr", "<hr>", "---"},
		{"hr self-closing", "<hr/>", "---"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Convert(tt.input)
			if result != tt.expected {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertLineBreak(t *testing.T) {
	input := "Line 1<br>Line 2"
	expected := "Line 1  \nLine 2"

	result := Convert(input)
	if result != expected {
		t.Errorf("Convert(%q) = %q, want %q", input, result, expected)
	}
}

func TestConvertMixed(t *testing.T) {
	input := `<h1>Title</h1><p>This is <strong>bold</strong> and <em>italic</em>.</p>`
	expected := `# Title

This is **bold** and *italic*.`

	result := Convert(input)
	if result != expected {
		t.Errorf("Convert(%q) = %q, want %q", input, result, expected)
	}
}
