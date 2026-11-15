package main

// Batch A (B-05, B-06, B-09, B-14, F-01, F-03): feed, sitemap, robots,
// health endpoint, response headers, gzip, language detection, excerpts.

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const siteURL = "https://cptx.ee"

// ---- language (F-01) --------------------------------------------------------

var wordRe = regexp.MustCompile(`[\p{L}']+`)

var esWords = map[string]bool{"de": true, "la": true, "que": true, "el": true, "en": true, "y": true, "los": true, "las": true, "del": true, "por": true, "una": true, "un": true, "con": true, "para": true, "es": true, "lo": true, "se": true, "no": true, "mi": true, "al": true, "como": true, "más": true}
var enWords = map[string]bool{"the": true, "and": true, "of": true, "to": true, "a": true, "in": true, "is": true, "it": true, "that": true, "was": true, "for": true, "my": true, "i": true, "with": true, "on": true, "you": true, "but": true, "not": true, "be": true, "have": true, "as": true, "this": true}

// postLang returns the frontmatter `lang` if set, otherwise "es" or "en" by
// counting common function words in the body. Spanish wins ties (site default).
func postLang(fmLang, body string) string {
	if l := strings.TrimSpace(strings.ToLower(fmLang)); l != "" {
		return l
	}
	es, en := 0, 0
	for _, w := range wordRe.FindAllString(strings.ToLower(body), 2000) {
		if esWords[w] {
			es++
		}
		if enWords[w] {
			en++
		}
	}
	if en > es {
		return "en"
	}
	return "es"
}

// withDoctype makes every rendered page start with <!DOCTYPE html> so browsers
// use standards mode instead of quirks mode (Lighthouse "doctype").
// It also versions local CSS/JS URLs (?v=<commit>) so a deploy always reaches
// browsers that cached the previous files.
func withDoctype(html string) string {
	html = assetVersioner.Replace(html)
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(html)), "<!doctype") {
		return html
	}
	return "<!DOCTYPE html>\n" + html
}

var assetVersion = strings.TrimSuffix(buildRevision(), "-dirty")

var assetVersioner = strings.NewReplacer(
	`href="/css/style.css"`, `href="/css/style.css?v=`+assetVersion+`"`,
	`src="/js/theme-switch.js"`, `src="/js/theme-switch.js?v=`+assetVersion+`"`,
	`src="/js/post-search.js"`, `src="/js/post-search.js?v=`+assetVersion+`"`,
	`src="/js/share.js"`, `src="/js/share.js?v=`+assetVersion+`"`,
	`src="/js/photos.js"`, `src="/js/photos.js?v=`+assetVersion+`"`,
	`src="/js/read.js"`, `src="/js/read.js?v=`+assetVersion+`"`,
)

// contentClass marks verse (tag "rimas") so it gets poem typography (F-09).
func contentClass(p *Post) string {
	if p.Section == "pockets" {
		return "single-content pk-body pk-" + p.Slug
	}
	for _, t := range p.Tags {
		if t == "rimas" {
			return "single-content poem"
		}
	}
	return "single-content"
}

// ---- excerpt for meta description / feed (F-03, B-05) ------------------------

var (
	mdLinkRe  = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	mdNoiseRe = regexp.MustCompile("[*_`#>]+")
	spaceRe   = regexp.MustCompile(`\s+`)
)

// excerpt returns roughly the first n characters of plain text from markdown,
// cut at a word boundary.
func excerpt(markdown string, n int) string {
	s := mdLinkRe.ReplaceAllString(markdown, "$1")
	s = mdNoiseRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)[:n]
	cut := string(r)
	if i := strings.LastIndex(cut, " "); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}

// ---- RSS feed of croquis (B-05) ---------------------------------------------

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

type rssDoc struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"xmlns:atom,attr"`
	Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Language    string `xml:"language"`
		Self        struct {
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"atom:link"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

// buildRSS returns an RSS 2.0 feed of the newest `limit` croquis (full content).
func buildRSS(posts []*Post, limit int) []byte {
	var cs []*Post
	for _, p := range posts {
		if p.Section == "croquis" && !p.Draft {
			cs = append(cs, p)
		}
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].Date.After(cs[j].Date) })
	if len(cs) > limit {
		cs = cs[:limit]
	}
	var d rssDoc
	d.Version, d.Atom = "2.0", "http://www.w3.org/2005/Atom"
	d.Channel.Title = "cptx — croquis"
	d.Channel.Link = siteURL + "/"
	d.Channel.Description = "a collection of mathoms"
	d.Channel.Language = "es"
	d.Channel.Self.Href, d.Channel.Self.Rel, d.Channel.Self.Type = siteURL+"/index.xml", "self", "application/rss+xml"
	for _, p := range cs {
		d.Channel.Items = append(d.Channel.Items, rssItem{
			Title: p.Title, Link: siteURL + p.Path, GUID: siteURL + p.Path,
			PubDate: p.Date.Format(time.RFC1123Z), Description: string(p.Content),
		})
	}
	out, _ := xml.MarshalIndent(d, "", "  ")
	return append([]byte(xml.Header), out...)
}

// ---- sitemap + robots (B-05) ------------------------------------------------

// buildSitemap lists every page in dist/ except redirect stubs (legacy /posts/
// and /categories/pockets/), with lastmod for posts.
func buildSitemap(distDir string, posts []*Post) ([]byte, error) {
	lastmod := map[string]string{}
	for _, p := range posts {
		if !p.Date.IsZero() {
			lastmod[p.Path] = p.Date.Format("2006-01-02")
		}
	}
	var paths []string
	err := filepath.WalkDir(distDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "index.html" {
			return err
		}
		rel, _ := filepath.Rel(distDir, filepath.Dir(path))
		u := "/"
		if rel != "." {
			u = "/" + filepath.ToSlash(rel) + "/"
		}
		if strings.HasPrefix(u, "/posts/") || u == "/categories/pockets/" || retiredURLs[u] != "" {
			return nil
		}
		paths = append(paths, u)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var b bytes.Buffer
	b.WriteString(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, u := range paths {
		b.WriteString("  <url><loc>")
		xml.EscapeText(&b, []byte(siteURL+u))
		b.WriteString("</loc>")
		if lm := lastmod[u]; lm != "" {
			b.WriteString("<lastmod>" + lm + "</lastmod>")
		}
		b.WriteString("</url>\n")
	}
	b.WriteString("</urlset>\n")
	return b.Bytes(), nil
}

// retiredURLs redirect pages that were removed (C-05) so no URL ever 404s.
// They get a redirect stub in dist/ (keeps urls.txt stable) and stay out of the sitemap.
var retiredURLs = map[string]string{
	"/categories/manuscripts/": "/archives/",
	"/categories/archives/":    "/archives/", // L-01: the books moved into /archives/ (written by me)
}

func registerRetired(mux *http.ServeMux, distDir string) error {
	for from, to := range retiredURLs {
		to := to
		mux.HandleFunc(from, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, to, http.StatusMovedPermanently)
		})
		dir := filepath.Join(distDir, filepath.FromSlash(strings.Trim(from, "/")))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		stub := `<!DOCTYPE html><html><head><meta charset="utf-8"><meta http-equiv="refresh" content="0; url=` + to +
			`"><link rel="canonical" href="` + siteURL + to + `"><title>Redirecting...</title></head><body><a href="` + to + `">` + to + `</a></body></html>`
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(stub), 0o644); err != nil {
			return err
		}
	}
	return nil
}

const robotsTxt = `User-agent: *
Disallow: /pockets/cheatsheet/feed/refresh
Disallow: /addons/map/?
Disallow: /newsletter/thanks/
Disallow: /newsletter/welcome/
Disallow: /addons/today/?
Disallow: /addons/places/?
Disallow: /addons/words/?
Disallow: /archives/library/?
Disallow: /archives/library/covers/
Disallow: /search/?

Sitemap: https://cptx.ee/sitemap.xml
`

// ---- /healthz (B-06) --------------------------------------------------------

func buildRevision() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		rev, dirty := "", false
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				rev = s.Value
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
		if len(rev) > 7 {
			rev = rev[:7]
		}
		if dirty {
			rev += "-dirty"
		}
		if rev != "" {
			return rev
		}
	}
	return "unknown"
}

func healthHandler(postCount int, started time.Time) http.HandlerFunc {
	body, _ := json.Marshal(map[string]any{
		"status": "ok", "commit": buildRevision(), "posts": postCount,
		"started": started.UTC().Format(time.RFC3339),
	})
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(body)
	}
}

// ---- headers (B-09) ---------------------------------------------------------

const csp = "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; " +
	"script-src 'self' 'unsafe-inline'; font-src 'self'; connect-src 'self'; object-src 'none'; " +
	"frame-ancestors 'self'; base-uri 'self'; form-action 'self' https://buttondown.com"

func cacheControlFor(path string) string {
	switch {
	case strings.HasPrefix(path, "/fonts/"):
		return "public, max-age=2592000" // 30 days
	case strings.HasPrefix(path, "/css/"), strings.HasPrefix(path, "/js/"):
		return "public, max-age=3600" // 1 hour (files aren't fingerprinted)
	case strings.HasPrefix(path, "/favicon"), path == "/apple-touch-icon.png", path == "/android-chrome-192x192.png":
		return "public, max-age=604800" // 7 days
	case strings.HasPrefix(path, "/photos/") && hasImageExt(path):
		return "public, max-age=604800"
	}
	return ""
}

// cacheControlForRequest: CSS/JS requested with the per-deploy version (?v=<commit>,
// added by assetVersioner) never change at that URL, so browsers may keep them a year.
func cacheControlForRequest(path, version string) string {
	if version != "" && (strings.HasPrefix(path, "/css/") || strings.HasPrefix(path, "/js/")) {
		return "public, max-age=31536000, immutable"
	}
	return cacheControlFor(path)
}

func hasImageExt(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	}
	return false
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", csp)
		if cc := cacheControlForRequest(r.URL.Path, r.URL.Query().Get("v")); cc != "" {
			h.Set("Cache-Control", cc)
		}
		next.ServeHTTP(w, r)
	})
}

// ---- gzip (B-14) ------------------------------------------------------------

var gzPool = sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression); return w }}

func compressible(ct string) bool {
	for _, p := range []string{"text/", "application/json", "application/xml", "application/rss+xml", "image/svg+xml", "application/javascript"} {
		if strings.HasPrefix(ct, p) {
			return true
		}
	}
	return false
}

type gzWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	decided bool
}

func (g *gzWriter) decide(status int) {
	if g.decided {
		return
	}
	g.decided = true
	h := g.Header()
	if status == http.StatusOK && h.Get("Content-Encoding") == "" && h.Get("Content-Range") == "" && compressible(h.Get("Content-Type")) {
		h.Del("Content-Length")
		h.Set("Content-Encoding", "gzip")
		g.gz = gzPool.Get().(*gzip.Writer)
		g.gz.Reset(g.ResponseWriter)
	}
}

func (g *gzWriter) WriteHeader(status int) {
	g.decide(status)
	g.ResponseWriter.WriteHeader(status)
}

func (g *gzWriter) Write(b []byte) (int, error) {
	if !g.decided {
		if g.Header().Get("Content-Type") == "" {
			g.Header().Set("Content-Type", http.DetectContentType(b))
		}
		g.WriteHeader(http.StatusOK)
	}
	if g.gz != nil {
		return g.gz.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if r.Method == http.MethodHead || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		g := &gzWriter{ResponseWriter: w}
		defer func() {
			if g.gz != nil {
				g.gz.Close()
				gzPool.Put(g.gz)
			}
		}()
		next.ServeHTTP(g, r)
	})
}

// pathOnly hands a request to h without its query string, for routers that
// match the whole request URI (dispatch uses r.URL.RequestURI(), so
// "/croquis/?x=1" or Buttondown's "/newsletter/thanks/?email=…" would 404).
// Pages that read a query (search, addons, library) are on the plain mux.
func pathOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery == "" && !r.URL.ForceQuery {
			h.ServeHTTP(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		u := *r.URL
		u.RawQuery, u.ForceQuery = "", false
		r2.URL = &u
		r2.RequestURI = u.RequestURI()
		h.ServeHTTP(w, r2)
	})
}
