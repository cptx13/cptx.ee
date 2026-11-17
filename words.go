package main

// Addon: words — a heatmap of the most used words, leaving out articles,
// pronouns and other function words (Spanish and English).

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var wordTokenRe = regexp.MustCompile(`\p{L}+`)

// stopWords: articles, pronouns, prepositions, conjunctions, determiners,
// auxiliaries and the most common adverbs, in Spanish and English.
var stopWords = func() map[string]bool {
	list := `
el la los las un una unos unas lo al del
yo tú tu él ella ello nosotros nosotras vosotros vosotras ellos ellas usted ustedes me te se nos os le les mí ti sí conmigo contigo consigo
mi mis tus su sus nuestro nuestra nuestros nuestras vuestro vuestra vuestros vuestras mío mía míos mías tuyo tuya tuyos tuyas suyo suya suyos suyas
este esta estos estas ese esa esos esas aquel aquella aquellos aquellas esto eso aquello
que qué quien quién quienes cual cuál cuales cuáles cuyo cuya cuyos cuyas cuanto cuánto cuanta cuánta cuantos cuántos donde dónde cuando cuándo como cómo
a ante bajo cabe con contra de desde durante en entre hacia hasta mediante para por según sin so sobre tras
y e ni o u pero sino aunque porque pues si mas que
ser soy eres es somos sois son era eras éramos eran fui fue fuimos fueron sido siendo sea sean será serán sería serían
estar estoy estás está estamos están estaba estaban estuve estuvo estado estando esté estén
haber he has ha hemos han había habían hubo habrá habría haya hayan habido hay
muy más menos tan tanto también ya no sí aún todavía así bien sólo solo casi vez hoy uno tal
todo toda todos todas otro otra otros otras mismo misma mismos mismas algo nada alguien nadie algún alguno alguna algunos algunas ningún ninguno ninguna cada
the a an and or but nor so yet for of to in on at by with from into onto upon about over under after before than then as if because while
i me my mine myself you your yours yourself he him his himself she her hers herself it its itself we us our ours ourselves they them their theirs themselves
this that these those who whom whose which what where when why how
is am are was were be been being have has had having do does did done will would shall should can could may might must
not no very too also just only all any some each every other such own same there here
`
	m := map[string]bool{}
	for _, w := range strings.Fields(list) {
		m[w] = true
	}
	return m
}()

// wordCounts counts content words (3+ letters, not stop words) in plain text.
func wordCounts(text string) map[string]int {
	out := map[string]int{}
	for _, w := range wordTokenRe.FindAllString(strings.ToLower(text), -1) {
		if len([]rune(w)) < 3 || stopWords[w] {
			continue
		}
		out[w]++
	}
	return out
}

// buildWordsData returns the template data for /addons/words/.
func buildWordsData(idx []addonPost, q url.Values) map[string]any {
	section := "all"
	switch s := q.Get("section"); s {
	case "croquis", "rimas", "recuerdos":
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
	total := map[string]int{}
	for _, p := range posts {
		if year != 0 && p.Date.Year() != year {
			continue
		}
		for w, n := range p.WordCounts {
			total[w] += n
		}
	}
	words := make([]string, 0, len(total))
	for w := range total {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		return total[words[i]] > total[words[j]] || (total[words[i]] == total[words[j]] && words[i] < words[j])
	})
	if len(words) > 120 {
		words = words[:120]
	}
	selected := ""
	if w := strings.ToLower(q.Get("word")); total[w] > 0 {
		selected = w
	}
	link := func(sec string, yr int, w string) string {
		v := url.Values{}
		if sec != "all" {
			v.Set("section", sec)
		}
		if yr != 0 {
			v.Set("year", strconv.Itoa(yr))
		}
		if w != "" {
			v.Set("word", w)
		}
		if len(v) == 0 {
			return "/addons/words/"
		}
		return "/addons/words/?" + v.Encode()
	}
	max := 1
	if len(words) > 0 {
		max = total[words[0]]
	}
	var tiles []any
	for i, w := range words {
		n := total[w]
		lvl := 1 + 4*(n-1)/max // 1..5 (scaled linearly against the top word)
		if lvl > 5 {
			lvl = 5
		}
		cls := fmt.Sprintf("wd-tile wd-l%d", lvl)
		if w == selected {
			cls += " wd-sel"
		}
		tiles = append(tiles, map[string]any{"Word": w, "Count": n, "Rank": i + 1, "Class": cls,
			"Href": link(section, year, w) + "#word", "Label": fmt.Sprintf("#%d %s · %d", i+1, w, n)})
	}

	// selected word: uses per year + the posts that use it most
	var bars, selPosts []any
	if selected != "" {
		perYear := map[int]int{}
		type hit struct {
			p addonPost
			n int
		}
		var hits []hit
		for _, p := range posts {
			if n := p.WordCounts[selected]; n > 0 {
				perYear[p.Date.Year()] += n
				if year == 0 || p.Date.Year() == year {
					hits = append(hits, hit{p, n})
				}
			}
		}
		maxY := 1
		for _, n := range perYear {
			if n > maxY {
				maxY = n
			}
		}
		for _, y := range years {
			bars = append(bars, map[string]any{"Year": y, "Count": perYear[y], "Width": fmt.Sprintf("width:%d%%", 100*perYear[y]/maxY)})
		}
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].n > hits[j].n })
		if len(hits) > 20 {
			hits = hits[:20]
		}
		var ps []addonPost
		for _, h := range hits {
			ps = append(ps, h.p)
		}
		selPosts = postListMaps(ps)
	}
	opt := func(label string, active bool, href string) any {
		return map[string]any{"Label": label, "Active": active, "Href": href}
	}
	var sectionOpts, yearOpts []any
	for _, s := range []string{"all", "croquis", "rimas", "recuerdos"} {
		sectionOpts = append(sectionOpts, opt(s, s == section, link(s, 0, selected)))
	}
	yearOpts = append(yearOpts, opt("all", year == 0, link(section, 0, selected)))
	for _, y := range years {
		yearOpts = append(yearOpts, opt(strconv.Itoa(y), y == year, link(section, y, selected)))
	}
	yearLabel := "all years"
	if year != 0 {
		yearLabel = strconv.Itoa(year)
	}
	return map[string]any{
		"tiles": tiles, "distinct": len(total), "yearLabel": yearLabel,
		"sections": sectionOpts, "years": yearOpts,
		"selected": selected, "hasSelected": selected != "", "selCount": total[selected],
		"bars": bars, "selPosts": selPosts,
		"shareURL": siteURL + link(section, year, selected),
	}
}
