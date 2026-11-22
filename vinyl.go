package main

// /archives/vinyl/: the records. data/library/vinyl.json is made from vinyl.csv by
// scripts/vinyl-import.py on the VM (MusicBrainz → Cover Art Archive, Wikidata →
// Wikipedia, Spotify, YouTube Music). Covers live outside git in
// <library media>/vinyl/covers/. Every cover is made exactly square here at startup
// (squares/<key>.jpg 300 px, squares/<key>-l.jpg 600 px): near-square scans are
// centre-cropped, anything further off is padded with its own edge colour.

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/image/draw"
)

type vinylRecord struct {
	Key          string `json:"key"`
	Artist       string `json:"artist"`
	Title        string `json:"title"`
	Year         int    `json:"year"`
	Shelf        string `json:"shelf"`
	Notes        string `json:"notes"`
	MBID         string `json:"mbid"`
	Spotify      string `json:"spotify"`
	YouTubeMusic string `json:"youtube_music"`
	ArtistWiki   string `json:"artist_wikipedia"`
	Cover        string `json:"cover"`
	CoverPhoto   bool   `json:"cover_photo"` // a photo of our own copy, not the archive's cover

	// worked out at load time
	HasCover        bool
	Color, Ink      string
	Hue, Sat, Light float64
	sortName        string
	ver             string // cover file time, appended to cover URLs
}

type vinylCrate struct {
	Records []*vinylRecord
	ByKey   map[string]*vinylRecord
	Media   string // <library media>/vinyl: covers/ + thumbs/
}

var vinylShelves = []struct{ ID, Label string }{
	{"rock", "rock"},
	{"pop", "pop"},
	{"songwriters", "songwriters & folk"},
	{"espanol", "en español"},
	{"soul-funk-jazz", "soul, funk & jazz"},
	{"hip-hop", "hip-hop"},
	{"electronic", "electronic"},
	{"soundtracks", "soundtracks & stage"},
	{"eesti", "eesti"},
	{"classical", "classical"},
}

// loadVinyl reads vinyl.json; a missing file just means an empty crate.
func loadVinyl(dir, media string) (*vinylCrate, error) {
	c := &vinylCrate{ByKey: map[string]*vinylRecord{}, Media: filepath.Join(media, "vinyl")}
	raw, err := os.ReadFile(filepath.Join(dir, "vinyl.json"))
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(raw, &c.Records); err != nil {
		return c, fmt.Errorf("vinyl.json: %w", err)
	}
	os.MkdirAll(filepath.Join(c.Media, "squares"), 0o755)
	for _, r := range c.Records {
		c.ByKey[r.Key] = r
		r.sortName = foldKey(strings.TrimPrefix(r.Artist, "The ") + " " + r.Title)
		col := [3]uint8{}
		if r.Cover != "" {
			if _, _, mc, err := ensureSquare(c.Media, r.Key); err == nil {
				r.HasCover, col = true, mc
				if st, err := os.Stat(filepath.Join(c.Media, "squares", r.Key+".jpg")); err == nil {
					r.ver = fmt.Sprintf("?v=%d", st.ModTime().Unix())
				}
			}
		}
		if !r.HasCover {
			hex := libCloth[hashOf(r.Key)%uint32(len(libCloth))]
			fmt.Sscanf(hex, "#%02x%02x%02x", &col[0], &col[1], &col[2])
		}
		t := &libBook{}
		t.setColor(col)
		r.Color, r.Ink, r.Hue, r.Sat, r.Light = t.Color, t.Ink, t.Hue, t.Sat, t.Light
	}
	return c, nil
}

func (lib *library) crate() *vinylCrate {
	if lib.Vinyl == nil {
		lib.Vinyl = &vinylCrate{ByKey: map[string]*vinylRecord{}}
	}
	return lib.Vinyl
}

func (r *vinylRecord) path() string { return "/archives/vinyl/" + r.Key + "/" }

func (r *vinylRecord) item() map[string]any {
	return map[string]any{
		"Key":       r.Key,
		"Title":     r.Title,
		"Artist":    r.Artist,
		"Year":      yearString(r.Year),
		"Href":      r.path(),
		"Label":     strings.TrimSpace(r.Title + " — " + r.Artist),
		"HasCover":  r.HasCover,
		"Thumb":     "/archives/vinyl/covers/" + r.Key + ".jpg" + r.ver,
		"GenStyle":  fmt.Sprintf("background-color:%s;color:%s", r.Color, r.Ink),
		"DiscStyle": "--vn-label:" + r.Color,
	}
}

func vinylItems(rs []*vinylRecord) []any {
	out := make([]any, len(rs))
	for i, r := range rs {
		out[i] = r.item()
	}
	return out
}

// --- views ------------------------------------------------------------------------------

var vinylGroupings = []struct{ ID, Label string }{
	{"shelf", "genre"}, {"artist", "artist"}, {"decade", "decade"}, {"colour", "colour"},
}

type vinylParams struct{ By, Group string }

func parseVinylParams(q url.Values) vinylParams {
	p := vinylParams{By: "shelf"}
	for _, g := range vinylGroupings {
		if q.Get("by") == g.ID {
			p.By = g.ID
		}
	}
	p.Group = q.Get("g")
	return p
}

func (p vinylParams) url(change func(*vinylParams)) string {
	c := p
	if change != nil {
		change(&c)
	}
	q := url.Values{}
	if c.By != "shelf" {
		q.Set("by", c.By)
	}
	if c.Group != "" {
		q.Set("g", c.Group)
	}
	if len(q) == 0 {
		return "/archives/vinyl/"
	}
	return "/archives/vinyl/?" + q.Encode()
}

func vinylGroupOf(by string, r *vinylRecord) (id, label string, order float64) {
	switch by {
	case "colour":
		return hueFamily(&libBook{Hue: r.Hue, Sat: r.Sat, Light: r.Light})
	case "artist":
		rs := []rune(strings.ToUpper(r.sortName))
		if len(rs) == 0 || !unicode.IsLetter(rs[0]) {
			return "other", "#", 99
		}
		return strings.ToLower(string(rs[0])), string(rs[0]), float64(rs[0])
	case "decade":
		if r.Year == 0 {
			return "unknown", "year unknown", 9999
		}
		d := r.Year / 10 * 10
		return fmt.Sprint(d), fmt.Sprintf("%ds", d), float64(d)
	}
	for i, s := range vinylShelves {
		if s.ID == r.Shelf {
			return s.ID, s.Label, float64(i)
		}
	}
	return "unsorted", "unsorted", 98
}

type vinylGroup struct {
	ID, Label string
	Order     float64
	Records   []*vinylRecord
}

func (c *vinylCrate) groups(by string) []*vinylGroup {
	m := map[string]*vinylGroup{}
	var out []*vinylGroup
	for _, r := range c.Records {
		id, label, order := vinylGroupOf(by, r)
		g := m[id]
		if g == nil {
			g = &vinylGroup{ID: id, Label: label, Order: order}
			m[id] = g
			out = append(out, g)
		}
		g.Records = append(g.Records, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	for _, g := range out {
		rs := g.Records
		sort.SliceStable(rs, func(i, j int) bool {
			a, b := rs[i], rs[j]
			if by == "colour" && a.Hue != b.Hue {
				return a.Hue < b.Hue
			}
			if by == "decade" && a.Year != b.Year {
				return a.Year < b.Year
			}
			return a.sortName < b.sortName
		})
	}
	return out
}

func buildVinylData(lib *library, q url.Values) map[string]any {
	c := lib.crate()
	p := parseVinylParams(q)
	var groupings []any
	for _, g := range vinylGroupings {
		id := g.ID
		groupings = append(groupings, map[string]any{"Label": g.Label, "Href": p.url(func(x *vinylParams) { x.By, x.Group = id, "" }), "Active": p.By == id})
	}
	all := c.groups(p.By)
	var shelves []any
	focus := ""
	for _, g := range all {
		if p.Group != "" && g.ID != p.Group {
			continue
		}
		id := g.ID
		if p.Group != "" {
			focus = g.Label
		}
		shelves = append(shelves, map[string]any{
			"Label": g.Label,
			"Count": plural(len(g.Records), "record"),
			"Href":  p.url(func(x *vinylParams) { x.Group = id }),
			"Items": vinylItems(g.Records),
		})
	}
	n := len(c.Records)
	title := "vinyl"
	if focus != "" {
		title = "vinyl · " + focus
	}
	return map[string]any{
		"count":     n,
		"stats":     plural(n, "record") + " · " + plural(len(all), map[string]string{"shelf": "genre", "artist": "letter", "decade": "decade", "colour": "colour"}[p.By]),
		"groupings": groupings,
		"shelves":   shelves,
		"hasFocus":  focus != "",
		"focus":     focus,
		"allHref":   p.url(func(x *vinylParams) { x.Group = "" }),
		"pageTitle": title + " | cptx",
		"shareURL":  siteURL + p.url(nil),
		"empty":     n == 0,
		"canonical": siteURL + "/archives/vinyl/",
	}
}

func vinylSearchQuery(r *vinylRecord) string {
	return strings.TrimSpace(strings.NewReplacer(" / ", " ").Replace(r.Artist) + " " + r.Title)
}

func buildRecordData(lib *library, r *vinylRecord) map[string]any {
	c := lib.crate()
	_, shelfLabel, _ := vinylGroupOf("shelf", r)
	var neighbours []*vinylRecord
	for _, g := range c.groups("shelf") {
		if g.ID != r.Shelf {
			continue
		}
		for i, x := range g.Records {
			if x != r {
				continue
			}
			lo, hi := max(i-4, 0), min(i+5, len(g.Records))
			for _, y := range g.Records[lo:hi] {
				if y != r {
					neighbours = append(neighbours, y)
				}
			}
		}
	}
	var facts []any
	add := func(label, value string) {
		if value != "" {
			facts = append(facts, map[string]any{"Label": label, "Value": value})
		}
	}
	add("released", yearString(r.Year))
	add("genre", shelfLabel)
	add("our copy", r.Notes)

	q := url.QueryEscape(vinylSearchQuery(r))
	listen := []any{}
	if r.Spotify != "" {
		listen = append(listen, map[string]any{"Label": "Spotify", "Href": r.Spotify, "Search": false})
	} else {
		listen = append(listen, map[string]any{"Label": "Spotify", "Href": "https://open.spotify.com/search/" + strings.ReplaceAll(q, "+", "%20"), "Search": true})
	}
	if r.YouTubeMusic != "" {
		listen = append(listen, map[string]any{"Label": "YouTube Music", "Href": r.YouTubeMusic, "Search": false})
	} else {
		listen = append(listen, map[string]any{"Label": "YouTube Music", "Href": "https://music.youtube.com/search?q=" + q, "Search": true})
	}
	var about []any
	if r.ArtistWiki != "" {
		about = append(about, map[string]any{"Label": strings.Split(r.Artist, " / ")[0] + " on Wikipedia", "Href": r.ArtistWiki})
	}
	if r.MBID != "" {
		about = append(about, map[string]any{"Label": "MusicBrainz", "Href": "https://musicbrainz.org/release-group/" + r.MBID})
	}
	return map[string]any{
		"record":        r.item(),
		"title":         r.Title,
		"artist":        r.Artist,
		"year":          yearString(r.Year),
		"large":         "/archives/vinyl/covers/" + r.Key + "-l.jpg" + r.ver,
		"facts":         facts,
		"listen":        listen,
		"about":         about,
		"hasAbout":      len(about) > 0,
		"coverPhoto":    r.HasCover && r.CoverPhoto,
		"shelfLabel":    shelfLabel,
		"shelfHref":     vinylParams{By: "shelf", Group: r.Shelf}.url(nil),
		"neighbours":    vinylItems(neighbours),
		"hasNeighbours": len(neighbours) > 0,
		"pageTitle":     r.Title + " | vinyl | cptx",
		"description":   r.Title + " by " + r.Artist + " — one of the records Merzeede and I own.",
		"canonical":     siteURL + r.path(),
	}
}

// vinylCard is the archives hub card: count + a few covers.
func vinylCard(lib *library) map[string]any {
	c := lib.crate()
	var withCover []*vinylRecord
	for _, r := range c.Records {
		if r.HasCover {
			withCover = append(withCover, r)
		}
	}
	pick := withCover
	if len(pick) == 0 {
		pick = c.Records
	}
	if len(pick) > 9 {
		step := float64(len(pick)) / 9
		var s []*vinylRecord
		for i := 0.0; int(i) < len(pick) && len(s) < 9; i += step {
			s = append(s, pick[int(i)])
		}
		pick = s
	}
	return map[string]any{"vinylCount": len(c.Records), "hasVinyl": len(c.Records) > 0, "vinylStrip": vinylItems(pick)}
}

// serveVinyl handles everything under /archives/vinyl/.
func serveVinyl(w http.ResponseWriter, r *http.Request, lib *library, render func(http.ResponseWriter, string, map[string]any), notFound http.HandlerFunc) {
	c := lib.crate()
	path := r.URL.Path
	switch {
	case path == "/archives/vinyl/":
		render(w, "VinylPage", buildVinylData(lib, r.URL.Query()))
	case strings.HasPrefix(path, "/archives/vinyl/covers/"):
		name := strings.TrimPrefix(path, "/archives/vinyl/covers/")
		key, suffix := strings.TrimSuffix(name, ".jpg"), ""
		if strings.HasSuffix(key, "-l") {
			key, suffix = strings.TrimSuffix(key, "-l"), "-l"
		}
		rec := c.ByKey[key]
		if rec == nil || !rec.HasCover || !strings.HasSuffix(name, ".jpg") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=604800")
		http.ServeFile(w, r, filepath.Join(c.Media, "squares", key+suffix+".jpg"))
	default:
		key := strings.Trim(strings.TrimPrefix(path, "/archives/vinyl/"), "/")
		rec := c.ByKey[key]
		if rec == nil || !strings.HasSuffix(path, "/") || strings.Contains(key, "/") {
			notFound(w, r)
			return
		}
		render(w, "RecordPage", buildRecordData(lib, rec))
	}
}

// ensureSquare makes squares/<key>.jpg (300 px) and squares/<key>-l.jpg (600 px) from
// covers/<key>.jpg whenever the cover is newer, and returns the cover's main colour.
// Sleeves are square: a scan within 15% of square is centre-cropped (scan margins);
// anything further off is padded with the average colour of its edges, so nothing is cut.
func ensureSquare(dir, key string) (int, int, [3]uint8, error) {
	src := filepath.Join(dir, "covers", key+".jpg")
	small := filepath.Join(dir, "squares", key+".jpg")
	large := filepath.Join(dir, "squares", key+"-l.jpg")
	st, err := os.Stat(src)
	if err != nil {
		return 0, 0, [3]uint8{}, err
	}
	if t, err := os.Stat(small); err == nil && !t.ModTime().Before(st.ModTime()) {
		if _, err := os.Stat(large); err == nil {
			if f, err := os.Open(small); err == nil {
				img, _, err := image.Decode(f)
				f.Close()
				if err == nil {
					return 300, 300, mainColor(img), nil
				}
			}
		}
	}
	f, err := os.Open(src)
	if err != nil {
		return 0, 0, [3]uint8{}, err
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return 0, 0, [3]uint8{}, err
	}
	sq := squareOf(img)
	var thumb image.Image
	for _, out := range []struct {
		path string
		size int
	}{{large, 600}, {small, 300}} {
		n := out.size
		if b := sq.Bounds().Dx(); b < n {
			n = b
		}
		t := image.NewRGBA(image.Rect(0, 0, n, n))
		draw.CatmullRom.Scale(t, t.Bounds(), sq, sq.Bounds(), draw.Over, nil)
		if w, err := os.Create(out.path); err == nil {
			jpeg.Encode(w, t, &jpeg.Options{Quality: 85})
			w.Close()
		}
		thumb = t
	}
	return 300, 300, mainColor(thumb), nil
}

func squareOf(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == h {
		return img
	}
	r := float64(w) / float64(h)
	if r >= 0.87 && r <= 1.15 {
		side := min(w, h)
		x0, y0 := b.Min.X+(w-side)/2, b.Min.Y+(h-side)/2
		out := image.NewRGBA(image.Rect(0, 0, side, side))
		draw.Draw(out, out.Bounds(), img, image.Pt(x0, y0), draw.Src)
		return out
	}
	side := max(w, h)
	out := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(out, out.Bounds(), &image.Uniform{edgeColour(img)}, image.Point{}, draw.Src)
	off := image.Pt((side-w)/2, (side-h)/2)
	draw.Draw(out, image.Rectangle{off, off.Add(image.Pt(w, h))}, img, b.Min, draw.Src)
	return out
}

// edgeColour averages the outermost pixels, the colour a sleeve's border would have.
func edgeColour(img image.Image) color.Color {
	b := img.Bounds()
	var r, g, bl, n uint64
	add := func(x, y int) {
		cr, cg, cb, _ := img.At(x, y).RGBA()
		r, g, bl, n = r+uint64(cr>>8), g+uint64(cg>>8), bl+uint64(cb>>8), n+1
	}
	for x := b.Min.X; x < b.Max.X; x += 2 {
		add(x, b.Min.Y)
		add(x, b.Max.Y-1)
	}
	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		add(b.Min.X, y)
		add(b.Max.X-1, y)
	}
	return color.RGBA{uint8(r / n), uint8(g / n), uint8(bl / n), 255}
}
