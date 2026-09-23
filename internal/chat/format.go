package chat

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const maxDocRunes = 24000

func htmlToMarkdown(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return clipRunes(strings.Join(strings.Fields(raw), " "), maxDocRunes)
	}
	var b strings.Builder
	writeMarkdown(&b, findContentRoot(doc), 0)
	out := strings.TrimSpace(collapseBlanks(b.String()))
	return clipRunes(out, maxDocRunes)
}

func findContentRoot(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	var found *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found != nil || node == nil || node.Type != html.ElementNode {
			if node != nil {
				for c := node.FirstChild; c != nil && found == nil; c = c.NextSibling {
					walk(c)
				}
			}
			return
		}
		if node.Data == "main" || node.Data == "article" || hasClass(node, "body") || hasClass(node, "document") || attr(node, "role") == "main" {
			found = node
			return
		}
		for c := node.FirstChild; c != nil && found == nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	if found != nil {
		return found
	}
	return n
}

func writeMarkdown(b *strings.Builder, n *html.Node, listDepth int) {
	if n == nil {
		return
	}
	switch n.Type {
	case html.TextNode:
		t := strings.ReplaceAll(n.Data, "\u00a0", " ")
		if strings.TrimSpace(t) == "" {
			if strings.ContainsAny(t, "\n") {
				return
			}
			if b.Len() > 0 && !strings.HasSuffix(b.String(), " ") && !strings.HasSuffix(b.String(), "\n") {
				b.WriteByte(' ')
			}
			return
		}
		b.WriteString(strings.Join(strings.Fields(t), " "))
		return
	case html.ElementNode:
		if dropTag[n.Data] {
			return
		}
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			level := int(n.Data[1] - '0')
			b.WriteString("\n\n")
			b.WriteString(strings.Repeat("#", level))
			b.WriteByte(' ')
			writeInline(b, n)
			b.WriteString("\n\n")
			return
		case "p":
			b.WriteString("\n\n")
			writeInline(b, n)
			b.WriteString("\n\n")
			return
		case "br":
			b.WriteString("\n")
			return
		case "hr":
			b.WriteString("\n\n---\n\n")
			return
		case "pre":
			b.WriteString("\n\n```\n")
			b.WriteString(strings.TrimRight(textContent(n), "\n"))
			b.WriteString("\n```\n\n")
			return
		case "code":
			if n.Parent != nil && n.Parent.Data == "pre" {
				writeInline(b, n)
				return
			}
			b.WriteByte('`')
			writeInline(b, n)
			b.WriteByte('`')
			return
		case "ul", "ol":
			b.WriteString("\n")
			i := 0
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "li" {
					i++
					b.WriteString(strings.Repeat("  ", listDepth))
					if n.Data == "ol" {
						b.WriteString(itoa(i))
						b.WriteString(". ")
					} else {
						b.WriteString("- ")
					}
					writeInline(b, c)
					b.WriteByte('\n')
					for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
						if gc.Type == html.ElementNode && (gc.Data == "ul" || gc.Data == "ol") {
							writeMarkdown(b, gc, listDepth+1)
						}
					}
				}
			}
			b.WriteByte('\n')
			return
		case "blockquote":
			var inner strings.Builder
			writeInline(&inner, n)
			for _, line := range strings.Split(strings.TrimSpace(inner.String()), "\n") {
				b.WriteString("\n> ")
				b.WriteString(strings.TrimSpace(line))
			}
			b.WriteString("\n\n")
			return
		case "a":
			href := attr(n, "href")
			var inner strings.Builder
			writeInline(&inner, n)
			label := strings.TrimSpace(inner.String())
			if href != "" && label != "" && !strings.HasPrefix(href, "#") {
				b.WriteByte('[')
				b.WriteString(label)
				b.WriteString("](")
				b.WriteString(href)
				b.WriteByte(')')
				return
			}
			if label != "" {
				b.WriteString(label)
			}
			return
		case "strong", "b":
			b.WriteString("**")
			writeInline(b, n)
			b.WriteString("**")
			return
		case "em", "i":
			b.WriteByte('*')
			writeInline(b, n)
			b.WriteByte('*')
			return
		case "table":
			b.WriteString("\n\n")
			writeTable(b, n)
			b.WriteString("\n\n")
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writeMarkdown(b, c, listDepth)
	}
}

func writeInline(b *strings.Builder, n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol" || c.Data == "table" || c.Data == "pre") {
			continue
		}
		writeMarkdown(b, c, 0)
	}
}

func writeTable(b *strings.Builder, table *html.Node) {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var row []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					row = append(row, strings.TrimSpace(textContent(c)))
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	if len(rows) == 0 {
		return
	}
	writeRow := func(cells []string) {
		b.WriteString("| ")
		b.WriteString(strings.Join(cells, " | "))
		b.WriteString(" |\n")
	}
	writeRow(rows[0])
	sep := make([]string, len(rows[0]))
	for i := range sep {
		sep[i] = "---"
	}
	writeRow(sep)
	for _, row := range rows[1:] {
		writeRow(row)
	}
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textContent(c))
	}
	return b.String()
}

func attr(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, name string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == name {
			return true
		}
	}
	return false
}

var dropTag = map[string]bool{
	"script": true, "style": true, "iframe": true, "object": true, "embed": true,
	"form": true, "nav": true, "footer": true, "noscript": true, "svg": true,
}

func collapseBlanks(s string) string {
	var b strings.Builder
	nl := 0
	for _, r := range s {
		if r == '\n' {
			nl++
			if nl <= 2 {
				b.WriteRune(r)
			}
			continue
		}
		nl = 0
		if r == '\t' || !unicode.IsSpace(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func clipRunes(s string, n int) string {
	s = strings.TrimSpace(s)
	if n <= 0 || utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func formatDocument(title, url, body string) string {
	var b strings.Builder
	b.WriteString("### Document: ")
	if strings.TrimSpace(title) != "" {
		b.WriteString(strings.TrimSpace(title))
	} else {
		b.WriteString(url)
	}
	b.WriteByte('\n')
	if url != "" {
		b.WriteString("Source: ")
		b.WriteString(url)
		b.WriteString("\n\n")
	}
	b.WriteString(strings.TrimSpace(body))
	return strings.TrimSpace(b.String())
}
