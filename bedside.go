package main

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
)

// P-10: the bedsidetable pocket links each book to its library page, with a
// small cover, and the book page says when it was on the bedside table.
// The pocket stays plain markdown — "#### 2026.09" then "* _Title_ (2000) by
// Author" — so Carlos edits it as before; books that aren't in the library
// stay plain text.

var (
	bedsideItemRe  = regexp.MustCompile(`(?s)<h4[^>]*>(.*?)</h4>|<li><em>(.*?)</em>(.*?)</li>`)
	bedsideNonWord = regexp.MustCompile(`[^a-z0-9]+`)
)

// bedsideKey reduces a title to what two editions share: lower case, no
// accents, no leading article, nothing after a colon.
func bedsideKey(t string) string {
	t = foldKey(html.UnescapeString(t))
	if i := strings.Index(t, ":"); i > 0 {
		t = t[:i]
	}
	t = strings.Trim(bedsideNonWord.ReplaceAllString(t, " "), " ")
	for _, a := range []string{"the ", "el ", "la ", "los ", "las "} {
		t = strings.TrimPrefix(t, a)
	}
	return t
}

// bedsideMatch finds the library book for a pocket line: same title key, and
// the line shares a name with the book's author when the book has one
// ("Alan Moore" ~ "Alan Moore & Dave Gibbons", "Julián Ayesta" ~ "Julián
// Ayesta Prendes", "Akira Toriyama" ~ "鳥山 明 [Akira Toriyama]").
func bedsideMatch(lib *library, title, rest string) *libBook {
	key := bedsideKey(title)
	names := map[string]bool{}
	for _, w := range strings.Fields(bedsideNonWord.ReplaceAllString(foldKey(html.UnescapeString(rest)), " ")) {
		names[w] = true
	}
	for _, b := range lib.Books {
		if bedsideKey(b.Title) != key {
			continue
		}
		words := strings.Fields(bedsideNonWord.ReplaceAllString(foldKey(b.Author), " "))
		shared := len(words) == 0
		for _, w := range words {
			if len(w) >= 3 && names[w] {
				shared = true
			}
		}
		if shared {
			return b
		}
	}
	return nil
}

// linkBedside rewrites the pocket's HTML in place and records the months on
// each matched book (newest first, as the pocket lists them).
func linkBedside(p *Post, lib *library) {
	if p == nil || lib == nil {
		return
	}
	month := ""
	out := bedsideItemRe.ReplaceAllStringFunc(string(p.Content), func(m string) string {
		g := bedsideItemRe.FindStringSubmatch(m)
		if g[1] != "" {
			month = strings.TrimSpace(g[1])
			return m
		}
		b := bedsideMatch(lib, g[2], g[3])
		if b == nil {
			return m
		}
		if month != "" && (len(b.Bedside) == 0 || b.Bedside[len(b.Bedside)-1] != month) {
			b.Bedside = append(b.Bedside, month)
		}
		cover := ""
		if b.HasCover && b.ThumbH > 0 {
			h := 60
			cover = fmt.Sprintf(`<a class="bs-cover" href="%s" tabindex="-1" aria-hidden="true"><img src="/archives/library/covers/%s.jpg" alt="" width="%d" height="%d" loading="lazy"></a>`,
				b.path(), b.Key, b.ThumbW*h/b.ThumbH, h)
		}
		if cover == "" { // no cover: keep the plain list line, just linked
			return fmt.Sprintf(`<li><a href="%s"><em>%s</em></a>%s</li>`, b.path(), g[2], g[3])
		}
		return fmt.Sprintf(`<li class="bs-book">%s<span><a href="%s"><em>%s</em></a>%s</span></li>`, cover, b.path(), g[2], g[3])
	})
	p.Content = template.HTML(out)
}
