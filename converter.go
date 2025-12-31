package main

import (
	"regexp"
	"strconv"
	"strings"
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
	re := regexp.MustCompile(`[ \t]+`)
	return re.ReplaceAllString(s, " ")
}

func convertHeadings(s string) string {
	for i := 6; i >= 1; i-- {
		pattern := regexp.MustCompile(`(?i)<h` + string(rune('0'+i)) + `[^>]*>(.*?)</h` + string(rune('0'+i)) + `>`)
		prefix := strings.Repeat("#", i) + " "
		s = pattern.ReplaceAllStringFunc(s, func(match string) string {
			inner := pattern.FindStringSubmatch(match)[1]
			inner = strings.TrimSpace(inner)
			return "\n\n" + prefix + inner + "\n\n"
		})
	}
	return s
}

func convertParagraphs(s string) string {
	re := regexp.MustCompile(`(?i)<p[^>]*>(.*?)</p>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		inner = strings.TrimSpace(inner)
		return "\n\n" + inner + "\n\n"
	})
}

func convertBlockquotes(s string) string {
	re := regexp.MustCompile(`(?is)<blockquote[^>]*>(.*?)</blockquote>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
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
	re := regexp.MustCompile(`(?is)<pre[^>]*><code[^>]*>(.*?)</code></pre>`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		return "\n\n```\n" + inner + "\n```\n\n"
	})

	// Handle pre without code
	re = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		return "\n\n```\n" + inner + "\n```\n\n"
	})

	return s
}

func convertHorizontalRules(s string) string {
	re := regexp.MustCompile(`(?i)<hr\s*/?>`)
	return re.ReplaceAllString(s, "\n\n---\n\n")
}

func convertLists(s string) string {
	// Unordered lists
	s = convertUnorderedLists(s)
	// Ordered lists
	s = convertOrderedLists(s)
	return s
}

func convertUnorderedLists(s string) string {
	re := regexp.MustCompile(`(?is)<ul[^>]*>(.*?)</ul>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, false) + "\n\n"
	})
}

func convertOrderedLists(s string) string {
	re := regexp.MustCompile(`(?is)<ol[^>]*>(.*?)</ol>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		return "\n\n" + convertListItems(inner, true) + "\n\n"
	})
}

func convertListItems(s string, ordered bool) string {
	re := regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	matches := re.FindAllStringSubmatch(s, -1)
	var items []string
	for i, match := range matches {
		content := strings.TrimSpace(match[1])
		// Remove nested p tags
		content = regexp.MustCompile(`(?i)</?p[^>]*>`).ReplaceAllString(content, "")
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
	re := regexp.MustCompile(`(?is)<table[^>]*>(.*?)</table>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		return "\n\n" + convertTableContent(inner) + "\n\n"
	})
}

func convertTableContent(s string) string {
	// Extract rows
	rowRe := regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	rows := rowRe.FindAllStringSubmatch(s, -1)

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
	thRe := regexp.MustCompile(`(?is)<th[^>]*>(.*?)</th>`)
	tdRe := regexp.MustCompile(`(?is)<td[^>]*>(.*?)</td>`)

	var cells []string
	thMatches := thRe.FindAllStringSubmatch(row, -1)
	tdMatches := tdRe.FindAllStringSubmatch(row, -1)

	for _, m := range thMatches {
		cells = append(cells, strings.TrimSpace(m[1]))
	}
	for _, m := range tdMatches {
		cells = append(cells, strings.TrimSpace(m[1]))
	}

	return cells
}

func convertLinks(s string) string {
	re := regexp.MustCompile(`(?is)<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)
	return re.ReplaceAllString(s, "[$2]($1)")
}

func convertImages(s string) string {
	// Handle both self-closing and regular img tags
	re := regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']*)["'][^>]*alt=["']([^"']*)["'][^>]*/?>`)
	s = re.ReplaceAllString(s, "![$2]($1)")

	// Handle img with src only (no alt or alt first)
	re = regexp.MustCompile(`(?i)<img[^>]*alt=["']([^"']*)["'][^>]*src=["']([^"']*)["'][^>]*/?>`)
	s = re.ReplaceAllString(s, "![$1]($2)")

	// Handle img with src only
	re = regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']*)["'][^>]*/?>`)
	s = re.ReplaceAllString(s, "![]($1)")

	return s
}

func convertBold(s string) string {
	re := regexp.MustCompile(`(?is)<(strong|b)[^>]*>(.*?)</(strong|b)>`)
	return re.ReplaceAllString(s, "**$2**")
}

func convertItalic(s string) string {
	re := regexp.MustCompile(`(?is)<(em|i)[^>]*>(.*?)</(em|i)>`)
	return re.ReplaceAllString(s, "*$2*")
}

func convertInlineCode(s string) string {
	re := regexp.MustCompile(`(?is)<code[^>]*>(.*?)</code>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		inner = decodeHTMLEntities(inner)
		// Escape any remaining angle brackets to prevent cleanup from removing them
		inner = strings.ReplaceAll(inner, "<", "\x00LT\x00")
		inner = strings.ReplaceAll(inner, ">", "\x00GT\x00")
		return "`" + inner + "`"
	})
}

func convertLineBreaks(s string) string {
	re := regexp.MustCompile(`(?i)<br\s*/?>`)
	return re.ReplaceAllString(s, "  \n")
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
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")

	// Restore escaped angle brackets in code
	s = strings.ReplaceAll(s, "\x00LT\x00", "<")
	s = strings.ReplaceAll(s, "\x00GT\x00", ">")

	// Decode remaining entities
	s = decodeHTMLEntities(s)

	// Normalize multiple newlines to max 2
	re = regexp.MustCompile(`\n{3,}`)
	s = re.ReplaceAllString(s, "\n\n")

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
