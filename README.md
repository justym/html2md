# html2md

A lightweight HTML to Markdown converter written in Go.

## Features

- Automatic content extraction (Readability-inspired algorithm)
- Filters out navigation, sidebars, and advertisements
- Reads from stdin, writes to stdout
- Supports common HTML elements

## Installation

```bash
go install github.com/justym/html2md@latest
```

Or build from source:

```bash
git clone https://github.com/justym/html2md.git
cd html2md
go build
```

## Usage

```bash
# Convert HTML file to Markdown
cat index.html | html2md > output.md

# Convert from curl output
curl -s https://example.com | html2md

# Direct input
echo "<h1>Hello</h1><p>World</p>" | html2md
```

## Supported HTML Elements

| HTML | Markdown |
|------|----------|
| `<h1>` - `<h6>` | `#` - `######` |
| `<p>` | Plain text with blank lines |
| `<strong>`, `<b>` | `**bold**` |
| `<em>`, `<i>` | `*italic*` |
| `<a href="...">` | `[text](url)` |
| `<img src="..." alt="...">` | `![alt](src)` |
| `<code>` | `` `code` `` |
| `<pre><code>` | Fenced code block |
| `<ul>`, `<ol>`, `<li>` | `- item` / `1. item` |
| `<blockquote>` | `> quote` |
| `<table>` | Pipe table |
| `<hr>` | `---` |
| `<br>` | Two trailing spaces + newline |

## Examples

### Input

```html
<h1>Welcome</h1>
<p>This is <strong>bold</strong> and <em>italic</em> text.</p>
<ul>
  <li>Item 1</li>
  <li>Item 2</li>
</ul>
```

### Output

```markdown
# Welcome

This is **bold** and *italic* text.

- Item 1
- Item 2
```

## License

MIT License. See [LICENSE](LICENSE) for details.

This project includes code inspired by [Mozilla Readability](https://github.com/mozilla/readability),
which is licensed under the Apache License 2.0. See [NOTICE](NOTICE) for details.
