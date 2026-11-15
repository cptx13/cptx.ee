// Newsletter (batch K, U): the newsletter is every email sent to subscribers;
// the weekly ones are "issues". Issues live in content/newsletter/NNN.md, are
// shown at /newsletter/<n>/ and kept on the kiosk (/newsletter/kiosk/). How it
// works is the #newsletter section of the privacy policy.

package main

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

// isIssue reports whether p is a weekly issue (a numbered file in content/newsletter/).
func isIssue(p *Post) bool {
	if p.Section != "newsletter" {
		return false
	}
	_, err := strconv.Atoi(p.Slug)
	return err == nil
}

// issueLabel: "1" → "issue 01".
func issueLabel(slug string) string {
	n, err := strconv.Atoi(slug)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("issue %02d", n)
}

// coverIdx is the addon index (set at startup); its words per day light the
// week strip on each cover.
var coverIdx []addonPost

// pageTitleFor is a post's <title>: "Title | cptx", "issue 01: Title | cptx".
func pageTitleFor(p *Post) string {
	if isIssue(p) {
		return issueLabel(p.Slug) + ": " + p.Title + " | cptx"
	}
	return p.Title + " | cptx"
}

// newsletterIssues returns published issues, newest first, as template maps.
func newsletterIssues(posts []*Post) []any {
	var issues []*Post
	for _, p := range posts {
		if isIssue(p) && !p.Draft {
			issues = append(issues, p)
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		a, _ := strconv.Atoi(issues[i].Slug)
		b, _ := strconv.Atoi(issues[j].Slug)
		return a > b
	})
	words := map[string]int{}
	for _, p := range coverIdx {
		words[p.Day] += p.Words
	}
	out := make([]any, len(issues))
	for i, p := range issues {
		_, week := p.Date.ISOWeek()
		n, _ := strconv.Atoi(p.Slug)
		out[i] = map[string]any{
			"No": p.Slug, "Label": issueLabel(p.Slug), "Title": p.Title, "Path": p.Path,
			"Date": p.Date.Format("2 Jan 2006"), "Week": "week " + strconv.Itoa(week),
			"Summary": p.Summary, "Reads": readsLabel(siteStats.readsOf(p.Path)),
			"Palette": "kiosk-p" + strconv.Itoa((n-1+len(coverPalettes))%len(coverPalettes)),
			"Art":     coverArt(p, n, words),
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
	return map[string]any{"issues": shown, "hasIssues": len(all) > 0, "issueCount": len(all), "issueCountLabel": issueCountLabel(len(all)), "hasMore": len(all) > len(shown)}
}

// issueCountLabel: "1 issue", "2 issues".
func issueCountLabel(n int) string {
	if n == 1 {
		return "1 issue"
	}
	return strconv.Itoa(n) + " issues"
}

// coverPalettes: paper / shapes / soft shapes, one per issue in turn (CSS
// classes kiosk-p0…p4 set the paper and text colour to match).
var coverPalettes = [][3]string{
	{"#ffd166", "#e85d04", "#ffb703"}, // amber paper, orange shapes
	{"#1d1a14", "#ffb703", "#6b4a12"}, // ink paper, amber shapes
	{"#fff1d0", "#f48c06", "#fbb454"}, // cream paper, tangerine shapes
	{"#e85d04", "#ffd166", "#9d0208"}, // orange paper, butter shapes
	{"#6b4a12", "#ffd166", "#b07a12"}, // bronze paper, butter shapes
}

// coverArt draws an issue's cover pattern as an inline SVG (viewBox 150×200,
// behind the cover's text). Each issue gets a motif picked from its title and
// number, the issue number as a large outlined numeral, and a strip of seven
// tiles for the days of its week, lit by how much was written each day.
func coverArt(p *Post, n int, wordsByDay map[string]int) string {
	pal := coverPalettes[(n-1+len(coverPalettes))%len(coverPalettes)]
	strong, soft := pal[1], pal[2]
	h := fnv.New32a()
	h.Write([]byte(p.Title + "|" + p.Slug))
	seed := h.Sum32()
	var b strings.Builder
	b.WriteString(`<svg class="kiosk-art" viewBox="0 0 150 200" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false">`)

	switch seed % 5 {
	case 0: // concentric arcs from the top-right corner
		for r := 30; r <= 130; r += 20 {
			fmt.Fprintf(&b, `<circle cx="150" cy="0" r="%d" fill="none" stroke="%s" stroke-width="6" opacity="%.2f"/>`, r, strong, 0.25+float64(130-r)/200)
		}
	case 1: // dot grid that grows towards the top right
		for y := 0; y < 7; y++ {
			for x := 0; x < 6; x++ {
				rad := 1.0 + float64(x+6-y)*0.45
				fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%.1f" fill="%s"/>`, 22+x*22, 12+y*16, rad, strong)
			}
		}
	case 2: // diagonal stripes across the upper half
		fmt.Fprintf(&b, `<g transform="rotate(-30 75 60)">`)
		for i := 0; i < 7; i++ {
			c := strong
			if i%2 == 1 {
				c = soft
			}
			fmt.Fprintf(&b, `<rect x="-60" y="%d" width="270" height="9" fill="%s"/>`, 10+i*14, c)
		}
		b.WriteString(`</g>`)
	case 3: // rising sun: half disc and rays on the right
		for i := 0; i < 9; i++ {
			fmt.Fprintf(&b, `<rect x="146" y="78" width="80" height="4" rx="2" fill="%s" transform="rotate(%d 150 80)"/>`, soft, 100+i*20)
		}
		fmt.Fprintf(&b, `<circle cx="150" cy="80" r="40" fill="%s"/>`, strong)
	default: // checker of rounded tiles in the top-right corner
		for y := 0; y < 4; y++ {
			for x := 0; x < 4; x++ {
				if (x+y)%2 == 0 {
					fmt.Fprintf(&b, `<rect x="%d" y="%d" width="16" height="16" rx="3" fill="%s"/>`, 70+x*20, 26+y*20, strong)
				} else if (x*y+int(seed))%3 == 0 {
					fmt.Fprintf(&b, `<rect x="%d" y="%d" width="16" height="16" rx="3" fill="%s"/>`, 70+x*20, 26+y*20, soft)
				}
			}
		}
	}

	// the issue number, large and outlined
	fmt.Fprintf(&b, `<text x="146" y="150" text-anchor="end" font-family="JetBrains Mono, monospace" font-weight="700" font-size="64" fill="none" stroke="%s" stroke-width="1.4" opacity=".55">%02d</text>`, strong, n)

	// the week: Monday … Sunday of the issue's ISO week
	wd := int(p.Date.Weekday()+6) % 7 // Monday = 0
	monday := p.Date.AddDate(0, 0, -wd)
	for d := 0; d < 7; d++ {
		w := wordsByDay[monday.AddDate(0, 0, d).Format("2006-01-02")]
		x := 92 + d*7
		switch {
		case w >= 400:
			fmt.Fprintf(&b, `<rect x="%d" y="186" width="5" height="5" rx="1" fill="%s"/>`, x, strong)
		case w > 0:
			fmt.Fprintf(&b, `<rect x="%d" y="186" width="5" height="5" rx="1" fill="%s"/>`, x, soft)
		default:
			fmt.Fprintf(&b, `<rect x="%d.5" y="186.5" width="4" height="4" rx="1" fill="none" stroke="%s" stroke-width=".8" opacity=".6"/>`, x, strong)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}
