// Package main provides content extraction functionality.
//
// This file implements a Readability-inspired algorithm to extract
// the main content from HTML documents, filtering out navigation,
// sidebars, advertisements, and other non-essential elements.
//
// # Scoring Algorithm Overview
//
// The content extraction uses a scoring algorithm inspired by Mozilla Readability.
// Each candidate container element (article, main, section, div) is scored,
// and the highest-scoring element is selected as the main content.
//
// # Processing Flow
//
//  1. Preprocessing: Remove unwanted elements (script, style, noscript, hidden elements)
//  2. Candidate Selection: Find all container elements (article, main, section, div)
//  3. Scoring: Calculate a score for each candidate
//  4. Selection: Choose the highest-scoring candidate
//
// # Score Calculation
//
// The total score for a node is calculated as:
//
//	Score = BaseScore + PatternScore + DensityScore + ParagraphBonus + PunctuationBonus
//
// Where:
//   - BaseScore: Initial score based on tag name (e.g., article=+25, nav=-25)
//   - PatternScore: ±25 based on class/id pattern matching
//   - DensityScore: (textLength - linkTextLength) / textLength * textLength / 100
//   - ParagraphBonus: +3 per <p> element
//   - PunctuationBonus: +1 per comma/、 (max 10), indicates prose content
package main

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Pattern matching for class/id attribute scoring.
//
// These patterns are matched against the combined class and id attributes
// of each candidate element. Matching adds or subtracts 25 points from the score.
var (
	// positivePattern matches class/id names that indicate main content.
	// Matches: article, body, content, entry, main, page, post, text, blog, story, hentry
	// Example: <div class="article-content"> → +25 points
	positivePattern = regexp.MustCompile(`(?i)(article|body|content|entry|main|page|post|text|blog|story|hentry)`)

	// negativePattern matches class/id names that indicate non-content areas.
	// Matches: comment, meta, footer, footnote, sidebar, widget, banner, advertis,
	//          ad-, ad_, popup, social, share, related, recommend, nav, menu, breadcrumb, header
	// Example: <aside class="sidebar"> → -25 points
	negativePattern = regexp.MustCompile(`(?i)(comment|meta|footer|footnote|sidebar|widget|banner|advertis|ad[-_]|popup|social|share|related|recommend|nav|menu|breadcrumb|header)`)
)

// tagScores defines the base score for each HTML tag.
//
// Semantic content tags receive positive scores:
//   - article, main: +25 (primary content containers)
//   - section: +10 (generic content section)
//   - div: +5 (generic container)
//   - p: +3 (paragraph, indicates text content)
//
// Non-content tags receive negative scores:
//   - header, footer, nav, aside, form: -25 (typically contain navigation, metadata, or forms)
var tagScores = map[string]float64{
	"article": 25,
	"main":    25,
	"section": 10,
	"div":     5,
	"p":       3,
	"header":  -25,
	"footer":  -25,
	"nav":     -25,
	"aside":   -25,
	"form":    -25,
}

// Tags to remove during preprocessing.
var unwantedTags = map[string]bool{
	"script":   true,
	"style":    true,
	"noscript": true,
	"iframe":   true,
	"svg":      true,
}

// candidateTags defines container elements that can be content candidates.
var candidateTags = map[string]bool{
	"article": true,
	"main":    true,
	"section": true,
	"div":     true,
}

// ExtractContent extracts the main content from an HTML document.
//
// It uses a scoring algorithm inspired by Mozilla Readability to identify
// the most likely content area of the page.
//
// Preconditions:
//   - rawHTML can be any string, including empty or invalid HTML
//
// Invariants:
//   - Original input is not modified
//   - Processing is deterministic
//
// Postconditions:
//   - Returns extracted main content as HTML string
//   - If extraction fails or no body tag exists, returns original input
func ExtractContent(rawHTML string) string {
	// Skip extraction for simple HTML without body tag (backward compatibility)
	if !strings.Contains(strings.ToLower(rawHTML), "<body") {
		return rawHTML
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML
	}

	// Remove unwanted elements
	removeUnwantedElements(doc)

	// Find body element
	body := findElement(doc, "body")
	if body == nil {
		return rawHTML
	}

	// Find best candidate
	candidate := findBestCandidate(body)
	if candidate == nil {
		return rawHTML
	}

	// Render the candidate back to HTML
	return renderNode(candidate)
}

// removeUnwantedElements removes script, style, and other non-content elements.
//
// Preconditions:
//   - n is a valid HTML node tree
//
// Postconditions:
//   - Unwanted elements are removed from the tree
//   - Hidden elements are removed
func removeUnwantedElements(n *html.Node) {
	var toRemove []*html.Node

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Check if tag should be removed
			if unwantedTags[node.Data] {
				toRemove = append(toRemove, node)
				return
			}
			// Check for hidden attribute
			for _, attr := range node.Attr {
				if attr.Key == "hidden" {
					toRemove = append(toRemove, node)
					return
				}
				if attr.Key == "style" {
					normalized := strings.ToLower(strings.ReplaceAll(attr.Val, " ", ""))
					if strings.Contains(normalized, "display:none") {
						toRemove = append(toRemove, node)
						return
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	// Remove collected nodes
	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// findElement finds the first element with the given tag name.
func findElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// findBestCandidate finds the node most likely to contain the main content.
//
// Preconditions:
//   - body is the body element of the document
//
// Postconditions:
//   - Returns the highest-scoring candidate node
//   - Returns nil if no suitable candidate is found
func findBestCandidate(body *html.Node) *html.Node {
	var bestNode *html.Node
	var bestScore float64 = -1000

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Only consider container elements
			if candidateTags[n.Data] {
				score := scoreNode(n)
				if score > bestScore {
					bestScore = score
					bestNode = n
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(body)

	// If no good candidate found, return body itself
	if bestNode == nil || bestScore < 0 {
		return body
	}

	return bestNode
}

// scoreNode calculates a content score for a node.
//
// The score is calculated as the sum of the following components:
//
// 1. Base Score (from tagScores):
//   - article, main: +25
//   - section: +10
//   - div: +5
//
// 2. Pattern Score (from class/id matching):
//   - positivePattern match: +25
//   - negativePattern match: -25
//
// 3. Text Density Score:
//   - Formula: density * textLength / 100
//   - Where density = (textLength - linkTextLength) / textLength
//   - Higher density means more regular text relative to link text
//   - Penalizes link-heavy navigation areas
//
// 4. Paragraph Bonus:
//   - +3 points per <p> element
//   - More paragraphs indicate article-like content
//
// 5. Comma Bonus:
//   - +1 point per comma (including Japanese comma 、)
//   - Maximum 10 points
//   - Commas indicate prose content rather than lists or navigation
//
// Example score calculation for a typical article:
//
//	<article class="post">        → Base: +25, Pattern: +25
//	  <p>Text, with commas.</p>   → Paragraphs: +3, Commas: +2
//	  <p>More text here.</p>      → Paragraphs: +3
//	</article>
//	Total: 25 + 25 + 6 + 2 + density_score = ~60+
func scoreNode(n *html.Node) float64 {
	var score float64

	// Base score from tag
	if s, ok := tagScores[n.Data]; ok {
		score += s
	}

	// Class/ID pattern matching
	className := getAttr(n, "class")
	id := getAttr(n, "id")
	combined := className + " " + id

	if positivePattern.MatchString(combined) {
		score += 25
	}
	if negativePattern.MatchString(combined) {
		score -= 25
	}

	// Get text content once for multiple calculations
	text := getTextContent(n)
	textLen := len(strings.TrimSpace(text))

	// Text density score
	// When all text is within links (textLen == linkTextLen), density becomes 0,
	// which correctly penalizes navigation-heavy elements.
	linkTextLen := getLinkTextLength(n)

	if textLen > 0 {
		density := float64(textLen-linkTextLen) / float64(textLen)
		score += density * float64(textLen) / 100
	}

	// Paragraph bonus
	pCount := countElements(n, "p")
	score += float64(pCount) * 3

	// Punctuation bonus (indicates prose)
	// Counts both standard comma (,) and Japanese comma (、)
	punctuationCount := min(strings.Count(text, ",")+strings.Count(text, "、"), 10)
	score += float64(punctuationCount)

	return score
}

// getAttr returns the value of an attribute on a node.
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// getTextContent returns all text content within a node.
func getTextContent(n *html.Node) string {
	var buf bytes.Buffer
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			buf.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return buf.String()
}

// getTextLength returns the total length of text content.
func getTextLength(n *html.Node) int {
	return len(strings.TrimSpace(getTextContent(n)))
}

// getLinkTextLength returns the total length of text within <a> tags.
func getLinkTextLength(n *html.Node) int {
	var total int
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			total += getTextLength(node)
			return // Don't recurse into links
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return total
}

// countElements counts the number of elements with the given tag name.
func countElements(n *html.Node, tag string) int {
	var count int
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tag {
			count++
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return count
}

// renderNode renders a node back to HTML string.
func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		return ""
	}
	return buf.String()
}
