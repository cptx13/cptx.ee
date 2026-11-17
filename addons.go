package main

// F-15: /addons/ — archive heatmap (/addons/map/, shareable URL params) and
// "on this day" (/addons/today/ + a block on the home page). Everything is
// rendered on the server, so it works without JavaScript.

import (
	"fmt"
	"html"
	"os"
	"path/filepath"

	"github.com/dhamidi/htmlc"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Europe/Helsinki even if the host has no tz database
)

var siteTZ = func() *time.Location {
	if l, err := time.LoadLocation("Europe/Helsinki"); err == nil {
		return l
	}
	return time.UTC
}()

type addonPost struct {
	Title, Path, Section, Day string // Day = YYYY-MM-DD
	Tags                      []string
	Words                     int
	Date                      time.Time
	Places                    map[string]int // place name → mentions
	WordCounts                map[string]int // content word → uses
}

var tagRe = regexp.MustCompile(`<[^>]+>`)

func plainText(h string) string { return html.UnescapeString(tagRe.ReplaceAllString(h, " ")) }

// buildAddonIndex keeps published croquis, rimas and recuerdos, newest first.
// Poems count as their own section, "rimas", never as croquis.
func buildAddonIndex(posts []*Post) []addonPost {
	var out []addonPost
	for _, p := range posts {
		if p.Draft || p.Date.IsZero() || (p.Section != "croquis" && p.Section != "recuerdos") {
			continue
		}
		text := plainText(string(p.Content))
		section := p.Section
		if isRima(p) {
			section = "rimas"
		}
		out = append(out, addonPost{
			Title: p.Title, Path: p.Path, Section: section, Tags: p.Tags,
			Words: len(strings.Fields(text)), Date: p.Date, Day: p.Date.Format("2006-01-02"),
			Places: placeMentions(text), WordCounts: wordCounts(text),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out
}

func postListMaps(ps []addonPost) []any {
	out := make([]any, 0, len(ps))
	for _, p := range ps {
		title := p.Title
		if strings.TrimSpace(strings.TrimRight(title, " -")) == "" {
			title = p.Section + " " + p.Day
		}
		out = append(out, map[string]any{"Title": title, "Path": p.Path, "Section": p.Section,
			"Year": p.Date.Format("2006"), "Day": p.Day, "Words": p.Words})
	}
	return out
}

// ---- map params ---------------------------------------------------------------

type mapParams struct {
	Year         int    // 0 = all years
	Section, Tag string // section: all|croquis|rimas|recuerdos; tag: "" or a tag
	Metric, Day  string // metric: posts|words; day: YYYY-MM-DD or ""
}

func (m mapParams) url(change func(*mapParams)) string {
	c := m
	if change != nil {
		change(&c)
	}
	q := url.Values{}
	if c.Year != 0 {
		q.Set("year", strconv.Itoa(c.Year))
	}
	if c.Section != "all" {
		q.Set("section", c.Section)
	}
	if c.Tag != "" {
		q.Set("tag", c.Tag)
	}
	if c.Metric != "posts" {
		q.Set("metric", c.Metric)
	}
	if c.Day != "" {
		q.Set("day", c.Day)
	}
	if len(q) == 0 {
		return "/addons/map/"
	}
	return "/addons/map/?" + q.Encode()
}

func filterPosts(idx []addonPost, section, tag string) []addonPost {
	var out []addonPost
	for _, p := range idx {
		if section != "all" && p.Section != section {
			continue
		}
		if tag != "" {
			ok := false
			for _, t := range p.Tags {
				ok = ok || t == tag
			}
			if !ok {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

// yearsOf lists the years that have posts, ascending.
func yearsOf(ps []addonPost) []int {
	set := map[int]bool{}
	for _, p := range ps {
		set[p.Date.Year()] = true
	}
	var ys []int
	for y := range set {
		ys = append(ys, y)
	}
	sort.Ints(ys)
	return ys
}

// parseMapParams whitelists every parameter; anything unknown falls back to a
// default. The default year is "all" (0).
func parseMapParams(q url.Values, idx []addonPost, tags map[string]bool) (mapParams, []int) {
	m := mapParams{Section: "all", Metric: "posts"}
	switch s := q.Get("section"); s {
	case "croquis", "rimas", "recuerdos":
		m.Section = s
	}
	if q.Get("tag") == "rimas" { // old links from before rimas became a section
		m.Section = "rimas"
	}
	if t := q.Get("tag"); tags[t] {
		m.Tag = t
	}
	if q.Get("metric") == "words" {
		m.Metric = "words"
	}
	years := yearsOf(filterPosts(idx, m.Section, m.Tag))
	if y, err := strconv.Atoi(q.Get("year")); err == nil {
		for _, yy := range years {
			if yy == y {
				m.Year = y
			}
		}
	}
	if d, err := time.Parse("2006-01-02", q.Get("day")); err == nil && (m.Year == 0 || d.Year() == m.Year) {
		m.Day = d.Format("2006-01-02")
	}
	return m, years
}

type dayAgg struct {
	n, words int
	posts    []addonPost
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

// buildMapData returns the template data for /addons/map/.
func buildMapData(idx []addonPost, tags []string, q url.Values) map[string]any {
	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[t] = true
	}
	m, years := parseMapParams(q, idx, tagSet)
	posts := filterPosts(idx, m.Section, m.Tag)

	byDay := map[string]*dayAgg{}
	total, words := 0, 0
	for _, p := range posts {
		if m.Year != 0 && p.Date.Year() != m.Year {
			continue
		}
		a := byDay[p.Day]
		if a == nil {
			a = &dayAgg{}
			byDay[p.Day] = a
		}
		a.n++
		a.words += p.Words
		a.posts = append(a.posts, p)
		total++
		words += p.Words
	}
	val := func(a *dayAgg) int {
		if a == nil {
			return 0
		}
		if m.Metric == "words" {
			return a.words
		}
		return a.n
	}
	max := 1
	for _, a := range byDay {
		if v := val(a); v > max {
			max = v
		}
	}

	// longest run of consecutive days with posts
	var days []string
	for d := range byDay {
		days = append(days, d)
	}
	sort.Strings(days)
	best, run := 0, 0
	var prev time.Time
	for i, d := range days {
		t, _ := time.Parse("2006-01-02", d)
		if i > 0 && t.Sub(prev) == 24*time.Hour {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
		prev = t
	}

	showYears := years
	if m.Year != 0 {
		showYears = []int{m.Year}
	}
	var grids []any
	for i := len(showYears) - 1; i >= 0; i-- { // newest first
		y := showYears[i]
		grids = append(grids, map[string]any{"Year": y, "Weeks": yearWeeks(m, y, byDay, val, max)})
	}

	opt := func(label string, active bool, href string) any {
		return map[string]any{"Label": label, "Active": active, "Href": href}
	}
	yearOpts := []any{opt("all", m.Year == 0, m.url(func(c *mapParams) { c.Year = 0; c.Day = "" }))}
	for _, y := range years {
		yy := y
		yearOpts = append(yearOpts, opt(strconv.Itoa(y), y == m.Year, m.url(func(c *mapParams) { c.Year = yy; c.Day = "" })))
	}
	var sectionOpts, tagOpts, metricOpts []any
	for _, s := range []string{"all", "croquis", "rimas", "recuerdos"} {
		ss := s
		sectionOpts = append(sectionOpts, opt(s, s == m.Section, m.url(func(c *mapParams) { c.Section = ss; c.Day = "" })))
	}
	tagOpts = append(tagOpts, opt("any", m.Tag == "", m.url(func(c *mapParams) { c.Tag = ""; c.Day = "" })))
	for _, t := range tags {
		tt := t
		tagOpts = append(tagOpts, opt("#"+t, t == m.Tag, m.url(func(c *mapParams) { c.Tag = tt; c.Day = "" })))
	}
	for _, s := range []string{"posts", "words"} {
		ss := s
		metricOpts = append(metricOpts, opt(s, s == m.Metric, m.url(func(c *mapParams) { c.Metric = ss })))
	}
	var dayPosts []any
	if a := byDay[m.Day]; a != nil {
		dayPosts = postListMaps(a.posts)
	}
	title := "all years"
	yearLabel := "all"
	if m.Year != 0 {
		title = strconv.Itoa(m.Year)
		yearLabel = title
	}
	return map[string]any{
		"title": title, "yearLabel": yearLabel, "total": total, "words": words,
		"activeDays": len(byDay), "streakLabel": plural(best, "day"),
		"metric": m.Metric, "grids": grids,
		"years": yearOpts, "sections": sectionOpts, "tags": tagOpts, "metrics": metricOpts,
		"day": m.Day, "dayPosts": dayPosts, "hasDayPosts": len(dayPosts) > 0,
		"shareURL": siteURL + m.url(nil),
	}
}

// yearWeeks builds the Mon–Sun week columns of one year's calendar.
func yearWeeks(m mapParams, year int, byDay map[string]*dayAgg, val func(*dayAgg) int, max int) []any {
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	start := jan1.AddDate(0, 0, -((int(jan1.Weekday()) + 6) % 7))
	var weeks []any
	for d := start; d.Year() <= year; {
		var cells []any
		label := ""
		for i := 0; i < 7; i++ {
			day := d.Format("2006-01-02")
			a := byDay[day]
			v := val(a)
			cell := map[string]any{"Has": false, "Class": "hm-cell hm-l0", "Href": "", "Label": day}
			switch {
			case d.Year() != year:
				cell["Class"] = "hm-cell hm-out"
				cell["Label"] = ""
			case v > 0:
				lvl := (4*v + max - 1) / max // 1..4
				cls := fmt.Sprintf("hm-cell hm-l%d", lvl)
				if day == m.Day {
					cls += " hm-sel"
				}
				dd := day
				cell["Has"] = true
				cell["Class"] = cls
				cell["Href"] = m.url(func(c *mapParams) { c.Day = dd }) + "#day"
				cell["Label"] = fmt.Sprintf("%s · %s · %d words", day, plural(a.n, "post"), a.words)
			}
			if d.Year() == year && d.Day() == 1 {
				label = d.Format("Jan")
			}
			cells = append(cells, cell)
			d = d.AddDate(0, 0, 1)
		}
		weeks = append(weeks, map[string]any{"Days": cells, "Label": label})
	}
	return weeks
}

// ---- on this day --------------------------------------------------------------

// onThisDay returns posts from earlier years with the same month and day as md ("MM-DD").
func onThisDay(idx []addonPost, md string, before time.Time) []addonPost {
	var out []addonPost
	for _, p := range idx {
		if p.Date.Format("01-02") == md && p.Date.Before(before) {
			out = append(out, p)
		}
	}
	return out
}

func buildTodayData(idx []addonPost, q url.Values, now time.Time) map[string]any {
	today := now.In(siteTZ)
	d := time.Date(2024, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC) // 2024: leap year, so 02-29 is valid
	if t, err := time.Parse("2006-01-02", "2024-"+q.Get("date")); err == nil {
		d = t
	}
	md := d.Format("01-02")
	posts := onThisDay(idx, md, time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1))
	return map[string]any{
		"dateLabel": d.Format("2 January"), "md": md,
		"prevHref": "/addons/today/?date=" + d.AddDate(0, 0, -1).Format("01-02"),
		"nextHref": "/addons/today/?date=" + d.AddDate(0, 0, 1).Format("01-02"),
		"posts":    postListMaps(posts), "hasPosts": len(posts) > 0, "count": len(posts),
	}
}

// homeOnThisDay is the small block on the home page (up to 5 posts).
func homeOnThisDay(idx []addonPost, now time.Time) ([]any, bool) {
	t := now.In(siteTZ)
	ps := onThisDay(idx, t.Format("01-02"), time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
	if len(ps) > 5 {
		ps = ps[:5]
	}
	return postListMaps(ps), len(ps) > 0
}

// registerAddons wires the /addons/ pages into the mux.
func registerAddons(mux *http.ServeMux, render func(http.ResponseWriter, string, map[string]any), idx []addonPost, tags []string) {
	mux.HandleFunc("/addons/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/" {
			http.NotFound(w, r)
			return
		}
		render(w, "AddonsPage", map[string]any{"count": len(idx)})
	})
	mux.HandleFunc("/addons/map/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/map/" {
			http.NotFound(w, r)
			return
		}
		render(w, "MapPage", buildMapData(idx, tags, r.URL.Query()))
	})
	mux.HandleFunc("/addons/places/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/places/" {
			http.NotFound(w, r)
			return
		}
		render(w, "PlacesPage", buildPlacesData(idx, r.URL.Query()))
	})
	mux.HandleFunc("/addons/words/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/words/" {
			http.NotFound(w, r)
			return
		}
		render(w, "WordsPage", buildWordsData(idx, r.URL.Query()))
	})
	mux.HandleFunc("/addons/readers/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/readers/" {
			http.NotFound(w, r)
			return
		}
		render(w, "ReadersPage", buildReadersData(siteStats, idx, time.Now()))
	})
	mux.HandleFunc("/addons/today/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/today/" {
			http.NotFound(w, r)
			return
		}
		render(w, "TodayPage", buildTodayData(idx, r.URL.Query(), time.Now()))
	})
}

// addonTags lists tags used by croquis/recuerdos, sorted. "rimas" is a section
// now, so it is not offered as a tag.
func addonTags(idx []addonPost) []string {
	set := map[string]bool{}
	for _, p := range idx {
		for _, t := range p.Tags {
			if t != "" && t != "rimas" {
				set[t] = true
			}
		}
	}
	var out []string
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// writeAddonPages writes the default views into dist/ so urls.txt, the sitemap
// and check.sh cover them.
func writeAddonPages(engine *htmlc.Engine, distDir string, idx []addonPost, tags []string) error {
	pages := map[string]struct {
		comp string
		data map[string]any
	}{
		"addons":         {"AddonsPage", map[string]any{"count": len(idx)}},
		"addons/map":     {"MapPage", buildMapData(idx, tags, url.Values{})},
		"addons/today":   {"TodayPage", buildTodayData(idx, url.Values{}, time.Now())},
		"addons/places":  {"PlacesPage", buildPlacesData(idx, url.Values{})},
		"addons/words":   {"WordsPage", buildWordsData(idx, url.Values{})},
		"addons/readers": {"ReadersPage", buildReadersData(siteStats, idx, time.Now())},
	}
	for dir, pg := range pages {
		html, err := engine.RenderPageString(pg.comp, pg.data)
		if err != nil {
			return fmt.Errorf("%s: %w", pg.comp, err)
		}
		out := filepath.Join(distDir, filepath.FromSlash(dir))
		if err := os.MkdirAll(out, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, "index.html"), []byte(withDoctype(html)), 0o644); err != nil {
			return err
		}
	}
	return nil
}
