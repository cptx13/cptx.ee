package main

// /archives/: collections. The first one is the library (L-01…): books from a Libib
// export, enriched by scripts/library-import.py into books.json + covers/.
// Covers live outside git (like photos); thumbnails and each cover's main colour are
// worked out here at startup.

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/dhamidi/htmlc"
	"golang.org/x/image/draw"
	"golang.org/x/text/unicode/norm"
)

type libBook struct {
	Key         string   `json:"key"`
	ISBN13      string   `json:"isbn13"`
	ISBN10      string   `json:"isbn10"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Series      string   `json:"series"`
	Author      string   `json:"author"`
	Publisher   string   `json:"publisher"`
	Year        int      `json:"year"`
	FirstYear   int      `json:"first_year"`
	Pages       int      `json:"pages"`
	Lang        string   `json:"lang"`
	Subjects    []string `json:"subjects"`
	Categories  []string `json:"categories"`
	Shelf       string   `json:"shelf"`
	Own         bool     `json:"own"`
	Buy         string   `json:"buy"`
	Cover       string   `json:"cover"`
	OpenLibrary string   `json:"openlibrary"`
	Google      string   `json:"google"`

	// worked out at load time
	HasCover        bool
	ThumbW, ThumbH  int
	Color, Ink      string
	Hue, Sat, Light float64
	sortName        string
	Mentions        []addonPost
}

type library struct {
	Books []*libBook
	ByKey map[string]*libBook
	Dir   string
}

var libShelves = []struct{ ID, Label string }{
	{"fiction", "fiction"},
	{"fantasy-sf", "fantasy & science fiction"},
	{"classics", "classics & myth"},
	{"poetry", "poetry"},
	{"lives", "lives, letters & essays"},
	{"ideas", "ideas"},
	{"history", "history"},
	{"altered-states", "altered states"},
	{"science", "science & nature"},
	{"body", "body"},
	{"art-design", "art & design"},
	{"food", "food"},
	{"games-sport", "games & sport"},
	{"mallorca", "mallorca"},
	{"young", "young readers"},
	{"comics", "comics & manga"},
	{"reference", "reference"},
	{"mine", "written by me"},
}

var libLangs = map[string]string{"es": "español", "en": "English", "et": "eesti", "bn": "বাংলা", "de": "Deutsch", "nl": "Nederlands", "fr": "français", "ca": "català", "it": "italiano", "pt": "português"}

// cloth colours for books without a cover image
var libCloth = []string{"#7a2e2e", "#2f4858", "#5b6d3a", "#8a6d1f", "#3d3b6e", "#6b3f5b", "#2e5e4e", "#9a4b2b", "#4a4a4a", "#1f3a5f", "#7d5a50", "#3f6e6e"}

func libraryDir() string {
	if d := os.Getenv("LIBRARY_DIR"); d != "" {
		return d
	}
	return filepath.Join("data", "library")
}

func hashOf(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// loadLibrary reads books.json, makes missing thumbnails and measures cover colours.
// A missing books.json is not an error: the library is simply empty.
func loadLibrary(dir string, posts []addonPost, texts map[string]string) (*library, error) {
	lib := &library{ByKey: map[string]*libBook{}, Dir: dir}
	raw, err := os.ReadFile(filepath.Join(dir, "books.json"))
	if os.IsNotExist(err) {
		return lib, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &lib.Books); err != nil {
		return nil, fmt.Errorf("books.json: %w", err)
	}
	os.MkdirAll(filepath.Join(dir, "thumbs"), 0o755)
	for _, b := range lib.Books {
		lib.ByKey[b.Key] = b
		b.sortName = authorSortName(b.Author)
		if b.Cover != "" {
			if w, h, c, err := ensureThumb(dir, b.Key); err == nil {
				b.HasCover, b.ThumbW, b.ThumbH = true, w, h
				b.setColor(c)
			}
		}
		if !b.HasCover {
			hex := libCloth[hashOf(b.Key)%uint32(len(libCloth))]
			var r, g, bl uint8
			fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &bl)
			b.setColor([3]uint8{r, g, bl})
		}
	}
	lib.findMentions(posts, texts)
	return lib, nil
}

func authorSortName(a string) string {
	a = strings.TrimSpace(strings.Split(strings.Split(a, " & ")[0], ",")[0])
	ws := strings.Fields(a)
	if len(ws) == 0 {
		return "~"
	}
	last := ws[len(ws)-1]
	return foldKey(last + " " + a)
}

// foldKey lowercases and strips accents so "Álvarez" sorts with "a".
func foldKey(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(strings.ToLower(s)) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (b *libBook) setColor(c [3]uint8) {
	b.Color = fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
	r, g, bl := float64(c[0])/255, float64(c[1])/255, float64(c[2])/255
	b.Hue, b.Sat, b.Light = rgbToHSL(r, g, bl)
	if 0.2126*r+0.7152*g+0.0722*bl > 0.55 {
		b.Ink = "#1a1a1a"
	} else {
		b.Ink = "#f4f1ea"
	}
}

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max, min := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h * 60, s, l
}

// ensureThumb makes thumbs/<key>.jpg (300px tall) once and returns its size and the
// cover's main colour.
func ensureThumb(dir, key string) (int, int, [3]uint8, error) {
	src := filepath.Join(dir, "covers", key+".jpg")
	dst := filepath.Join(dir, "thumbs", key+".jpg")
	var thumb image.Image
	if f, err := os.Open(dst); err == nil {
		thumb, _, err = image.Decode(f)
		f.Close()
		if err != nil {
			thumb = nil
		}
	}
	if thumb == nil {
		f, err := os.Open(src)
		if err != nil {
			return 0, 0, [3]uint8{}, err
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			return 0, 0, [3]uint8{}, err
		}
		b := img.Bounds()
		if b.Dx() < 20 || b.Dy() < 20 {
			return 0, 0, [3]uint8{}, fmt.Errorf("cover too small")
		}
		h := 300
		if b.Dy() < h {
			h = b.Dy()
		}
		w := b.Dx() * h / b.Dy()
		t := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.CatmullRom.Scale(t, t.Bounds(), img, b, draw.Over, nil)
		out, err := os.Create(dst)
		if err == nil {
			jpeg.Encode(out, t, &jpeg.Options{Quality: 82})
			out.Close()
		}
		thumb = t
	}
	b := thumb.Bounds()
	return b.Dx(), b.Dy(), mainColor(thumb), nil
}

// mainColor picks the most common colourful tone of an image (near-white and
// near-black count less), so a red cover with white letters reads as red.
func mainColor(img image.Image) [3]uint8 {
	type acc struct{ w, r, g, b float64 }
	bins := map[int]*acc{}
	bd := img.Bounds()
	step := bd.Dx() / 40
	if step < 1 {
		step = 1
	}
	for y := bd.Min.Y; y < bd.Max.Y; y += step {
		for x := bd.Min.X; x < bd.Max.X; x += step {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r, g, b := float64(r16>>8), float64(g16>>8), float64(b16>>8)
			_, s, l := rgbToHSL(r/255, g/255, b/255)
			w := 0.25 + s
			if l > 0.9 || l < 0.08 {
				w *= 0.3
			}
			k := int(r)/32<<6 | int(g)/32<<3 | int(b)/32
			a := bins[k]
			if a == nil {
				a = &acc{}
				bins[k] = a
			}
			a.w += w
			a.r += r * w
			a.g += g * w
			a.b += b * w
		}
	}
	var best *acc
	for _, a := range bins {
		if best == nil || a.w > best.w {
			best = a
		}
	}
	if best == nil || best.w == 0 {
		return [3]uint8{120, 120, 120}
	}
	return [3]uint8{uint8(best.r / best.w), uint8(best.g / best.w), uint8(best.b / best.w)}
}

// findMentions links books to posts that name them (title, or the author's full name).
func (lib *library) findMentions(posts []addonPost, texts map[string]string) {
	low := map[string]string{}
	for path, t := range texts {
		low[path] = foldKey(t)
	}
	for _, b := range lib.Books {
		title := foldKey(strings.TrimSpace(strings.Split(b.Title, ":")[0]))
		author := foldKey(strings.Split(b.Author, " & ")[0])
		surname := ""
		if ws := strings.Fields(author); len(ws) > 1 {
			surname = ws[len(ws)-1]
		} else {
			author = ""
		}
		if len(title) < 6 && author == "" {
			continue
		}
		titleRe, authorRe, surnameRe := phraseRe(title), phraseRe(author), phraseRe(surname)
		for _, p := range posts {
			t := low[p.Path]
			hit := false
			if len(title) >= 6 && strings.Contains(t, title) && titleRe.MatchString(t) &&
				(len([]rune(title)) >= 14 || (surname != "" && strings.Contains(t, surname) && surnameRe.MatchString(t))) {
				hit = true
			}
			if !hit && len(author) >= 8 && strings.Contains(t, author) && authorRe.MatchString(t) {
				hit = true
			}
			if hit {
				b.Mentions = append(b.Mentions, p)
			}
		}
		sort.Slice(b.Mentions, func(i, j int) bool { return b.Mentions[i].Date.After(b.Mentions[j].Date) })
	}
}

func phraseRe(phrase string) *regexp.Regexp {
	return regexp.MustCompile(`(^|[^\p{L}])` + regexp.QuoteMeta(phrase) + `($|[^\p{L}])`)
}

// --- views ------------------------------------------------------------------------------

type libParams struct{ By, View, Group string }

func (p libParams) url(change func(*libParams)) string {
	c := p
	if change != nil {
		change(&c)
	}
	q := url.Values{}
	if c.By != "shelf" {
		q.Set("by", c.By)
	}
	if c.View != "covers" {
		q.Set("view", c.View)
	}
	if c.Group != "" {
		q.Set("g", c.Group)
	}
	if len(q) == 0 {
		return "/archives/library/"
	}
	return "/archives/library/?" + q.Encode()
}

var libGroupings = []struct{ ID, Label string }{
	{"shelf", "shelf"}, {"colour", "colour"}, {"author", "author"}, {"decade", "decade"}, {"lang", "language"}, {"size", "size"},
}

func parseLibParams(q url.Values) libParams {
	p := libParams{By: "shelf", View: "covers"}
	for _, g := range libGroupings {
		if q.Get("by") == g.ID {
			p.By = g.ID
		}
	}
	if q.Get("view") == "spines" {
		p.View = "spines"
	}
	p.Group = q.Get("g")
	return p
}

type libGroup struct {
	ID, Label string
	Order     float64
	Books     []*libBook
}

func hueFamily(b *libBook) (string, string, float64) {
	switch {
	case b.Light > 0.82 && b.Sat < 0.35 || b.Light > 0.92:
		return "white", "whites & creams", 9
	case b.Light < 0.16:
		return "black", "blacks", 11
	case b.Sat < 0.16:
		return "grey", "greys", 10
	}
	h := b.Hue
	switch {
	case h < 15 || h >= 340:
		return "red", "reds", 1
	case h < 45:
		return "orange", "oranges & browns", 2
	case h < 70:
		return "yellow", "yellows", 3
	case h < 165:
		return "green", "greens", 4
	case h < 250:
		return "blue", "blues", 5
	case h < 290:
		return "purple", "purples", 6
	default:
		return "pink", "pinks", 7
	}
}

func groupOf(by string, b *libBook) (id, label string, order float64) {
	switch by {
	case "colour":
		return hueFamily(b)
	case "author":
		r := []rune(strings.ToUpper(b.sortName))
		if len(r) == 0 || b.Author == "" || !unicode.IsLetter(r[0]) {
			return "unknown", "unknown author", 99
		}
		return strings.ToLower(string(r[0])), string(r[0]), float64(r[0])
	case "decade":
		y := b.FirstYear
		if y == 0 || (b.Year > 0 && y > b.Year) {
			y = b.Year
		}
		switch {
		case y == 0:
			return "unknown", "year unknown", 9999
		case y < 1900:
			return "before-1900", "before 1900", 1899
		case y < 1950:
			return "1900-1949", "1900–1949", 1900
		default:
			d := y / 10 * 10
			return fmt.Sprint(d), fmt.Sprintf("%ds", d), float64(d)
		}
	case "lang":
		l := libLangs[b.Lang]
		if l == "" {
			l = b.Lang
		}
		if l == "" {
			return "unknown", "language unknown", 99
		}
		order := 50.0
		switch b.Lang {
		case "es":
			order = 1
		case "en":
			order = 2
		}
		return b.Lang, l, order
	case "size":
		p := b.Pages
		switch {
		case p == 0:
			return "unknown", "pages unknown", 9
		case p < 150:
			return "thin", "thin · under 150 pages", 1
		case p < 300:
			return "slim", "slim · 150–300", 2
		case p < 500:
			return "medium", "medium · 300–500", 3
		case p < 800:
			return "thick", "thick · 500–800", 4
		default:
			return "doorstop", "doorstops · 800+", 5
		}
	}
	for i, s := range libShelves {
		if s.ID == b.Shelf {
			return s.ID, s.Label, float64(i)
		}
	}
	return "unsorted", "unsorted", 98
}

func (lib *library) groups(by string) []*libGroup {
	m := map[string]*libGroup{}
	var out []*libGroup
	for _, b := range lib.Books {
		id, label, order := groupOf(by, b)
		g := m[id]
		if g == nil {
			g = &libGroup{ID: id, Label: label, Order: order}
			m[id] = g
			out = append(out, g)
		}
		g.Books = append(g.Books, b)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	for _, g := range out {
		sortBooks(g.Books, by)
	}
	return out
}

func sortBooks(bs []*libBook, by string) {
	sort.SliceStable(bs, func(i, j int) bool {
		a, b := bs[i], bs[j]
		if by == "colour" {
			if math.Abs(a.Hue-b.Hue) > 4 {
				return a.Hue < b.Hue
			}
			return a.Light > b.Light
		}
		if by == "size" && a.Pages != b.Pages {
			return a.Pages < b.Pages
		}
		if a.sortName != b.sortName {
			return a.sortName < b.sortName
		}
		if ya, yb := a.displayYear(), b.displayYear(); ya != yb && ya > 0 && yb > 0 {
			return ya < yb // same author: series in reading order
		}
		return foldKey(a.Title) < foldKey(b.Title)
	})
}

func (b *libBook) path() string { return "/archives/library/" + b.Key + "/" }

func (b *libBook) displayYear() int {
	if b.FirstYear > 0 && (b.Year == 0 || b.FirstYear < b.Year) {
		return b.FirstYear
	}
	return b.Year
}

// item is what a template needs to draw one book as a cover or a spine.
func (b *libBook) item() map[string]any {
	w := 20
	if b.Pages > 0 {
		w = 11 + b.Pages/26
	}
	if w > 54 {
		w = 54
	}
	h := 118 + int(hashOf(b.Key+"h")%40)
	coverW := 100
	if b.HasCover && b.ThumbH > 0 {
		coverW = b.ThumbW * 150 / b.ThumbH
	}
	main, sub := b.splitTitle()
	surname := ""
	if ws := strings.Fields(strings.Split(b.Author, " & ")[0]); len(ws) > 0 {
		surname = ws[len(ws)-1]
	}
	return map[string]any{
		"Key":        b.Key,
		"Main":       main,
		"Sub":        sub,
		"Year":       yearString(b.displayYear()),
		"Lang":       langOr(b.Lang),
		"Surname":    surname,
		"ShowAuthor": w >= 24 && surname != "",
		"CapStyle":   fmt.Sprintf("width:%dpx", coverW),
		"Title":      b.Title,
		"Author":     b.Author,
		"Href":       b.path(),
		"Label":      strings.TrimSpace(b.Title + " — " + b.Author),
		"HasCover":   b.HasCover,
		"Thumb":      "/archives/library/covers/" + b.Key + ".jpg",
		"CoverW":     coverW,
		"GenStyle":   fmt.Sprintf("background-color:%s;color:%s", b.Color, b.Ink),
		"SpineStyle": fmt.Sprintf("width:%dpx;height:%dpx;background-color:%s;color:%s", w, h, b.Color, b.Ink),
		"MiniStyle":  fmt.Sprintf("width:%dpx;height:%dpx;background-color:%s", (w+2)/3, h*2/5, b.Color),
	}
}

// splitTitle turns "Debt: The First 5,000 Years" into main + subtitle; a fetched
// subtitle wins when the title itself has none.
func (b *libBook) splitTitle() (string, string) {
	main, sub := b.Title, b.Subtitle
	if i := strings.Index(main, ": "); i > 0 {
		main, sub = strings.TrimSpace(main[:i]), strings.TrimSpace(main[i+2:])
	}
	if sub == "" {
		if i := strings.Index(main, " - "); i > 0 {
			main, sub = strings.TrimSpace(main[:i]), strings.TrimSpace(main[i+3:])
		}
	}
	return niceCase(main), niceCase(sub)
}

// niceCase fixes catalogue casing: ALL CAPS → Title Case, and a lowercase first letter.
func niceCase(t string) string {
	if t == "" {
		return t
	}
	if strings.ToUpper(t) == t && strings.ToLower(t) != t {
		ws := strings.Fields(strings.ToLower(t))
		for i, w := range ws {
			if i == 0 || len([]rune(w)) > 3 {
				r := []rune(w)
				r[0] = unicode.ToUpper(r[0])
				ws[i] = string(r)
			}
		}
		return strings.Join(ws, " ")
	}
	r := []rune(t)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func itemsOf(bs []*libBook) []any {
	out := make([]any, len(bs))
	for i, b := range bs {
		out[i] = b.item()
	}
	return out
}

func buildLibraryData(lib *library, q url.Values) map[string]any {
	p := parseLibParams(q)
	var groupings, views []any
	for _, g := range libGroupings {
		id := g.ID
		groupings = append(groupings, map[string]any{"Label": g.Label, "Href": p.url(func(c *libParams) { c.By, c.Group = id, "" }), "Active": p.By == id})
	}
	for _, v := range []string{"covers", "spines"} {
		v := v
		views = append(views, map[string]any{"Label": v, "Href": p.url(func(c *libParams) { c.View = v }), "Active": p.View == v})
	}
	var shelves []any
	focus := ""
	all := lib.groups(p.By)
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
			"Count": plural(len(g.Books), "book"),
			"Href":  p.url(func(c *libParams) { c.Group = id }),
			"Items": itemsOf(g.Books),
		})
	}
	n := len(lib.Books)
	stats := plural(n, "book") + " · " + plural(len(all), groupNoun(p.By))
	stats = strings.Replace(stats, "shelfs", "shelves", 1)
	title := "library"
	if focus != "" {
		title = "library · " + focus
	}
	return map[string]any{
		"count":     n,
		"stats":     stats,
		"groupings": groupings,
		"views":     views,
		"spines":    p.View == "spines",
		"shelves":   shelves,
		"hasFocus":  focus != "",
		"focus":     focus,
		"allHref":   p.url(func(c *libParams) { c.Group = "" }),
		"pageTitle": title + " | cptx",
		"shareURL":  siteURL + p.url(nil),
		"empty":     n == 0,
		"canonical": siteURL + "/archives/library/",
	}
}

func groupNoun(by string) string {
	switch by {
	case "shelf":
		return "shelf"
	case "colour":
		return "colour"
	case "author":
		return "letter"
	case "decade":
		return "period"
	case "lang":
		return "language"
	}
	return "size"
}

func buildBookData(lib *library, b *libBook) map[string]any {
	_, shelfLabel, _ := groupOf("shelf", b)
	var neighbours []*libBook
	for _, g := range lib.groups("shelf") {
		if g.ID != b.Shelf {
			continue
		}
		for i, x := range g.Books {
			if x == b {
				lo, hi := i-4, i+5
				if lo < 0 {
					lo = 0
				}
				if hi > len(g.Books) {
					hi = len(g.Books)
				}
				for _, y := range g.Books[lo:hi] {
					if y != b {
						neighbours = append(neighbours, y)
					}
				}
			}
		}
	}
	var facts []any
	add := func(label, value string) {
		if value != "" && value != "0" {
			facts = append(facts, map[string]any{"Label": label, "Value": value})
		}
	}
	add("first published", yearString(b.FirstYear))
	add("this edition", strings.TrimSpace(yearString(b.Year)+" · "+b.Publisher))
	if b.Pages > 0 {
		add("pages", fmt.Sprint(b.Pages))
	}
	add("language", libLangs[b.Lang])
	add("ISBN", b.ISBN13)
	var subjects []any
	seen := map[string]bool{}
	for _, s := range append(append([]string{}, b.Categories...), b.Subjects...) {
		k := strings.ToLower(s)
		if seen[k] || len(subjects) >= 10 || strings.Contains(k, "nyt:") || strings.Contains(k, "=") {
			continue
		}
		seen[k] = true
		subjects = append(subjects, s)
	}
	var links []any
	if b.OpenLibrary != "" {
		links = append(links, map[string]any{"Label": "Open Library", "Href": b.OpenLibrary})
	}
	if b.Google != "" {
		links = append(links, map[string]any{"Label": "Google Books", "Href": b.Google})
	}
	cw, ch := 220, 330
	if b.HasCover && b.ThumbH > 0 {
		ch = 330
		cw = b.ThumbW * ch / b.ThumbH
	}
	bookMain, bookSub := b.splitTitle()
	desc := b.Title
	if b.Author != "" {
		desc += " by " + b.Author
	}
	return map[string]any{
		"book":          b.item(),
		"title":         bookMain,
		"subtitle":      bookSub,
		"author":        b.Author,
		"series":        b.Series,
		"year":          yearString(b.displayYear()),
		"large":         "/archives/library/covers/" + b.Key + "-l.jpg",
		"coverW":        cw,
		"coverH":        ch,
		"facts":         facts,
		"subjects":      subjects,
		"hasSubjects":   len(subjects) > 0,
		"links":         links,
		"hasLinks":      len(links) > 0,
		"shelfLabel":    shelfLabel,
		"shelfHref":     libParams{By: "shelf", View: "covers", Group: b.Shelf}.url(nil),
		"neighbours":    itemsOf(neighbours),
		"hasNeighbours": len(neighbours) > 0,
		"mentions":      postListMaps(b.Mentions),
		"hasMentions":   len(b.Mentions) > 0,
		"buy":           b.Buy,
		"own":           b.Own,
		"pageTitle":     b.Title + " | library | cptx",
		"description":   desc + " — one of the books Merzeede and I own.",
		"canonical":     siteURL + b.path(),
		"lang":          langOr(b.Lang),
	}
}

func langOr(l string) string {
	if l == "" {
		return "en"
	}
	return l
}

func yearString(y int) string {
	if y == 0 {
		return ""
	}
	return fmt.Sprint(y)
}

func buildArchivesData(lib *library) map[string]any {
	var mine []any
	var strip []*libBook
	for _, b := range lib.Books {
		if b.Own {
			it := b.item()
			it["Buy"] = b.Buy
			it["Year"] = yearString(b.Year)
			mine = append(mine, it)
		}
	}
	// a little strip of spines for the library card, in colour order
	strip = append(strip, lib.Books...)
	sortBooks(strip, "colour")
	if len(strip) > 60 {
		step := float64(len(strip)) / 60
		var s []*libBook
		for i := 0.0; int(i) < len(strip) && len(s) < 60; i += step {
			s = append(s, strip[int(i)])
		}
		strip = s
	}
	sort.SliceStable(mine, func(i, j int) bool {
		return mine[i].(map[string]any)["Year"].(string) < mine[j].(map[string]any)["Year"].(string)
	})
	return map[string]any{
		"count":   len(lib.Books),
		"mine":    mine,
		"hasMine": len(mine) > 0,
		"strip":   itemsOf(strip),
	}
}

// registerLibrary wires /archives/ into the mux.
func registerLibrary(mux *http.ServeMux, render func(http.ResponseWriter, string, map[string]any), notFound http.HandlerFunc, lib *library) {
	mux.HandleFunc("/archives/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/archives/":
			render(w, "ArchivesPage", buildArchivesData(lib))
		case path == "/archives/library/":
			render(w, "LibraryPage", buildLibraryData(lib, r.URL.Query()))
		case strings.HasPrefix(path, "/archives/library/covers/"):
			serveCover(w, r, lib, strings.TrimPrefix(path, "/archives/library/covers/"))
		case strings.HasPrefix(path, "/archives/library/"):
			key := strings.Trim(strings.TrimPrefix(path, "/archives/library/"), "/")
			b := lib.ByKey[key]
			if b == nil || !strings.HasSuffix(path, "/") || strings.Contains(key, "/") {
				notFound(w, r)
				return
			}
			render(w, "BookPage", buildBookData(lib, b))
		default:
			notFound(w, r)
		}
	})
}

func serveCover(w http.ResponseWriter, r *http.Request, lib *library, name string) {
	key := strings.TrimSuffix(name, ".jpg")
	sub := "thumbs"
	if strings.HasSuffix(key, "-l") {
		key, sub = strings.TrimSuffix(key, "-l"), "covers"
	}
	b := lib.ByKey[key]
	if b == nil || !b.HasCover || !strings.HasSuffix(name, ".jpg") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeFile(w, r, filepath.Join(lib.Dir, sub, key+".jpg"))
}

// writeLibraryPages writes /archives/, the default shelf and every book page into dist/
// so urls.txt, the sitemap and check.sh cover them.
func writeLibraryPages(engine *htmlc.Engine, distDir string, lib *library) error {
	pages := map[string]struct {
		comp string
		data map[string]any
	}{
		"archives":         {"ArchivesPage", buildArchivesData(lib)},
		"archives/library": {"LibraryPage", buildLibraryData(lib, url.Values{})},
	}
	for _, b := range lib.Books {
		pages["archives/library/"+b.Key] = struct {
			comp string
			data map[string]any
		}{"BookPage", buildBookData(lib, b)}
	}
	for dir, pg := range pages {
		html, err := engine.RenderPageString(pg.comp, pg.data)
		if err != nil {
			return fmt.Errorf("%s %s: %w", pg.comp, dir, err)
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
