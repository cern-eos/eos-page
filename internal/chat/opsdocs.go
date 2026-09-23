package chat

import (
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type opsSection struct {
	Path string
	Body string
}

type opsIndex struct {
	sections []opsSection
}

var fileSplit = regexp.MustCompile(`(?m)^={8,}\s*\nFILE:\s*(.+)\s*\n={8,}\s*\n`)

func loadOpsDocs(path string) *opsIndex {
	idx := &opsIndex{}
	path = strings.TrimSpace(path)
	if path == "" {
		return idx
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ask eos: ops docs not loaded (%s): %v", path, err)
		return idx
	}
	idx.sections = splitOpsDocs(string(raw))
	log.Printf("ask eos: loaded %d EOS ops doc sections from %s", len(idx.sections), path)
	return idx
}

func splitOpsDocs(raw string) []opsSection {
	locs := fileSplit.FindAllStringSubmatchIndex(raw, -1)
	if len(locs) == 0 {
		body := strings.TrimSpace(raw)
		if body == "" {
			return nil
		}
		return []opsSection{{Path: "all_docs.md", Body: body}}
	}
	out := make([]opsSection, 0, len(locs))
	for i, loc := range locs {
		path := strings.TrimSpace(raw[loc[2]:loc[3]])
		start := loc[1]
		end := len(raw)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		body := strings.TrimSpace(raw[start:end])
		if body == "" {
			continue
		}
		out = append(out, opsSection{Path: path, Body: body})
	}
	return out
}

func (idx *opsIndex) Search(query string, limit int) []opsSection {
	if idx == nil || len(idx.sections) == 0 || limit <= 0 {
		return nil
	}
	terms := searchTerms(query)
	if len(terms) == 0 {
		return nil
	}
	type scored struct {
		sec   opsSection
		score int
	}
	var hits []scored
	for _, sec := range idx.sections {
		blob := strings.ToLower(sec.Path + "\n" + sec.Body)
		score := 0
		for _, t := range terms {
			score += strings.Count(blob, t)
			if strings.Contains(sec.Path, t) {
				score += 4
			}
		}
		if score == 0 {
			continue
		}
		hits = append(hits, scored{sec: sec, score: score})
	}
	for i := 0; i < len(hits); i++ {
		best := i
		for j := i + 1; j < len(hits); j++ {
			if hits[j].score > hits[best].score {
				best = j
			}
		}
		hits[i], hits[best] = hits[best], hits[i]
	}
	if len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]opsSection, 0, len(hits))
	for _, h := range hits {
		sec := h.sec
		sec.Body = clipRunes(sec.Body, 4500)
		out = append(out, sec)
	}
	return out
}

func (idx *opsIndex) FormatHits(query string, limit int) string {
	hits := idx.Search(query, limit)
	if len(hits) == 0 {
		return ""
	}
	var b strings.Builder
	for _, h := range hits {
		b.WriteString("### EOS operators doc: ")
		b.WriteString(h.Path)
		b.WriteString("\n\n")
		b.WriteString(h.Body)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func searchTerms(q string) []string {
	var out []string
	seen := map[string]bool{}
	var buf strings.Builder
	flush := func() {
		w := strings.ToLower(buf.String())
		buf.Reset()
		if len(w) < 3 || seen[w] {
			return
		}
		seen[w] = true
		out = append(out, w)
	}
	for _, r := range q {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}
