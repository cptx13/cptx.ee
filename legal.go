// Legal docs (/legaldocs/): privacy, cookies, how the site works.
// How the newsletter works is a section of the privacy policy (#newsletter).

package main

import "sort"

// legalMeta orders the hub cards and picks their icons; unknown docs go last.
var legalMeta = map[string]struct {
	Icon  string
	Order int
}{
	"privacy":      {"shield", 1},
	"cookies":      {"cookie", 2},
	"how-it-works": {"gear", 3},
}

// buildLegalData returns the cards for the /legaldocs/ hub.
func buildLegalData(posts []*Post) []any {
	docs := make([]*Post, 0, len(posts))
	for _, p := range posts {
		if p.Section == "legaldocs" {
			docs = append(docs, p)
		}
	}
	order := func(p *Post) int {
		if m, ok := legalMeta[p.Slug]; ok {
			return m.Order
		}
		return 99
	}
	sort.SliceStable(docs, func(i, j int) bool {
		if order(docs[i]) != order(docs[j]) {
			return order(docs[i]) < order(docs[j])
		}
		return docs[i].Title < docs[j].Title
	})
	out := make([]any, len(docs))
	for i, p := range docs {
		icon := "paper"
		if m, ok := legalMeta[p.Slug]; ok {
			icon = m.Icon
		}
		out[i] = map[string]any{
			"Title":   p.Title,
			"Href":    p.Path,
			"Desc":    p.Summary,
			"Icon":    icon,
			"Updated": p.DateFormatted,
		}
	}
	return out
}
