package main

// /archives/films/: the films. data/library/films.json is made from films.csv by
// scripts/films-import.py on the VM (OMDb: IMDb id, director, runtime, country, poster).
// Posters live outside git in <library media>/films/posters/. Every poster is made
// exactly 2:3 here at startup (frames/<key>.jpg 200×300, frames/<key>-l.jpg 400×600):
// posters close to 2:3 are centre-cropped, anything further off is padded with its own
// edge colour.

import (
	"encoding/json"
	"fmt"
	"image"
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

type film struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Year        int    `json:"year"`
	Shelf       string `json:"shelf"`  // genre
	Format      string `json:"format"` // dvd, blu-ray, 4k ("·" between several)
	Notes       string `json:"notes"`  // about our copy
	IMDb        string `json:"imdb"`
	Director    string `json:"director"`
	Runtime     string `json:"runtime"`
	Country     string `json:"country"`
	Language    string `json:"language"`
	Poster      string `json:"poster"`
	PosterPhoto bool   `json:"poster_photo"` // a photo of our own copy

	// worked out at load time
	HasPoster       bool
	Color, Ink      string
	Hue, Sat, Light float64
	sortName        string
	ver             string
}

type filmShelf struct {
	Films []*film
	ByKey map[string]*film
	Media string // <library media>/films: posters/ + frames/
}

var filmShelves = []struct{ ID, Label string }{
	{"drama", "drama"},
	{"crime", "crime & thrillers"},
	{"fantasy-sf", "fantasy & science fiction"},
	{"adventure", "adventure"},
	{"horror", "horror & mystery"},
	{"comedy", "comedy"},
	{"romance", "romance & musicals"},
	{"animation", "animation"},
	{"series", "series & documentaries"},
}

const (
	filmW, filmH = 200, 300
)

// loadFilms reads films.json; a missing file just means an empty shelf.
func loadFilms(dir, media string) (*filmShelf, error) {
	s := &filmShelf{ByKey: map[string]*film{}, Media: filepath.Join(media, "films")}
	raw, err := os.ReadFile(filepath.Join(dir, "films.json"))
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(raw, &s.Films); err != nil {
		return s, fmt.Errorf("films.json: %w", err)
	}
	os.MkdirAll(filepath.Join(s.Media, "frames"), 0o755)
	for _, f := range s.Films {
		s.ByKey[f.Key] = f
		f.sortName = foldKey(articleless(f.Title))
		col := [3]uint8{}
		if f.Poster != "" {
			if c, err := ensureFrame(s.Media, f.Key); err == nil {
				f.HasPoster, col = true, c
				if st, err := os.Stat(filepath.Join(s.Media, "frames", f.Key+".jpg")); err == nil {
					f.ver = fmt.Sprintf("?v=%d", st.ModTime().Unix())
				}
			}
		}
		if !f.HasPoster {
			hex := libCloth[hashOf(f.Key)%uint32(len(libCloth))]
			fmt.Sscanf(hex, "#%02x%02x%02x", &col[0], &col[1], &col[2])
		}
		t := &libBook{}
		t.setColor(col)
		f.Color, f.Ink, f.Hue, f.Sat, f.Light = t.Color, t.Ink, t.Hue, t.Sat, t.Light
	}
	return s, nil
}

// articleless drops a leading English article so "The Pianist" files under P.
func articleless(t string) string {
	for _, a := range []string{"The ", "A ", "An "} {
		if strings.HasPrefix(t, a) {
			return t[len(a):]
		}
	}
	return t
}

func (lib *library) filmShelf() *filmShelf {
	if lib.Films == nil {
		lib.Films = &filmShelf{ByKey: map[string]*film{}}
	}
	return lib.Films
}

func (f *film) path() string { return "/archives/films/" + f.Key + "/" }

// firstDirector is the first name in OMDb's comma list.
func (f *film) firstDirector() string {
	return strings.TrimSpace(strings.Split(f.Director, ",")[0])
}

func (f *film) item() map[string]any {
	return map[string]any{
		"Key":      f.Key,
		"Title":    f.Title,
		"Director": f.firstDirector(),
		"Year":     yearString(f.Year),
		"Href":     f.path(),
		"Label":    strings.TrimSpace(f.Title + " (" + yearString(f.Year) + ")"),
		"HasCover": f.HasPoster,
		"Thumb":    "/archives/films/posters/" + f.Key + ".jpg" + f.ver,
		"GenStyle": fmt.Sprintf("background-color:%s;color:%s", f.Color, f.Ink),
	}
}

func filmItems(fs []*film) []any {
	out := make([]any, len(fs))
	for i, f := range fs {
		out[i] = f.item()
	}
	return out
}

// --- views ------------------------------------------------------------------------------

var filmGroupings = []struct{ ID, Label string }{
	{"shelf", "genre"}, {"title", "title"}, {"director", "director"}, {"decade", "decade"}, {"colour", "colour"},
}

type filmParams struct{ By, Group string }

func parseFilmParams(q url.Values) filmParams {
	p := filmParams{By: "shelf"}
	for _, g := range filmGroupings {
		if q.Get("by") == g.ID {
			p.By = g.ID
		}
	}
	p.Group = q.Get("g")
	return p
}

func (p filmParams) url(change func(*filmParams)) string {
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
		return "/archives/films/"
	}
	return "/archives/films/?" + q.Encode()
}

func letterGroup(name string) (string, string, float64) {
	rs := []rune(strings.ToUpper(name))
	if len(rs) == 0 || !unicode.IsLetter(rs[0]) {
		return "other", "#", 99
	}
	return strings.ToLower(string(rs[0])), string(rs[0]), float64(rs[0])
}

func filmGroupOf(by string, f *film) (id, label string, order float64) {
	switch by {
	case "colour":
		return hueFamily(&libBook{Hue: f.Hue, Sat: f.Sat, Light: f.Light})
	case "title":
		return letterGroup(f.sortName)
	case "director":
		d := f.firstDirector()
		if d == "" {
			return "unknown", "director unknown", 1e9
		}
		// grouped by surname's initial, so the crates stay a reasonable size
		parts := strings.Fields(d)
		return letterGroup(foldKey(parts[len(parts)-1]))
	case "decade":
		if f.Year == 0 {
			return "unknown", "year unknown", 9999
		}
		d := f.Year / 10 * 10
		return fmt.Sprint(d), fmt.Sprintf("%ds", d), float64(d)
	}
	for i, s := range filmShelves {
		if s.ID == f.Shelf {
			return s.ID, s.Label, float64(i)
		}
	}
	return "unsorted", "unsorted", 98
}

type filmGroup struct {
	ID, Label string
	Order     float64
	Films     []*film
}

func (s *filmShelf) groups(by string) []*filmGroup {
	m := map[string]*filmGroup{}
	var out []*filmGroup
	for _, f := range s.Films {
		id, label, order := filmGroupOf(by, f)
		g := m[id]
		if g == nil {
			g = &filmGroup{ID: id, Label: label, Order: order}
			m[id] = g
			out = append(out, g)
		}
		g.Films = append(g.Films, f)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	for _, g := range out {
		fs := g.Films
		sort.SliceStable(fs, func(i, j int) bool {
			a, b := fs[i], fs[j]
			switch by {
			case "colour":
				if a.Hue != b.Hue {
					return a.Hue < b.Hue
				}
			case "decade", "shelf":
				if a.Year != b.Year {
					return a.Year < b.Year
				}
			case "director":
				da, db := surnameKey(a), surnameKey(b)
				if da != db {
					return da < db
				}
				if a.Year != b.Year {
					return a.Year < b.Year
				}
			}
			return a.sortName < b.sortName
		})
	}
	return out
}

func surnameKey(f *film) string {
	parts := strings.Fields(f.firstDirector())
	if len(parts) == 0 {
		return ""
	}
	return foldKey(parts[len(parts)-1] + " " + strings.Join(parts[:len(parts)-1], " "))
}

func buildFilmsData(lib *library, q url.Values) map[string]any {
	s := lib.filmShelf()
	p := parseFilmParams(q)
	var groupings []any
	for _, g := range filmGroupings {
		id := g.ID
		groupings = append(groupings, map[string]any{"Label": g.Label, "Href": p.url(func(x *filmParams) { x.By, x.Group = id, "" }), "Active": p.By == id})
	}
	all := s.groups(p.By)
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
			"Count": plural(len(g.Films), "film"),
			"Href":  p.url(func(x *filmParams) { x.Group = id }),
			"Items": filmItems(g.Films),
		})
	}
	n := len(s.Films)
	title := "films"
	if focus != "" {
		title = "films · " + focus
	}
	unit := map[string]string{"shelf": "genre", "title": "letter", "director": "letter", "decade": "decade", "colour": "colour"}[p.By]
	return map[string]any{
		"count":     n,
		"stats":     plural(n, "film") + " · " + plural(len(all), unit),
		"groupings": groupings,
		"shelves":   shelves,
		"hasFocus":  focus != "",
		"focus":     focus,
		"allHref":   p.url(func(x *filmParams) { x.Group = "" }),
		"pageTitle": title + " | cptx",
		"shareURL":  siteURL + p.url(nil),
		"empty":     n == 0,
		"canonical": siteURL + "/archives/films/",
	}
}

var filmFormats = map[string]string{"dvd": "DVD", "blu-ray": "Blu-ray", "4k": "4K Blu-ray"}

func (f *film) formatLabel() string {
	var out []string
	for _, x := range strings.Split(f.Format, "·") {
		x = strings.TrimSpace(x)
		if l, ok := filmFormats[x]; ok {
			out = append(out, l)
		} else if x != "" {
			out = append(out, x)
		}
	}
	return strings.Join(out, " and ")
}

func buildFilmData(lib *library, f *film) map[string]any {
	s := lib.filmShelf()
	_, shelfLabel, _ := filmGroupOf("shelf", f)
	var neighbours []*film
	for _, g := range s.groups("shelf") {
		if g.ID != f.Shelf {
			continue
		}
		for i, x := range g.Films {
			if x != f {
				continue
			}
			lo, hi := max(i-4, 0), min(i+5, len(g.Films))
			for _, y := range g.Films[lo:hi] {
				if y != f {
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
	add("released", yearString(f.Year))
	add("directed by", f.Director)
	add("runtime", f.Runtime)
	add("country", f.Country)
	add("language", f.Language)
	add("genre", shelfLabel)
	copyNote := f.formatLabel()
	if f.Notes != "" {
		copyNote = strings.TrimSpace(copyNote + "; " + f.Notes)
		copyNote = strings.TrimPrefix(copyNote, "; ")
	}
	add("our copy", copyNote)

	q := url.QueryEscape(strings.TrimSpace(f.Title + " " + yearString(f.Year)))
	watch := []any{
		map[string]any{"Label": "trailer", "Href": "https://www.youtube.com/results?search_query=" + url.QueryEscape(f.Title+" "+yearString(f.Year)+" trailer"), "Search": true},
		map[string]any{"Label": "where to stream", "Href": "https://www.justwatch.com/ee/search?q=" + url.QueryEscape(f.Title), "Search": true},
	}
	var about []any
	if f.IMDb != "" {
		about = append(about,
			map[string]any{"Label": "IMDb", "Href": "https://www.imdb.com/title/" + f.IMDb + "/"},
			map[string]any{"Label": "Letterboxd", "Href": "https://letterboxd.com/imdb/" + f.IMDb + "/"},
		)
	} else {
		about = append(about, map[string]any{"Label": "IMDb (search)", "Href": "https://www.imdb.com/find/?q=" + q})
	}
	desc := f.Title + " (" + yearString(f.Year) + ")"
	if d := f.firstDirector(); d != "" {
		desc += ", directed by " + d
	}
	return map[string]any{
		"film":          f.item(),
		"title":         f.Title,
		"director":      f.firstDirector(),
		"hasDirector":   f.Director != "",
		"year":          yearString(f.Year),
		"large":         "/archives/films/posters/" + f.Key + "-l.jpg" + f.ver,
		"facts":         facts,
		"watch":         watch,
		"about":         about,
		"posterPhoto":   f.HasPoster && f.PosterPhoto,
		"shelfLabel":    shelfLabel,
		"shelfHref":     filmParams{By: "shelf", Group: f.Shelf}.url(nil),
		"neighbours":    filmItems(neighbours),
		"hasNeighbours": len(neighbours) > 0,
		"pageTitle":     f.Title + " | films | cptx",
		"description":   desc + " — one of the films Merzeede and I own.",
		"canonical":     siteURL + f.path(),
	}
}

// filmsCard is the archives hub card: count + a few posters.
func filmsCard(lib *library) map[string]any {
	s := lib.filmShelf()
	var withPoster []*film
	for _, f := range s.Films {
		if f.HasPoster {
			withPoster = append(withPoster, f)
		}
	}
	pick := withPoster
	if len(pick) == 0 {
		pick = s.Films
	}
	if len(pick) > 9 {
		step := float64(len(pick)) / 9
		var out []*film
		for i := 0.0; int(i) < len(pick) && len(out) < 9; i += step {
			out = append(out, pick[int(i)])
		}
		pick = out
	}
	return map[string]any{"filmsCount": len(s.Films), "hasFilms": len(s.Films) > 0, "filmsStrip": filmItems(pick)}
}

// serveFilms handles everything under /archives/films/.
func serveFilms(w http.ResponseWriter, r *http.Request, lib *library, render func(http.ResponseWriter, string, map[string]any), notFound http.HandlerFunc) {
	s := lib.filmShelf()
	path := r.URL.Path
	switch {
	case path == "/archives/films/":
		render(w, "FilmsPage", buildFilmsData(lib, r.URL.Query()))
	case strings.HasPrefix(path, "/archives/films/posters/"):
		name := strings.TrimPrefix(path, "/archives/films/posters/")
		key, suffix := strings.TrimSuffix(name, ".jpg"), ""
		if strings.HasSuffix(key, "-l") {
			key, suffix = strings.TrimSuffix(key, "-l"), "-l"
		}
		f := s.ByKey[key]
		if f == nil || !f.HasPoster || !strings.HasSuffix(name, ".jpg") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=604800")
		http.ServeFile(w, r, filepath.Join(s.Media, "frames", key+suffix+".jpg"))
	default:
		key := strings.Trim(strings.TrimPrefix(path, "/archives/films/"), "/")
		f := s.ByKey[key]
		if f == nil || !strings.HasSuffix(path, "/") || strings.Contains(key, "/") {
			notFound(w, r)
			return
		}
		render(w, "FilmPage", buildFilmData(lib, f))
	}
}

// ensureFrame makes frames/<key>.jpg (200×300) and frames/<key>-l.jpg (400×600) from
// posters/<key>.jpg whenever the poster is newer, and returns the poster's main colour.
func ensureFrame(dir, key string) ([3]uint8, error) {
	src := filepath.Join(dir, "posters", key+".jpg")
	small := filepath.Join(dir, "frames", key+".jpg")
	large := filepath.Join(dir, "frames", key+"-l.jpg")
	st, err := os.Stat(src)
	if err != nil {
		return [3]uint8{}, err
	}
	if t, err := os.Stat(small); err == nil && !t.ModTime().Before(st.ModTime()) {
		if _, err := os.Stat(large); err == nil {
			if fh, err := os.Open(small); err == nil {
				img, _, err := image.Decode(fh)
				fh.Close()
				if err == nil {
					return mainColor(img), nil
				}
			}
		}
	}
	fh, err := os.Open(src)
	if err != nil {
		return [3]uint8{}, err
	}
	img, _, err := image.Decode(fh)
	fh.Close()
	if err != nil {
		return [3]uint8{}, err
	}
	fr := frameOf(img)
	var thumb image.Image
	for _, out := range []struct {
		path string
		w    int
	}{{large, 2 * filmW}, {small, filmW}} {
		w := min(out.w, fr.Bounds().Dx()) &^ 1
		h := w * 3 / 2
		t := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.CatmullRom.Scale(t, t.Bounds(), fr, fr.Bounds(), draw.Over, nil)
		if wf, err := os.Create(out.path); err == nil {
			jpeg.Encode(wf, t, &jpeg.Options{Quality: 85})
			wf.Close()
		}
		thumb = t
	}
	return mainColor(thumb), nil
}

// frameOf returns img as an exact 2:3 portrait: within ~12% of 2:3 it is centre-cropped to
// the largest 2:3 box inside it, otherwise padded with the colour of its edges to the
// smallest 2:3 box around it, so nothing is cut.
func frameOf(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w*3 == h*2 {
		return img
	}
	r := float64(w) / float64(h) / (2.0 / 3.0)
	if k := min(w/2, h/3); r >= 0.88 && r <= 1.12 && k > 0 {
		cw, ch := 2*k, 3*k
		x0, y0 := b.Min.X+(w-cw)/2, b.Min.Y+(h-ch)/2
		out := image.NewRGBA(image.Rect(0, 0, cw, ch))
		draw.Draw(out, out.Bounds(), img, image.Pt(x0, y0), draw.Src)
		return out
	}
	k := max((w+1)/2, (h+2)/3)
	ow, oh := 2*k, 3*k
	out := image.NewRGBA(image.Rect(0, 0, ow, oh))
	draw.Draw(out, out.Bounds(), &image.Uniform{edgeColour(img)}, image.Point{}, draw.Src)
	off := image.Pt((ow-w)/2, (oh-h)/2)
	draw.Draw(out, image.Rectangle{off, off.Add(image.Pt(w, h))}, img, b.Min, draw.Src)
	return out
}
