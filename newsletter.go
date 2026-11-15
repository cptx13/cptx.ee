// Newsletter (batch K): weekly issues live in content/newsletter/NNN.md and are
// shown at /newsletter/<n>/; /newsletter/kiosk/ shows every issue as a cover.

package main

import (
	"sort"
	"strconv"
)

// newsletterIssues returns published issues, newest first, as template maps.
func newsletterIssues(posts []*Post) []any {
	var issues []*Post
	for _, p := range posts {
		if p.Section == "newsletter" && !p.Draft {
			if _, err := strconv.Atoi(p.Slug); err == nil {
				issues = append(issues, p)
			}
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		a, _ := strconv.Atoi(issues[i].Slug)
		b, _ := strconv.Atoi(issues[j].Slug)
		return a > b
	})
	out := make([]any, len(issues))
	for i, p := range issues {
		_, week := p.Date.ISOWeek()
		out[i] = map[string]any{
			"No": p.Slug, "Title": p.Title, "Path": p.Path,
			"Date": p.Date.Format("2 Jan 2006"), "Week": "week " + strconv.Itoa(week),
			"Summary": p.Summary,
		}
	}
	return out
}

// newsletterData is the template data for /newsletter/ (latest covers) and /newsletter/kiosk/ (all).
func newsletterData(posts []*Post, latest int) map[string]any {
	all := newsletterIssues(posts)
	shown := all
	if latest > 0 && len(shown) > latest {
		shown = shown[:latest]
	}
	return map[string]any{"issues": shown, "hasIssues": len(all) > 0, "issueCount": len(all), "hasMore": len(all) > len(shown)}
}
