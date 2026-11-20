// Reads: /addons/reads/ — articles Carlos kept (data/reads.tsv) and the latest
// posts from the blogs in the cheatsheet (the feed reader in feeds.go) (AD-11).

package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const readsFile = "data/reads.tsv"

// siteReads is the kept list, loaded once at startup.
var siteReads []Read

// Read is one kept article.
type Read struct {
	Kept      string // YYYY-MM-DD or YYYY-MM
	Title     string
	Author    string
	URL       string
	Published string // YYYY-MM-DD, optional
	Note      string
}

// Month returns the kept month as YYYY-MM.
func (r Read) Month() string {
	if len(r.Kept) >= 7 {
		return r.Kept[:7]
	}
	return r.Kept
}

// parseReads reads the tab-separated kept list. Lines starting with # and
// blank lines are skipped; a bad line is an error so a typo never ships.
func parseReads(text string) ([]Read, error) {
	var out []Read
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		for len(f) < 6 {
			f = append(f, "")
		}
		r := Read{Kept: strings.TrimSpace(f[0]), Title: strings.TrimSpace(f[1]), Author: strings.TrimSpace(f[2]),
			URL: strings.TrimSpace(f[3]), Published: strings.TrimSpace(f[4]), Note: strings.TrimSpace(f[5])}
		if _, err := time.Parse("2006-01", r.Month()); err != nil || (len(r.Kept) != 7 && len(r.Kept) != 10) {
			return nil, fmt.Errorf("%s:%d: kept date %q is not YYYY-MM-DD or YYYY-MM", readsFile, n, r.Kept)
		}
		if len(r.Kept) == 10 {
			if _, err := time.Parse("2006-01-02", r.Kept); err != nil {
				return nil, fmt.Errorf("%s:%d: kept date %q: %v", readsFile, n, r.Kept, err)
			}
		}
		if r.Title == "" {
			return nil, fmt.Errorf("%s:%d: no title", readsFile, n)
		}
		if u, err := url.Parse(r.URL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return nil, fmt.Errorf("%s:%d: bad url %q", readsFile, n, r.URL)
		}
		out = append(out, r)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Kept > out[j].Kept })
	return out, nil
}

// loadReads reads data/reads.tsv; a missing file is an empty list.
func loadReads() ([]Read, error) {
	b, err := os.ReadFile(readsFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return parseReads(string(b))
}

// linkKey normalises a link for matching feed items to kept articles.
func linkKey(s string) string {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return s
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	return host + strings.TrimSuffix(u.Path, "/")
}

func readHost(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(u.Host), "www.")
}

func monthLabel(m string) string {
	return strings.Replace(m, "-", ".", 1)
}

// readsSummary is "N articles · YYYY.MM" for the pockets hub card.
func readsSummary(reads []Read) string {
	if len(reads) == 0 {
		return ""
	}
	return fmt.Sprintf("%d articles · %s", len(reads), monthLabel(reads[0].Month()))
}

// buildReadsData is the template data for /addons/reads/. view is "kept"
// (default) or "latest"; cache may be nil when the feed was never fetched.
func buildReadsData(reads []Read, cache *FeedCache, q url.Values) map[string]any {
	view := q.Get("view")
	if view != "latest" {
		view = "kept"
	}
	kept := map[string]bool{}
	for _, r := range reads {
		kept[linkKey(r.URL)] = true
	}

	var months []any
	for i := 0; i < len(reads); {
		m := reads[i].Month()
		var items []any
		for ; i < len(reads) && reads[i].Month() == m; i++ {
			r := reads[i]
			items = append(items, map[string]any{
				"Title": r.Title, "URL": r.URL, "Author": r.Author, "Host": readHost(r.URL),
				"Note": r.Note, "HasAuthor": r.Author != "", "HasNote": r.Note != "",
			})
		}
		months = append(months, map[string]any{"Label": monthLabel(m), "Items": items})
	}

	var latest []any
	var errors []any
	lastRefreshed := ""
	if cache != nil {
		lastRefreshed = cache.LastRefreshed.Format("2006.01.02 15:04 MST")
		for _, it := range cache.Items {
			latest = append(latest, map[string]any{
				"Title": it.Title, "Link": it.Link, "Source": it.Source, "Date": it.DateFormatted,
				"Kept": kept[linkKey(it.Link)],
			})
		}
		for _, e := range cache.Errors {
			errors = append(errors, e)
		}
	}
	sources := make([]any, 0, len(feedSources))
	for _, s := range feedSources {
		sources = append(sources, s.Name)
	}

	return map[string]any{
		"view":          view,
		"isKept":        view == "kept",
		"isLatest":      view == "latest",
		"keptCount":     len(reads),
		"months":        months,
		"hasKept":       len(reads) > 0,
		"latest":        latest,
		"hasLatest":     len(latest) > 0,
		"hasRefreshed":  cache != nil,
		"lastRefreshed": lastRefreshed,
		"errors":        errors,
		"hasErrors":     len(errors) > 0,
		"errorCount":    len(errors),
		"sourceCount":   len(feedSources),
		"sources":       sources,
	}
}

// readsHandlers serves /addons/reads/ (both views) and its rate-limited refresh.
func readsHandlers(mux *http.ServeMux, render func(http.ResponseWriter, string, map[string]any)) {
	mux.HandleFunc("/addons/reads/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/addons/reads/" {
			http.NotFound(w, r)
			return
		}
		cache, err := loadFeedCache()
		if err != nil {
			cache = nil
		}
		render(w, "ReadsPage", buildReadsData(siteReads, cache, r.URL.Query()))
	})
	mux.HandleFunc("/addons/reads/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !refreshFeeds(feedManualGap, "manual") {
			log.Println("Feed refresh skipped: running or refreshed recently")
		}
		http.Redirect(w, r, "/addons/reads/?view=latest", http.StatusSeeOther)
	})
}
