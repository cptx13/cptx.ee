package main

import (
	"image"
	_ "image/gif"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// pocketMeta is the fixed order and labels of the pockets page.
var pocketMeta = []struct {
	Slug string
	Icon string
	Desc string
	Unit string // what the list items are; "" = don't show a count
}{
	{"bedsidetable", "📚", "books I've enjoyed", "books"},
	{"cheatsheet", "🔗", "links at hand", "links"},
	{"workbench", "🔧", "projects & resources", ""},
	{"weekendreads", "📰", "articles I enjoyed", "articles"},
}

// buildPocketsData returns the template data for /pockets/: one entry per
// pocket with its rendered content, an item count and the last-updated month.
func buildPocketsData(pocketPosts []*Post) []any {
	var out []any
	for _, pm := range pocketMeta {
		var contentHTML, updated string
		for _, p := range pocketPosts {
			if p.Slug == pm.Slug {
				contentHTML = string(p.Content)
				if !p.Date.IsZero() {
					updated = p.Date.Format("2006.01")
				}
				break
			}
		}
		meta := ""
		if pm.Unit != "" {
			if n := strings.Count(contentHTML, "<li>"); n > 0 {
				meta = strconv.Itoa(n) + " " + pm.Unit
			}
		}
		if updated != "" {
			if meta != "" {
				meta += " · "
			}
			meta += updated
		}
		out = append(out, map[string]any{
			"Slug":        pm.Slug,
			"Title":       pm.Slug,
			"Icon":        pm.Icon,
			"Desc":        pm.Desc,
			"Meta":        meta,
			"Href":        "/pockets/" + pm.Slug + "/",
			"ContentHTML": contentHTML,
		})
	}
	return out
}

// imageSize reads only the image header; 0, 0 when the format is unknown.
func imageSize(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// folderCover is the thumbnail shown on a folder's card.
func folderCover(f PhotoFolder) string {
	if len(f.Photos) == 0 {
		return ""
	}
	return f.Photos[0].ThumbURL
}

// photoYears is the span of years across the folders, e.g. "2016–2021".
func photoYears(folders []PhotoFolder) string {
	lo, hi := "", ""
	for _, f := range folders {
		if f.Year == "" {
			continue
		}
		if lo == "" || f.Year < lo {
			lo = f.Year
		}
		if hi == "" || f.Year > hi {
			hi = f.Year
		}
	}
	if lo == hi {
		return lo
	}
	return lo + "–" + hi
}

// photoRatio is width/height as a CSS number, for justified rows ("--r").
func photoRatio(p Photo) string {
	if p.W <= 0 || p.H <= 0 {
		return "1.5"
	}
	return fmt.Sprintf("%.3f", float64(p.W)/float64(p.H))
}
