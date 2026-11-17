// Site-wide search (F-06): /search/?q= looks through the titles and full text of
// every croquis, rima and recuerdo, and the library's titles and authors. Server-side,
// so it works without JavaScript. Accents and case are ignored ("cafe" finds "café").

package main

import (
	"html"
	"sort"
	"strings"
	"unicode"
)

type searchDoc struct {
	Title, Path, Day, Section string
	title                     string   // folded title
	words                     []string // original words of the text
	folded                    []string // folded words, same positions
}

type searchBook struct {
	Title, Author, Path string
	folded              string
}

type searchIndex struct {
	docs  []searchDoc
	books []searchBook
}

// foldWord lowercases, strips accents and trims punctuation around a word.
func foldWord(w string) string {
	return strings.TrimFunc(foldKey(w), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
}

func buildSearchIndex(idx []addonPost, texts map[string]string, lib *library) *searchIndex {
	s := &searchIndex{}
	for _, p := range idx {
		words := strings.Fields(texts[p.Path])
		folded := make([]string, len(words))
		for i, w := range words {
			folded[i] = foldWord(w)
		}
		s.docs = append(s.docs, searchDoc{Title: p.Title, Path: p.Path, Day: p.Day, Section: p.Section,
			title: foldKey(p.Title), words: words, folded: folded})
	}
	if lib != nil {
		for _, b := range lib.Books {
			s.books = append(s.books, searchBook{Title: b.Title, Author: b.Author, Path: b.path(),
				folded: foldKey(b.Title + " " + b.Subtitle + " " + b.Author + " " + b.Series)})
		}
	}
	return s
}

// searchTerms splits a query into folded words (max 8).
func searchTerms(q string) []string {
	var out []string
	for _, w := range strings.Fields(q) {
		if f := foldWord(w); f != "" && len(out) < 8 {
			out = append(out, f)
		}
	}
	return out
}

// search returns matching posts (every term must appear, as the start of a word
// in the text or anywhere in the title) and books, best first.
func (s *searchIndex) search(q string, limit int) (posts []any, books []any, total int) {
	terms := searchTerms(q)
	if len(terms) == 0 {
		return nil, nil, 0
	}
	type hit struct {
		d     *searchDoc
		score int
		first int
	}
	var hits []hit
	for i := range s.docs {
		d := &s.docs[i]
		score, first, ok := 0, -1, true
		for _, t := range terms {
			n := 0
			if strings.Contains(d.title, t) {
				score += 20
				n++
			}
			for j, w := range d.folded {
				if strings.HasPrefix(w, t) {
					n++
					if first < 0 {
						first = j
					}
				}
			}
			if n == 0 {
				ok = false
				break
			}
			if n > 10 {
				n = 10
			}
			score += n
		}
		if ok {
			hits = append(hits, hit{d, score, first})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].d.Day > hits[j].d.Day
	})
	total = len(hits)
	for i, h := range hits {
		if i >= limit {
			break
		}
		posts = append(posts, map[string]any{"Title": h.d.Title, "Path": h.d.Path, "Day": h.d.Day,
			"Section": h.d.Section, "Snippet": snippet(h.d, h.first, terms)})
	}
	for _, b := range s.books {
		all := true
		for _, t := range terms {
			if !strings.Contains(b.folded, t) {
				all = false
				break
			}
		}
		if all && len(books) < 30 {
			books = append(books, map[string]any{"Title": b.Title, "Author": b.Author, "Path": b.Path})
		}
	}
	return posts, books, total
}

// snippet is ~30 words around the first match, matches wrapped in <mark>; HTML-escaped.
func snippet(d *searchDoc, first int, terms []string) string {
	if len(d.words) == 0 {
		return ""
	}
	if first < 0 {
		first = 0
	}
	from, to := first-10, first+20
	if from < 0 {
		from = 0
	}
	if to > len(d.words) {
		to = len(d.words)
	}
	var b strings.Builder
	if from > 0 {
		b.WriteString("… ")
	}
	for i := from; i < to; i++ {
		if i > from {
			b.WriteByte(' ')
		}
		w := html.EscapeString(d.words[i])
		match := false
		for _, t := range terms {
			if strings.HasPrefix(d.folded[i], t) {
				match = true
				break
			}
		}
		if match {
			b.WriteString("<mark>" + w + "</mark>")
		} else {
			b.WriteString(w)
		}
	}
	if to < len(d.words) {
		b.WriteString(" …")
	}
	return b.String()
}

// searchData is the template data for /search/.
func searchData(s *searchIndex, q string) map[string]any {
	q = strings.TrimSpace(q)
	if len(q) > 100 {
		q = q[:100]
	}
	posts, books, total := s.search(q, 50)
	return map[string]any{
		"q": q, "hasQuery": q != "", "posts": posts, "books": books, "total": total,
		"hasPosts": len(posts) > 0, "hasBooks": len(books) > 0, "more": total > len(posts),
		"none": q != "" && len(posts) == 0 && len(books) == 0,
	}
}
