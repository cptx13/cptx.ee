// Readers' counters (A-01…A-04): visits, reads and countries, counted on this
// server without cookies or third parties. Only daily totals are stored, in
// <data>/stats/YYYY-MM.json. A visitor is recognised for one day only, by a
// hash of IP + browser + a random salt that lives in memory and is replaced at
// midnight; neither the IP nor the hash is ever written down.

package main

import (
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
)

//go:embed data/countries.tsv
var countriesTSV string

// statsDirs: STATS_DIR (or "off"), else ~/cptx-data/stats when ~/cptx-data exists.
// The country file lives next to it in geo/. Both "" = count in memory only.
func statsDirs() (statsDir, geoFile string) {
	if d := os.Getenv("STATS_DIR"); d != "" {
		if d == "off" {
			return "", ""
		}
		return d, filepath.Join(filepath.Dir(d), "geo", "dbip-country-lite.csv.gz")
	}
	if h, err := os.UserHomeDir(); err == nil {
		base := filepath.Join(h, "cptx-data")
		if st, err := os.Stat(base); err == nil && st.IsDir() {
			return filepath.Join(base, "stats"), filepath.Join(base, "geo", "dbip-country-lite.csv.gz")
		}
	}
	return "", ""
}

// dayStats is one day of totals, as stored on disk.
type dayStats struct {
	Visitors  int            `json:"visitors"`  // unique visitors that day
	Visits    map[string]int `json:"visits"`    // page path → page views
	Reads     map[string]int `json:"reads"`     // post path → reads (once per visitor per day)
	Countries map[string]int `json:"countries"` // ISO code → unique visitors ("ZZ" = unknown)
}

func newDay() *dayStats {
	return &dayStats{Visits: map[string]int{}, Reads: map[string]int{}, Countries: map[string]int{}}
}

type statsStore struct {
	mu       sync.Mutex
	dir      string // "" = keep in memory only (CI, previews)
	loc      *time.Location
	days     map[string]*dayStats // "2006-01-02" → totals
	dirty    map[string]bool      // months to save ("2006-01")
	today    string
	salt     [16]byte
	seen     map[[16]byte]bool // today's visitors
	seenRead map[[16]byte]bool // today's visitor+post reads
	reads    map[string]int    // post path → reads, all time
	readable map[string]bool   // posts that can be "read"
	geo      *geoKeeper
}

var siteStats *statsStore

// newStatsStore loads saved months from dir (if any).
func newStatsStore(dir string, readable map[string]bool, geo *geoKeeper) *statsStore {
	loc, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		loc = time.UTC
	}
	s := &statsStore{dir: dir, loc: loc, days: map[string]*dayStats{}, dirty: map[string]bool{},
		seen: map[[16]byte]bool{}, seenRead: map[[16]byte]bool{}, reads: map[string]int{}, readable: readable, geo: geo}
	if dir != "" {
		files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			m := map[string]*dayStats{}
			if err := json.Unmarshal(b, &m); err != nil {
				log.Printf("stats: %s: %v", f, err)
				continue
			}
			for d, ds := range m {
				if ds.Visits == nil {
					ds.Visits = map[string]int{}
				}
				if ds.Reads == nil {
					ds.Reads = map[string]int{}
				}
				if ds.Countries == nil {
					ds.Countries = map[string]int{}
				}
				s.days[d] = ds
				for p, n := range ds.Reads {
					s.reads[p] += n
				}
			}
		}
	}
	return s
}

// rollDay starts a new day (new salt, forget who was seen). Caller holds mu.
func (s *statsStore) rollDay(now time.Time) *dayStats {
	day := now.In(s.loc).Format("2006-01-02")
	if day != s.today {
		s.today = day
		rand.Read(s.salt[:])
		s.seen = map[[16]byte]bool{}
		s.seenRead = map[[16]byte]bool{}
	}
	ds := s.days[day]
	if ds == nil {
		ds = newDay()
		s.days[day] = ds
	}
	s.dirty[day[:7]] = true
	return ds
}

func (s *statsStore) visitor(ip, ua string, extra string) [16]byte {
	h := sha256.New()
	h.Write(s.salt[:])
	io.WriteString(h, ip+"\x00"+ua+"\x00"+extra)
	var out [16]byte
	copy(out[:], h.Sum(nil))
	return out
}

// countView records one page view.
func (s *statsStore) countView(path, ip, ua string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ds := s.rollDay(now)
	ds.Visits[path]++
	v := s.visitor(ip, ua, "")
	if !s.seen[v] {
		s.seen[v] = true
		ds.Visitors++
		cc := ""
		if s.geo != nil {
			if a, err := netip.ParseAddr(ip); err == nil {
				cc = s.geo.lookup(a)
			}
		}
		if cc == "" {
			cc = "ZZ"
		}
		ds.Countries[cc]++
	}
}

// countRead records a read; false when it doesn't count (unknown post, repeat).
func (s *statsStore) countRead(path, ip, ua string, now time.Time) bool {
	if !s.readable[path] {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ds := s.rollDay(now)
	k := s.visitor(ip, ua, path)
	if s.seenRead[k] {
		return false
	}
	s.seenRead[k] = true
	ds.Reads[path]++
	s.reads[path]++
	return true
}

// readsOf is the all-time read count of a post.
func (s *statsStore) readsOf(path string) int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reads[path]
}

// save writes the months that changed. Safe to call any time.
func (s *statsStore) save() {
	if s == nil || s.dir == "" {
		return
	}
	s.mu.Lock()
	months := map[string]map[string]*dayStats{}
	for m := range s.dirty {
		months[m] = map[string]*dayStats{}
	}
	for d, ds := range s.days {
		if mm, ok := months[d[:7]]; ok {
			cp := *ds // marshalled below, still under the lock
			mm[d] = &cp
		}
	}
	var blobs = map[string][]byte{}
	for m, mm := range months {
		b, err := json.MarshalIndent(mm, "", " ")
		if err == nil {
			blobs[m] = b
		}
	}
	s.dirty = map[string]bool{}
	s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		log.Printf("stats: %v", err)
		return
	}
	for m, b := range blobs {
		f := filepath.Join(s.dir, m+".json")
		if err := os.WriteFile(f+".tmp", b, 0o644); err == nil {
			os.Rename(f+".tmp", f)
		} else {
			log.Printf("stats: %v", err)
		}
	}
}

func (s *statsStore) saveLoop() {
	for {
		time.Sleep(5 * time.Minute)
		s.save()
	}
}

// ---- who counts --------------------------------------------------------------

var botRe = regexp.MustCompile(`(?i)bot|crawl|spider|slurp|curl|wget|python|go-http|java/|httpclient|headless|lighthouse|monitor|uptime|preview|facebookexternalhit|embedly|feed|rss|validator|scrapy|okhttp|axios|node-fetch|libwww|bingpreview|petal|semrush|ahrefs|archive|check`)

func isBot(r *http.Request) bool {
	ua := r.UserAgent()
	if ua == "" || botRe.MatchString(ua) {
		return true
	}
	p := r.Header.Get("Sec-Purpose") + r.Header.Get("Purpose") // link prefetches
	return strings.Contains(p, "prefetch")
}

// clientIP: the exe.dev proxy appends the client's address as the last
// X-Forwarded-For entry; fall back to the connection's address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// withStats counts successful HTML page views by people (not bots).
func withStats(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if siteStats == nil || r.Method != http.MethodGet || isBot(r) {
			next.ServeHTTP(w, r)
			return
		}
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == http.StatusOK && strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
			siteStats.countView(r.URL.Path, clientIP(r), r.UserAgent(), time.Now())
		}
	})
}

// readHandler takes the beacon from /js/read.js: the body is the post's path.
func readHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if siteStats != nil && !isBot(r) {
		b, _ := io.ReadAll(io.LimitReader(r.Body, 300))
		siteStats.countRead(strings.TrimSpace(string(b)), clientIP(r), r.UserAgent(), time.Now())
	}
	w.WriteHeader(http.StatusNoContent)
}

// pageLabel names a page that isn't a post: "/" → "home", "/addons/readers/" → "addons / readers".
func pageLabel(path string) string {
	p := strings.Trim(path, "/")
	if p == "" {
		return "home"
	}
	p = strings.TrimPrefix(p, "categories/")
	return strings.ReplaceAll(p, "/", " / ")
}

// readsLabel is the line under a post ("read 12 times"); "" until it has reads.
func readsLabel(n int) string {
	switch {
	case n <= 0:
		return ""
	case n == 1:
		return "read once"
	case n == 2:
		return "read twice"
	}
	return fmt.Sprintf("read %d times", n)
}

// ---- /addons/readers/ --------------------------------------------------------

type countryInfo struct {
	Name     string
	Lat, Lon float64
	HasPos   bool
}

var countryTable = func() map[string]countryInfo {
	m := map[string]countryInfo{}
	for _, l := range strings.Split(countriesTSV, "\n") {
		f := strings.Split(l, "\t")
		if len(f) != 4 || strings.HasPrefix(l, "#") {
			continue
		}
		ci := countryInfo{Name: f[1]}
		if f[2] != "" {
			fmt.Sscanf(f[2], "%g", &ci.Lat)
			fmt.Sscanf(f[3], "%g", &ci.Lon)
			ci.HasPos = true
		}
		m[f[0]] = ci
	}
	return m
}()

func countryName(cc string) string {
	if c, ok := countryTable[cc]; ok {
		return c.Name
	}
	if cc == "ZZ" {
		return "unknown"
	}
	return cc
}

// readerExtras are readable pages outside the addon index (newsletter issues),
// set at startup so /addons/readers/ can name them.
var readerExtras []addonPost

// buildReadersData returns the template data for /addons/readers/.
func buildReadersData(s *statsStore, idx []addonPost, now time.Time) map[string]any {
	titles := map[string]addonPost{}
	for _, p := range idx {
		titles[p.Path] = p
	}
	for _, p := range readerExtras {
		titles[p.Path] = p
	}
	visits, visitors, reads := 0, 0, 0
	byMonth := map[string][3]int{}
	pageVisits := map[string]int{}
	postReads := map[string]int{}
	countries := map[string]int{}
	first := ""
	if s != nil {
		s.mu.Lock()
		for d, ds := range s.days {
			if first == "" || d < first {
				first = d
			}
			m := byMonth[d[:7]]
			for p, n := range ds.Visits {
				visits += n
				m[0] += n
				pageVisits[p] += n
			}
			visitors += ds.Visitors
			m[1] += ds.Visitors
			for p, n := range ds.Reads {
				reads += n
				m[2] += n
				postReads[p] += n
			}
			for cc, n := range ds.Countries {
				countries[cc] += n
			}
			byMonth[d[:7]] = m
		}
		s.mu.Unlock()
	}

	// last 12 months, oldest first
	var months []any
	maxV := 1
	for i := 11; i >= 0; i-- {
		if v := byMonth[now.AddDate(0, -i, 0).Format("2006-01")][0]; v > maxV {
			maxV = v
		}
	}
	for i := 11; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		m := byMonth[t.Format("2006-01")]
		months = append(months, map[string]any{
			"Label": t.Format("Jan 06"), "Visits": m[0], "Visitors": m[1], "Reads": m[2],
			"Width": fmt.Sprintf("width:%.1f%%", 100*float64(m[0])/float64(maxV)),
		})
	}

	type kv struct {
		k string
		n int
	}
	top := func(m map[string]int, keep func(string) bool, limit int) []kv {
		var out []kv
		for k, n := range m {
			if keep == nil || keep(k) {
				out = append(out, kv{k, n})
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].n > out[j].n || (out[i].n == out[j].n && out[i].k < out[j].k) })
		if len(out) > limit {
			out = out[:limit]
		}
		return out
	}
	var mostRead []any
	for _, e := range top(postReads, func(p string) bool { _, ok := titles[p]; return ok }, 20) {
		p := titles[e.k]
		mostRead = append(mostRead, map[string]any{"Title": p.Title, "Path": p.Path, "Day": p.Day, "Reads": readsLabel(e.n), "N": e.n})
	}
	var mostVisited []any
	for _, e := range top(pageVisits, nil, 15) {
		label := pageLabel(e.k)
		if p, ok := titles[e.k]; ok {
			label = p.Title
		}
		mostVisited = append(mostVisited, map[string]any{"Label": label, "Path": e.k, "N": e.n})
	}

	// world map: one light per country, at the mean position of its cities
	reg := placeRegion["world"]
	w := float64(len(reg.Mask[0])) * reg.Cell
	h := float64(len(reg.Mask)) * reg.Cell
	cs := top(countries, func(cc string) bool { return cc != "ZZ" }, 300)
	maxC := 1
	if len(cs) > 0 {
		maxC = cs[0].n
	}
	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg class="pl-map" viewBox="0 0 %.0f %.0f" role="img" aria-label="Map of readers' countries"><path class="pl-land" d="%s"/>`, w, h, reg.landPath)
	var rows []any
	for _, e := range cs {
		ci := countryTable[e.k]
		name := countryName(e.k)
		rows = append(rows, map[string]any{"Name": name, "N": e.n})
		if !ci.HasPos {
			continue
		}
		x := (ci.Lon - reg.Lon0) / (reg.Lon1 - reg.Lon0) * w
		y := (reg.Lat1 - ci.Lat) / (reg.Lat1 - reg.Lat0) * h
		k := math.Log1p(float64(e.n)) / math.Log1p(float64(maxC))
		r := reg.Cell * (0.45 + 1.1*k)
		label := html.EscapeString(fmt.Sprintf("%s · %d", name, e.n))
		fmt.Fprintf(&svg, `<g><title>%s</title><circle class="pl-glow" cx="%.1f" cy="%.1f" r="%.1f" style="opacity:%.2f"/><circle class="pl-light" cx="%.1f" cy="%.1f" r="%.1f"/></g>`,
			label, x, y, r*2.6, 0.18+0.35*k, x, y, r)
	}
	svg.WriteString(`</svg>`)
	since := ""
	if first != "" {
		if t, err := time.Parse("2006-01-02", first); err == nil {
			since = t.Format("2 Jan 2006")
		}
	}
	return map[string]any{
		"visits": visits, "visitors": visitors, "reads": reads, "countryCount": len(cs),
		"since": since, "hasSince": since != "", "months": months,
		"mostRead": mostRead, "hasRead": len(mostRead) > 0,
		"mostVisited": mostVisited, "hasVisited": len(mostVisited) > 0,
		"svg": svg.String(), "countries": rows, "hasCountries": len(rows) > 0,
	}
}
