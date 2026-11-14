package main

// Addon: city lights — a pixel map of the world where every city mentioned in
// a post glows. Place names come from data/places.tsv (GeoNames via
// all-the-cities, CC BY 4.0, plus Spanish exonyms); words that are also place
// names somewhere live in data/places-exclude.txt. Land masks are rasterised
// from Natural Earth.

import (
	_ "embed"
	"fmt"
	"html"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed data/places.tsv
var placesTSV string

//go:embed data/places-exclude.txt
var placesExcludeTxt string

//go:embed data/land-world.txt
var landWorldTxt string

//go:embed data/land-europe.txt
var landEuropeTxt string

type place struct {
	Name, Country string
	Lat, Lon      float64
}

type region struct {
	Key, Label             string
	Lon0, Lon1, Lat0, Lat1 float64 // west, east, south, north
	Cell                   float64 // SVG units per mask cell
	Mask                   []string
	landPath               string
}

var (
	gazetteer   map[string]place // surface form → place
	placeRegion = map[string]*region{}
	capRunRe    = regexp.MustCompile(`\p{Lu}[\p{L}'’-]+(?: (?:de |del |la |el |am )?\p{Lu}[\p{L}'’-]+){0,3}`)
)

func init() {
	excl := map[string]bool{}
	for _, l := range strings.Split(placesExcludeTxt, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			excl[l] = true
		}
	}
	gazetteer = map[string]place{}
	for _, l := range strings.Split(placesTSV, "\n") {
		f := strings.Split(l, "\t")
		if len(f) != 5 || strings.HasPrefix(l, "#") || excl[f[0]] {
			continue
		}
		lat, _ := strconv.ParseFloat(f[3], 64)
		lon, _ := strconv.ParseFloat(f[4], 64)
		gazetteer[f[0]] = place{Name: f[1], Country: f[2], Lat: lat, Lon: lon}
	}
	for _, r := range []*region{
		{Key: "world", Label: "world", Lon0: -180, Lon1: 180, Lat0: -58, Lat1: 84, Cell: 6, Mask: maskLines(landWorldTxt)},
		{Key: "europe", Label: "europe", Lon0: -12, Lon1: 34, Lat0: 34, Lat1: 64, Cell: 10, Mask: maskLines(landEuropeTxt)},
	} {
		var b strings.Builder
		d := r.Cell * 0.62
		for y, row := range r.Mask {
			for x, c := range row {
				if c == '#' {
					fmt.Fprintf(&b, "M%.1f %.1fh%.1fv%.1fh-%.1fz", float64(x)*r.Cell+(r.Cell-d)/2, float64(y)*r.Cell+(r.Cell-d)/2, d, d, d)
				}
			}
		}
		r.landPath = b.String()
		placeRegion[r.Key] = r
	}
	canonicalPlaces = placeByName()
}

func maskLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

// placeMentions counts the places named in a post's text (capitalised runs,
// longest match first).
func placeMentions(text string) map[string]int {
	out := map[string]int{}
	for _, run := range capRunRe.FindAllString(text, -1) {
		parts := strings.Split(run, " ")
		for i := 0; i < len(parts); {
			matched := 0
			for j := len(parts); j > i; j-- {
				if p, ok := gazetteer[strings.Join(parts[i:j], " ")]; ok {
					out[p.Name]++
					matched = j - i
					break
				}
			}
			if matched == 0 {
				matched = 1
			}
			i += matched
		}
	}
	return out
}

// placeByName maps each canonical name to its coordinates.
func placeByName() map[string]place {
	m := map[string]place{}
	for _, p := range gazetteer {
		m[p.Name] = p
	}
	return m
}

var canonicalPlaces map[string]place

// buildPlacesData returns the template data for /addons/places/.
func buildPlacesData(idx []addonPost, q url.Values) map[string]any {
	section := "all"
	switch s := q.Get("section"); s {
	case "croquis", "recuerdos":
		section = s
	}
	posts := filterPosts(idx, section, "")
	years := yearsOf(posts)
	year := 0
	if y, err := strconv.Atoi(q.Get("year")); err == nil {
		for _, yy := range years {
			if yy == y {
				year = y
			}
		}
	}
	reg := placeRegion["world"]
	if q.Get("region") == "europe" {
		reg = placeRegion["europe"]
	}
	type agg struct {
		mentions int
		posts    []addonPost
	}
	counts := map[string]*agg{}
	for _, p := range posts {
		if year != 0 && p.Date.Year() != year {
			continue
		}
		for name, n := range p.Places {
			a := counts[name]
			if a == nil {
				a = &agg{}
				counts[name] = a
			}
			a.mentions += n
			a.posts = append(a.posts, p)
		}
	}
	selected := ""
	if _, ok := counts[q.Get("place")]; ok {
		selected = q.Get("place")
	}
	link := func(reg2, sec string, yr int, pl string) string {
		v := url.Values{}
		if reg2 != "world" {
			v.Set("region", reg2)
		}
		if sec != "all" {
			v.Set("section", sec)
		}
		if yr != 0 {
			v.Set("year", strconv.Itoa(yr))
		}
		if pl != "" {
			v.Set("place", pl)
		}
		if len(v) == 0 {
			return "/addons/places/"
		}
		return "/addons/places/?" + v.Encode()
	}

	// SVG: land dots + one light per place inside the region.
	w := float64(len(reg.Mask[0])) * reg.Cell
	h := float64(len(reg.Mask)) * reg.Cell
	maxM := 1
	for _, a := range counts {
		if a.mentions > maxM {
			maxM = a.mentions
		}
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { // big lights first, so small ones stay clickable on top
		return counts[names[i]].mentions > counts[names[j]].mentions || (counts[names[i]].mentions == counts[names[j]].mentions && names[i] < names[j])
	})
	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg class="pl-map" viewBox="0 0 %.0f %.0f" role="img" aria-label="Map of places mentioned (%s)"><path class="pl-land" d="%s"/>`, w, h, reg.Label, reg.landPath)
	inView := 0
	for _, n := range names {
		p := canonicalPlaces[n]
		if p.Lon < reg.Lon0 || p.Lon > reg.Lon1 || p.Lat < reg.Lat0 || p.Lat > reg.Lat1 {
			continue
		}
		inView++
		a := counts[n]
		x := (p.Lon - reg.Lon0) / (reg.Lon1 - reg.Lon0) * w
		y := (reg.Lat1 - p.Lat) / (reg.Lat1 - reg.Lat0) * h
		k := math.Log1p(float64(a.mentions)) / math.Log1p(float64(maxM)) // 0..1
		r := reg.Cell * (0.35 + 0.9*k)
		cls := "pl-light"
		if n == selected {
			cls += " pl-sel"
		}
		label := html.EscapeString(fmt.Sprintf("%s · %s", n, plural(a.mentions, "mention")))
		fmt.Fprintf(&svg, `<a href="%s#place" aria-label="%s"><title>%s</title><circle class="pl-glow" cx="%.1f" cy="%.1f" r="%.1f" style="opacity:%.2f"/><circle class="%s" cx="%.1f" cy="%.1f" r="%.1f"/></a>`,
			html.EscapeString(link(reg.Key, section, year, n)), label, label, x, y, r*2.6, 0.18+0.35*k, cls, x, y, r)
	}
	svg.WriteString(`</svg>`)

	var rows []any
	for i, n := range names {
		if i >= 60 {
			break
		}
		a := counts[n]
		rows = append(rows, map[string]any{"Name": n, "Country": canonicalPlaces[n].Country, "Mentions": a.mentions,
			"Posts": len(a.posts), "Href": link(reg.Key, section, year, n) + "#place", "Selected": n == selected})
	}
	opt := func(label string, active bool, href string) any {
		return map[string]any{"Label": label, "Active": active, "Href": href}
	}
	var regionOpts, sectionOpts, yearOpts []any
	for _, r := range []string{"world", "europe"} {
		regionOpts = append(regionOpts, opt(r, r == reg.Key, link(r, section, year, selected)))
	}
	for _, s := range []string{"all", "croquis", "recuerdos"} {
		sectionOpts = append(sectionOpts, opt(s, s == section, link(reg.Key, s, 0, "")))
	}
	yearOpts = append(yearOpts, opt("all", year == 0, link(reg.Key, section, 0, selected)))
	for _, y := range years {
		yearOpts = append(yearOpts, opt(strconv.Itoa(y), y == year, link(reg.Key, section, y, "")))
	}
	var selPosts []any
	selMore, selPostCount := 0, 0
	if a := counts[selected]; a != nil {
		ps := a.posts
		selPostCount = len(ps)
		if len(ps) > 30 { // newest 30; the rest is summarised
			selMore = len(ps) - 30
			ps = ps[:30]
		}
		selPosts = postListMaps(ps)
	}
	total := 0
	for _, a := range counts {
		total += a.mentions
	}
	yearLabel := "all years"
	if year != 0 {
		yearLabel = strconv.Itoa(year)
	}
	return map[string]any{
		"svg": svg.String(), "placeCount": len(counts), "inView": inView, "mentions": total, "yearLabel": yearLabel,
		"regions": regionOpts, "sections": sectionOpts, "years": yearOpts, "rows": rows,
		"selected": selected, "selPosts": selPosts, "hasSelected": selected != "",
		"selMore": selMore, "hasMore": selMore > 0, "selPostCount": selPostCount,
		"shareURL": siteURL + link(reg.Key, section, year, selected),
	}
}
