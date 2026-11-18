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

// isIssue reports whether p is a weekly issue (a numbered file in content/newsletter/),
// in English. Its Spanish version (NNN.es.md, /newsletter/<n>/es/) is isIssueES.
func isIssue(p *Post) bool {
	return isNumberedNewsletter(p) && !strings.HasSuffix(p.Path, "/es/")
}

// isIssueES reports whether p is the Spanish version of an issue.
func isIssueES(p *Post) bool {
	return isNumberedNewsletter(p) && strings.HasSuffix(p.Path, "/es/")
}

func isNumberedNewsletter(p *Post) bool {
	if p.Section != "newsletter" {
		return false
	}
	_, err := strconv.Atoi(p.Slug)
	return err == nil
}

// issueAltPath gives, for an issue or its Spanish version, the path of the
// other language when it exists ("" otherwise).
func issueAltPath(p *Post, byPath map[string]*Post) string {
	alt := ""
	switch {
	case isIssue(p):
		alt = p.Path + "es/"
	case isIssueES(p):
		alt = strings.TrimSuffix(p.Path, "es/")
	default:
		return ""
	}
	if q, ok := byPath[alt]; ok && !q.Draft {
		return alt
	}
	return ""
}

// issueLabel: "1" → "issue 01".
func issueLabel(slug string) string {
	n, err := strconv.Atoi(slug)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("issue %02d", n)
}

// buttondownUser is the newsletter's Buttondown username. While it is empty the
// sign-up form on /newsletter/ stays closed ("opens soon"); once set, the form
// posts to Buttondown, which sends the confirmation email (double opt-in).
const buttondownUser = "cptx"

// newsletterNotices are the pages Buttondown redirects to (Settings →
// Subscribing → Redirects): slug → {title, text}.
var newsletterNotices = map[string][2]string{
	"thanks":  {"almost there", "Thanks for subscribing. Check your inbox: an email from Buttondown asks you to confirm, and nothing is sent until you do. If it isn't there in a few minutes, look in spam."},
	"welcome": {"welcome", "You're subscribed. The next issue arrives by email, in English and Spanish; until then, every past issue is on the kiosk."},
}

// coverIdx is the addon index (set at startup); its words per day light the
// week strip on each cover.
var coverIdx []addonPost

// pageTitleFor is a post's <title>: "Title | cptx", "issue 01: Title | cptx".
func pageTitleFor(p *Post) string {
	if isIssue(p) || isIssueES(p) {
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
			"No": p.Slug, "NoPadded": fmt.Sprintf("%02d", n), "Label": issueLabel(p.Slug), "Title": p.Title, "Path": p.Path,
			"Date": p.Date.Format("2 Jan 2006"), "Week": "week " + strconv.Itoa(week),
			"Summary": p.Summary, "Reads": readsLabel(siteStats.visitsOf(p.Path)),
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
	return map[string]any{"notice": "", "noticeTitle": "", "signupOpen": buttondownUser != "", "signupAction": "https://buttondown.com/api/emails/embed-subscribe/" + buttondownUser,
		"issues": shown, "hasIssues": len(all) > 0, "issueCount": len(all), "issueCountLabel": issueCountLabel(len(all)), "hasMore": len(all) > len(shown)}
}

// issueCountLabel: "1 issue", "2 issues".
func issueCountLabel(n int) string {
	if n == 1 {
		return "1 issue"
	}
	return strconv.Itoa(n) + " issues"
}

// coverPalettes: paper, then three art colours (a, b, and c for the art
// panel's ground), one palette per issue in turn. CSS classes kiosk-p0…p4 set
// the matching text colours.
var coverPalettes = [][4]string{
	{"#ffd166", "#e85d04", "#9d0208", "#fff1d0"}, // amber paper
	{"#1d1a14", "#ffb703", "#e85d04", "#3a2a0c"}, // ink paper
	{"#fff1d0", "#f48c06", "#6b4a12", "#fbb454"}, // cream paper
	{"#e85d04", "#ffd166", "#3a2a0c", "#9d0208"}, // orange paper
	{"#6b4a12", "#ffd166", "#e85d04", "#3a2a0c"}, // bronze paper
}

// coverArt draws an issue's cover as an inline SVG (viewBox 150×200) behind
// its text: a paper grain, a darker spine, a full-width art panel (y 52–136)
// with one of six motifs picked from the title and number, and a label with
// seven tiles for the days of the issue's week, lit by how much was written.
// The masthead (with the number) and the headline sit above and below the
// panel, so the art never runs under text.
func coverArt(p *Post, n int, wordsByDay map[string]int) string {
	pal := coverPalettes[(n-1+len(coverPalettes))%len(coverPalettes)]
	paper, a, bb, c := pal[0], pal[1], pal[2], pal[3]
	h := fnv.New32a()
	h.Write([]byte(p.Title + "|" + p.Slug))
	seed := h.Sum32()
	var b strings.Builder
	b.WriteString(`<svg class="kiosk-art" viewBox="0 0 150 200" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false">`)
	fmt.Fprintf(&b, `<defs><clipPath id="kc%d"><rect x="0" y="52" width="150" height="84"/></clipPath>`+
		`<filter id="kg%d"><feTurbulence type="fractalNoise" baseFrequency=".85" numOctaves="2" stitchTiles="stitch"/><feColorMatrix type="saturate" values="0"/></filter></defs>`, n, n)
	fmt.Fprintf(&b, `<rect width="150" height="200" fill="%s"/>`, paper)
	fmt.Fprintf(&b, `<g clip-path="url(#kc%d)"><rect x="0" y="52" width="150" height="84" fill="%s"/>`, n, c)

	switch seed % 6 {
	case 0: // sun over a horizon, with rays
		for i := 0; i < 12; i++ {
			fmt.Fprintf(&b, `<rect x="104" y="117" width="120" height="3" fill="%s" transform="rotate(%d 104 118)"/>`, bb, 180+i*15)
		}
		fmt.Fprintf(&b, `<circle cx="104" cy="118" r="30" fill="%s"/><rect x="0" y="118" width="150" height="18" fill="%s"/>`, a, bb)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(&b, `<rect x="0" y="%d" width="150" height="1.6" fill="%s"/>`, 122+i*5, c)
		}
	case 1: // a rainbow of half rings rising from the bottom edge
		for i, r := range []int{70, 58, 46, 34, 22} {
			col := a
			if i%2 == 1 {
				col = bb
			}
			fmt.Fprintf(&b, `<circle cx="%d" cy="136" r="%d" fill="%s"/>`, 40+int(seed%70), r, col)
		}
		fmt.Fprintf(&b, `<circle cx="%d" cy="136" r="10" fill="%s"/>`, 40+int(seed%70), c)
	case 2: // bold diagonal stripes
		b.WriteString(`<g transform="rotate(-35 75 94)">`)
		for i := 0; i < 12; i++ {
			col := a
			if i%3 == 1 {
				col = bb
			} else if i%3 == 2 {
				continue
			}
			fmt.Fprintf(&b, `<rect x="-40" y="%d" width="230" height="10" fill="%s"/>`, 6+i*15, col)
		}
		b.WriteString(`</g>`)
	case 3: // Bauhaus tiles: quarter circles turned by the seed
		k := 0
		for y := 52; y < 136; y += 21 {
			for x := 0; x < 150; x += 21 {
				r := (seed >> (uint(k) % 29)) & 3
				col := a
				if (seed>>(uint(k+7)%29))&1 == 1 {
					col = bb
				}
				cx, cy := x, y
				switch r {
				case 1:
					cx = x + 21
				case 2:
					cx, cy = x+21, y+21
				case 3:
					cy = y + 21
				}
				fmt.Fprintf(&b, `<clipPath id="kt%d-%d"><rect x="%d" y="%d" width="21" height="21"/></clipPath><circle clip-path="url(#kt%d-%d)" cx="%d" cy="%d" r="21" fill="%s"/>`, n, k, x, y, n, k, cx, cy, col)
				k++
			}
		}
	case 4: // hills under a moon
		fmt.Fprintf(&b, `<circle cx="%d" cy="72" r="11" fill="%s"/>`, 30+int(seed%90), a)
		fmt.Fprintf(&b, `<ellipse cx="40" cy="140" rx="80" ry="38" fill="%s"/>`, bb)
		fmt.Fprintf(&b, `<ellipse cx="125" cy="146" rx="70" ry="36" fill="%s"/>`, a)
		fmt.Fprintf(&b, `<ellipse cx="70" cy="150" rx="60" ry="22" fill="%s" opacity=".85"/>`, bb)
	default: // halftone: dots growing across the panel
		for y := 0; y < 6; y++ {
			for x := 0; x < 11; x++ {
				r := 0.6 + float64(x)*0.62 + float64(y%2)*0.3
				fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%.1f" fill="%s"/>`, 7+x*14+(y%2)*7, 58+y*14, r, a)
			}
		}
		fmt.Fprintf(&b, `<rect x="0" y="126" width="150" height="10" fill="%s"/>`, bb)
	}
	b.WriteString(`</g>`)

	// spine and grain
	b.WriteString(`<rect width="3.5" height="200" fill="#000" opacity=".18"/>`)
	fmt.Fprintf(&b, `<rect width="150" height="200" filter="url(#kg%d)" opacity=".09" style="mix-blend-mode:multiply"/>`, n)

	// the week, Monday … Sunday, on a paper label in the panel's corner
	wd := int(p.Date.Weekday()+6) % 7 // Monday = 0
	monday := p.Date.AddDate(0, 0, -wd)
	fmt.Fprintf(&b, `<rect x="6" y="124" width="56" height="12" fill="%s"/>`, paper)
	for d := 0; d < 7; d++ {
		w := wordsByDay[monday.AddDate(0, 0, d).Format("2006-01-02")]
		x := 10 + d*7
		switch {
		case w >= 400:
			fmt.Fprintf(&b, `<rect x="%d" y="127" width="5" height="5" rx="1" fill="%s"/>`, x, a)
		case w > 0:
			fmt.Fprintf(&b, `<rect x="%d" y="127" width="5" height="5" rx="1" fill="%s"/>`, x, bb)
		default:
			fmt.Fprintf(&b, `<rect x="%d.5" y="127.5" width="4" height="4" rx="1" fill="none" stroke="%s" stroke-width=".8" opacity=".7"/>`, x, a)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}
