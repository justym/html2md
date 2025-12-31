package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Precompiled regex patterns
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

// Convert converts HTML to Markdown
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

func normalizeWhitespace(s string) string {
	// Replace multiple whitespace with single space (except newlines in pre)
	return reWhitespace.ReplaceAllString(s, " ")
}

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

func convertParagraphs(s string) string {
	return reParagraph.ReplaceAllStringFunc(s, func(match string) string {
		inner := reParagraph.FindStringSubmatch(match)[1]
		inner = strings.TrimSpace(inner)
		return "\n\n" + inner + "\n\n"
	})
}

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

func convertHorizontalRules(s string) string {
	return reHr.ReplaceAllString(s, "\n\n---\n\n")
}

func convertLists(s string) string {
	// Unordered lists
	s = convertUnorderedLists(s)
	// Ordered lists
	s = convertOrderedLists(s)
	return s
}

func convertUnorderedLists(s string) string {
	return reUl.ReplaceAllStringFunc(s, func(match string) string {
		inner := reUl.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, false) + "\n\n"
	})
}

func convertOrderedLists(s string) string {
	return reOl.ReplaceAllStringFunc(s, func(match string) string {
		inner := reOl.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, true) + "\n\n"
	})
}

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

func convertTables(s string) string {
	return reTable.ReplaceAllStringFunc(s, func(match string) string {
		inner := reTable.FindStringSubmatch(match)[1]
		return "\n\n" + convertTableContent(inner) + "\n\n"
	})
}

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

func convertLinks(s string) string {
	return reLink.ReplaceAllString(s, "[$2]($1)")
}

func convertImages(s string) string {
	// Handle both self-closing and regular img tags
	s = reImgSrcAlt.ReplaceAllString(s, "![$2]($1)")

	// Handle img with src only (no alt or alt first)
	s = reImgAltSrc.ReplaceAllString(s, "![$1]($2)")

	// Handle img with src only
	s = reImgSrc.ReplaceAllString(s, "![]($1)")

	return s
}

func convertBold(s string) string {
	return reBold.ReplaceAllString(s, "**$2**")
}

func convertItalic(s string) string {
	return reItalic.ReplaceAllString(s, "*$2*")
}

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

func convertLineBreaks(s string) string {
	return reBr.ReplaceAllString(s, "  \n")
}

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
