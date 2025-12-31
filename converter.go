// Package main provides HTML to Markdown conversion functionality.
//
// This package converts HTML strings to Markdown format using regex-based parsing.
// It supports common HTML elements including headings, paragraphs, links, images,
// lists, tables, code blocks, and inline formatting.
//
// Preconditions:
//   - Input should be valid or semi-valid HTML string
//   - Empty string input is acceptable
//
// Invariants:
//   - All regex patterns are precompiled at package initialization
//   - Conversion is stateless and thread-safe
//
// Postconditions:
//   - Output is valid Markdown text
//   - Unsupported HTML tags are stripped from output
package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Precompiled regex patterns for HTML element matching.
// These are compiled once at package initialization for performance.
var (
	reWhitespace   = regexp.MustCompile(`[ \t]+`)
	reH1           = regexp.MustCompile(`(?i)<h1[^>]*>(.*?)</h1>`)
	reH2           = regexp.MustCompile(`(?i)<h2[^>]*>(.*?)</h2>`)
	reH3           = regexp.MustCompile(`(?i)<h3[^>]*>(.*?)</h3>`)
	reH4           = regexp.MustCompile(`(?i)<h4[^>]*>(.*?)</h4>`)
	reH5           = regexp.MustCompile(`(?i)<h5[^>]*>(.*?)</h5>`)
	reH6           = regexp.MustCompile(`(?i)<h6[^>]*>(.*?)</h6>`)
	reParagraph    = regexp.MustCompile(`(?i)<p[^>]*>(.*?)</p>`)
	reBlockquote   = regexp.MustCompile(`(?is)<blockquote[^>]*>(.*?)</blockquote>`)
	rePreCode      = regexp.MustCompile(`(?is)<pre[^>]*><code[^>]*>(.*?)</code></pre>`)
	rePre          = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	reHr           = regexp.MustCompile(`(?i)<hr\s*/?>`)
	reUl           = regexp.MustCompile(`(?is)<ul[^>]*>(.*?)</ul>`)
	reOl           = regexp.MustCompile(`(?is)<ol[^>]*>(.*?)</ol>`)
	reLi           = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	rePTag         = regexp.MustCompile(`(?i)</?p[^>]*>`)
	reTable        = regexp.MustCompile(`(?is)<table[^>]*>(.*?)</table>`)
	reRow          = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	reTh           = regexp.MustCompile(`(?is)<th[^>]*>(.*?)</th>`)
	reTd           = regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)
	reLink         = regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)
	reImgSrcAlt    = regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']*)["'][^>]*alt=["']([^"']*)["'][^>]*/?>`)
	reImgAltSrc    = regexp.MustCompile(`(?i)<img[^>]*alt=["']([^"']*)["'][^>]*src=["']([^"']*)["'][^>]*/?>`)
	reImgSrc       = regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']*)["'][^>]*/?>`)
	reBold         = regexp.MustCompile(`(?is)<(strong|b)[^>]*>(.*?)</(strong|b)>`)
	reItalic       = regexp.MustCompile(`(?is)<(em|i)[^>]*>(.*?)</(em|i)>`)
	reInlineCode   = regexp.MustCompile(`(?is)<code[^>]*>(.*?)</code>`)
	reBr           = regexp.MustCompile(`(?i)<br\s*/?>`)
	reHtmlTag      = regexp.MustCompile(`<[^>]*>`)
	reMultiNewline = regexp.MustCompile(`\n{3,}`)
)

// Convert transforms an HTML string into Markdown format.
//
// It processes block elements (headings, paragraphs, lists, tables, code blocks)
// first, then inline elements (links, images, bold, italic, code), and finally
// cleans up the output.
//
// Preconditions:
//   - html can be any string, including empty string
//   - html does not need to be well-formed XML
//
// Invariants:
//   - Processing order is deterministic: block elements before inline elements
//   - Original input is not modified
//
// Postconditions:
//   - Returns trimmed Markdown string
//   - All recognized HTML tags are converted to Markdown equivalents
//   - Unrecognized HTML tags are removed from output
//   - Multiple consecutive newlines are normalized to at most two
func Convert(html string) string {
	// Normalize whitespace and newlines
	html = normalizeWhitespace(html)

	// Process block elements first
	html = convertHeadings(html)
	html = convertParagraphs(html)
	html = convertBlockquotes(html)
	html = convertCodeBlocks(html)
	html = convertHorizontalRules(html)
	html = convertLists(html)
	html = convertTables(html)

	// Process inline elements
	html = convertLinks(html)
	html = convertImages(html)
	html = convertBold(html)
	html = convertItalic(html)
	html = convertInlineCode(html)
	html = convertLineBreaks(html)

	// Clean up
	html = cleanupOutput(html)

	return strings.TrimSpace(html)
}

// normalizeWhitespace collapses consecutive spaces and tabs into a single space.
//
// Preconditions:
//   - s is any string
//
// Invariants:
//   - Only horizontal whitespace (spaces and tabs) is affected
//   - Newlines are preserved
//
// Postconditions:
//   - All sequences of spaces/tabs are replaced with a single space
func normalizeWhitespace(s string) string {
	return reWhitespace.ReplaceAllString(s, " ")
}

// convertHeadings converts HTML heading tags (h1-h6) to Markdown headings.
//
// Preconditions:
//   - s may contain <h1> through <h6> tags
//
// Invariants:
//   - Headings are processed from h6 to h1 to handle nested cases correctly
//   - Tag attributes are ignored
//
// Postconditions:
//   - <h1> becomes "# text", <h2> becomes "## text", etc.
//   - Each heading is surrounded by blank lines
//   - Inner content is trimmed of whitespace
func convertHeadings(s string) string {
	headings := []struct {
		re     *regexp.Regexp
		prefix string
	}{
		{reH6, "###### "},
		{reH5, "##### "},
		{reH4, "#### "},
		{reH3, "### "},
		{reH2, "## "},
		{reH1, "# "},
	}
	for _, h := range headings {
		s = h.re.ReplaceAllStringFunc(s, func(match string) string {
			inner := h.re.FindStringSubmatch(match)[1]
			inner = strings.TrimSpace(inner)
			return "\n\n" + h.prefix + inner + "\n\n"
		})
	}
	return s
}

// convertParagraphs converts HTML <p> tags to plain text with surrounding blank lines.
//
// Preconditions:
//   - s may contain <p> tags
//
// Invariants:
//   - Tag attributes are ignored
//
// Postconditions:
//   - <p> tags are removed, content is preserved
//   - Each paragraph is surrounded by blank lines
//   - Inner content is trimmed of whitespace
func convertParagraphs(s string) string {
	return reParagraph.ReplaceAllStringFunc(s, func(match string) string {
		inner := reParagraph.FindStringSubmatch(match)[1]
		inner = strings.TrimSpace(inner)
		return "\n\n" + inner + "\n\n"
	})
}

// convertBlockquotes converts HTML <blockquote> tags to Markdown blockquotes.
//
// Preconditions:
//   - s may contain <blockquote> tags
//
// Invariants:
//   - Multi-line content is preserved with each line prefixed by "> "
//   - Empty lines within blockquote are removed
//
// Postconditions:
//   - Each non-empty line is prefixed with "> "
//   - Blockquote is surrounded by blank lines
func convertBlockquotes(s string) string {
	return reBlockquote.ReplaceAllStringFunc(s, func(match string) string {
		inner := reBlockquote.FindStringSubmatch(match)[1]
		inner = strings.TrimSpace(inner)
		lines := strings.Split(inner, "\n")
		var quoted []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				quoted = append(quoted, "> "+line)
			}
		}
		return "\n\n" + strings.Join(quoted, "\n") + "\n\n"
	})
}

// convertCodeBlocks converts HTML <pre> and <pre><code> tags to Markdown fenced code blocks.
//
// Preconditions:
//   - s may contain <pre><code>...</code></pre> or <pre>...</pre> tags
//
// Invariants:
//   - <pre><code> is processed before <pre> to avoid double conversion
//   - HTML entities inside code are decoded
//
// Postconditions:
//   - Code is wrapped in ``` fences
//   - Code block is surrounded by blank lines
func convertCodeBlocks(s string) string {
	s = rePreCode.ReplaceAllStringFunc(s, func(match string) string {
		inner := rePreCode.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		return "\n\n```\n" + inner + "\n```\n\n"
	})

	// Handle pre without code
	s = rePre.ReplaceAllStringFunc(s, func(match string) string {
		inner := rePre.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		return "\n\n```\n" + inner + "\n```\n\n"
	})

	return s
}

// convertHorizontalRules converts HTML <hr> tags to Markdown horizontal rules.
//
// Preconditions:
//   - s may contain <hr> or <hr/> tags
//
// Invariants:
//   - Both self-closing and non-self-closing forms are handled
//
// Postconditions:
//   - <hr> becomes "---" surrounded by blank lines
func convertHorizontalRules(s string) string {
	return reHr.ReplaceAllString(s, "\n\n---\n\n")
}

// convertLists converts both unordered and ordered HTML lists to Markdown.
//
// Preconditions:
//   - s may contain <ul> and/or <ol> tags with nested <li> items
//
// Invariants:
//   - Unordered lists are processed before ordered lists
//
// Postconditions:
//   - <ul> lists become "- item" format
//   - <ol> lists become "1. item" format with sequential numbering
func convertLists(s string) string {
	// Unordered lists
	s = convertUnorderedLists(s)
	// Ordered lists
	s = convertOrderedLists(s)
	return s
}

// convertUnorderedLists converts HTML <ul> tags to Markdown unordered lists.
//
// Preconditions:
//   - s may contain <ul> tags
//
// Invariants:
//   - Delegates item processing to convertListItems
//
// Postconditions:
//   - Each <li> becomes "- item"
//   - List is surrounded by blank lines
func convertUnorderedLists(s string) string {
	return reUl.ReplaceAllStringFunc(s, func(match string) string {
		inner := reUl.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, false) + "\n\n"
	})
}

// convertOrderedLists converts HTML <ol> tags to Markdown ordered lists.
//
// Preconditions:
//   - s may contain <ol> tags
//
// Invariants:
//   - Delegates item processing to convertListItems
//
// Postconditions:
//   - Each <li> becomes "N. item" with sequential numbering starting from 1
//   - List is surrounded by blank lines
func convertOrderedLists(s string) string {
	return reOl.ReplaceAllStringFunc(s, func(match string) string {
		inner := reOl.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, true) + "\n\n"
	})
}

// convertListItems extracts and formats list items from HTML <li> tags.
//
// Preconditions:
//   - s contains the inner content of a <ul> or <ol> tag
//   - ordered indicates whether to use numbered or bulleted format
//
// Invariants:
//   - Nested <p> tags within list items are removed
//   - Items are numbered sequentially starting from 1 for ordered lists
//
// Postconditions:
//   - Returns newline-separated list items
//   - Each item is prefixed with "- " (unordered) or "N. " (ordered)
func convertListItems(s string, ordered bool) string {
	matches := reLi.FindAllStringSubmatch(s, -1)
	var items []string
	for i, match := range matches {
		content := strings.TrimSpace(match[1])
		// Remove nested p tags
		content = rePTag.ReplaceAllString(content, "")
		content = strings.TrimSpace(content)
		if ordered {
			items = append(items, strconv.Itoa(i+1)+". "+content)
		} else {
			items = append(items, "- "+content)
		}
	}
	return strings.Join(items, "\n")
}

// convertTables converts HTML <table> tags to Markdown tables.
//
// Preconditions:
//   - s may contain <table> tags with <tr>, <th>, and <td> elements
//
// Invariants:
//   - Delegates row processing to convertTableContent
//
// Postconditions:
//   - Table is converted to pipe-delimited Markdown format
//   - Table is surrounded by blank lines
func convertTables(s string) string {
	return reTable.ReplaceAllStringFunc(s, func(match string) string {
		inner := reTable.FindStringSubmatch(match)[1]
		return "\n\n" + convertTableContent(inner) + "\n\n"
	})
}

// convertTableContent processes the inner content of an HTML table.
//
// Preconditions:
//   - s contains <tr> rows with <th> and/or <td> cells
//
// Invariants:
//   - First row is treated as header
//   - Separator row is inserted after header
//
// Postconditions:
//   - Returns pipe-delimited table with header separator
//   - Empty rows are skipped
//   - Returns empty string if no valid rows found
func convertTableContent(s string) string {
	// Extract rows
	rows := reRow.FindAllStringSubmatch(s, -1)

	if len(rows) == 0 {
		return ""
	}

	var result []string
	headerWritten := false

	for _, row := range rows {
		cells := extractCells(row[1])
		if len(cells) == 0 {
			continue
		}

		line := "| " + strings.Join(cells, " | ") + " |"
		result = append(result, line)

		// Add separator after first row (header)
		if !headerWritten {
			var sep strings.Builder
			sep.WriteString("|")
			for range cells {
				sep.WriteString(" --- |")
			}
			result = append(result, sep.String())
			headerWritten = true
		}
	}

	return strings.Join(result, "\n")
}

// extractCells extracts cell contents from an HTML table row.
//
// Preconditions:
//   - row contains <th> and/or <td> elements
//
// Invariants:
//   - <th> cells are extracted before <td> cells
//   - Cell content is trimmed of whitespace
//
// Postconditions:
//   - Returns slice of cell contents in order
//   - Returns empty slice if no cells found
func extractCells(row string) []string {
	// Try th first, then td
	var cells []string
	thMatches := reTh.FindAllStringSubmatch(row, -1)
	tdMatches := reTd.FindAllStringSubmatch(row, -1)

	for _, m := range thMatches {
		cells = append(cells, strings.TrimSpace(m[1]))
	}
	for _, m := range tdMatches {
		cells = append(cells, strings.TrimSpace(m[1]))
	}

	return cells
}

// convertLinks converts HTML <a> tags to Markdown link syntax.
//
// Preconditions:
//   - s may contain <a href="...">...</a> tags
//
// Invariants:
//   - Only href attribute is extracted; other attributes are ignored
//
// Postconditions:
//   - <a href="url">text</a> becomes [text](url)
func convertLinks(s string) string {
	return reLink.ReplaceAllString(s, "[$2]($1)")
}

// convertImages converts HTML <img> tags to Markdown image syntax.
//
// Preconditions:
//   - s may contain <img> tags with src and optional alt attributes
//
// Invariants:
//   - Handles both src-before-alt and alt-before-src attribute orders
//   - Handles images with src only (no alt)
//
// Postconditions:
//   - <img src="url" alt="text"> becomes ![text](url)
//   - <img src="url"> becomes ![](url)
func convertImages(s string) string {
	// Handle both self-closing and regular img tags
	s = reImgSrcAlt.ReplaceAllString(s, "![$2]($1)")

	// Handle img with src only (no alt or alt first)
	s = reImgAltSrc.ReplaceAllString(s, "![$1]($2)")

	// Handle img with src only
	s = reImgSrc.ReplaceAllString(s, "![]($1)")

	return s
}

// convertBold converts HTML <strong> and <b> tags to Markdown bold syntax.
//
// Preconditions:
//   - s may contain <strong> or <b> tags
//
// Invariants:
//   - Both <strong> and <b> are treated equivalently
//
// Postconditions:
//   - Content is wrapped in ** markers
func convertBold(s string) string {
	return reBold.ReplaceAllString(s, "**$2**")
}

// convertItalic converts HTML <em> and <i> tags to Markdown italic syntax.
//
// Preconditions:
//   - s may contain <em> or <i> tags
//
// Invariants:
//   - Both <em> and <i> are treated equivalently
//
// Postconditions:
//   - Content is wrapped in * markers
func convertItalic(s string) string {
	return reItalic.ReplaceAllString(s, "*$2*")
}

// convertInlineCode converts HTML <code> tags to Markdown inline code syntax.
//
// Preconditions:
//   - s may contain <code> tags (not wrapped in <pre>)
//
// Invariants:
//   - HTML entities inside code are decoded
//   - Angle brackets are temporarily escaped to survive cleanup
//
// Postconditions:
//   - Content is wrapped in backticks
//   - HTML entities like &lt; are converted to actual characters
func convertInlineCode(s string) string {
	return reInlineCode.ReplaceAllStringFunc(s, func(match string) string {
		inner := reInlineCode.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		// Escape any remaining angle brackets to prevent cleanup from removing them
		inner = strings.ReplaceAll(inner, "<", "\x00LT\x00")
		inner = strings.ReplaceAll(inner, ">", "\x00GT\x00")
		return "`" + inner + "`"
	})
}

// convertLineBreaks converts HTML <br> tags to Markdown line breaks.
//
// Preconditions:
//   - s may contain <br> or <br/> tags
//
// Invariants:
//   - Both self-closing and non-self-closing forms are handled
//
// Postconditions:
//   - <br> becomes two trailing spaces followed by newline
func convertLineBreaks(s string) string {
	return reBr.ReplaceAllString(s, "  \n")
}

// decodeHTMLEntities converts common HTML entities to their character equivalents.
//
// Preconditions:
//   - s may contain HTML entities
//
// Invariants:
//   - Only predefined entities are decoded
//   - Unknown entities are left unchanged
//
// Postconditions:
//   - &lt; &gt; &amp; &quot; &#39; &apos; &nbsp; are decoded
func decodeHTMLEntities(s string) string {
	replacements := map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": "\"",
		"&#39;":  "'",
		"&apos;": "'",
		"&nbsp;": " ",
	}
	for entity, char := range replacements {
		s = strings.ReplaceAll(s, entity, char)
	}
	return s
}

// cleanupOutput performs final cleanup on the converted Markdown output.
//
// Preconditions:
//   - s has been processed by all conversion functions
//   - s may contain remaining HTML tags and escape sequences
//
// Invariants:
//   - Escape sequences from convertInlineCode are restored before tag removal
//   - Markdown line breaks (two trailing spaces) are preserved
//
// Postconditions:
//   - All remaining HTML tags are removed
//   - Escape sequences are restored to actual characters
//   - Multiple consecutive newlines are normalized to at most two
//   - Trailing whitespace is removed (except Markdown line breaks)
func cleanupOutput(s string) string {
	// Remove remaining HTML tags
	s = reHtmlTag.ReplaceAllString(s, "")

	// Restore escaped angle brackets in code
	s = strings.ReplaceAll(s, "\x00LT\x00", "<")
	s = strings.ReplaceAll(s, "\x00GT\x00", ">")

	// Decode remaining entities
	s = decodeHTMLEntities(s)

	// Normalize multiple newlines to max 2
	s = reMultiNewline.ReplaceAllString(s, "\n\n")

	// Remove trailing whitespace from lines, but preserve markdown line breaks (two spaces before newline)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// Check if line ends with markdown line break (two spaces)
		if strings.HasSuffix(line, "  ") {
			// Preserve the two trailing spaces, trim any tabs
			lines[i] = strings.TrimRight(line, "\t")
		} else {
			lines[i] = strings.TrimRight(line, " \t")
		}
	}
	s = strings.Join(lines, "\n")

	return s
}
