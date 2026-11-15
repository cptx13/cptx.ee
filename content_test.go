package main

// B-12: tests for URL/ID generation, frontmatter parsing and small helpers.

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestComputePostIDs(t *testing.T) {
	a := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-03"), FilePath: "content/posts/croquis/b-second.md", Slug: "b-second"}
	b := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-03"), FilePath: "content/posts/croquis/a-first.md", Slug: "a-first"}
	c := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-04"), FilePath: "content/posts/croquis/c.md", Slug: "c"}
	r := &Post{Section: "recuerdos", Date: day("2026-10-03"), FilePath: "content/posts/recuerdos/x.md", Slug: "x"}
	pk := &Post{Section: "pockets", Date: day("2026-10-03"), FilePath: "content/pockets/p.md", Path: "/pockets/p/"}

	computePostIDs([]*Post{a, b, c, r, pk})

	cases := []struct {
		p        *Post
		id, path string
		oldPath  string
	}{
		{b, "202610031", "/croquis/202610031/", "/posts/croquis/a-first/"}, // same day: sorted by filename
		{a, "202610032", "/croquis/202610032/", "/posts/croquis/b-second/"},
		{c, "202610041", "/croquis/202610041/", "/posts/croquis/c/"},
		{r, "202610031", "/recuerdos/202610031/", "/posts/recuerdos/x/"}, // own category, no categories → section
	}
	for _, tc := range cases {
		if tc.p.ID != tc.id || tc.p.Path != tc.path || tc.p.OldPath != tc.oldPath {
			t.Errorf("%s: got id=%q path=%q old=%q, want %q %q %q", tc.p.FilePath, tc.p.ID, tc.p.Path, tc.p.OldPath, tc.id, tc.path, tc.oldPath)
		}
	}
	if pk.ID != "" || pk.Path != "/pockets/p/" {
		t.Errorf("pockets post should be untouched, got id=%q path=%q", pk.ID, pk.Path)
	}
}

func TestParseFrontmatter(t *testing.T) {
	fm, body := parseFrontmatter("\n---\ntitle: \"Hola: mundo\"\ndate: 2026-10-03\ntags: [a, 'b', \"c\"]\ndraft: yes\nnot a field\n---\nBody text\n")
	if fm["title"] != "Hola: mundo" {
		t.Errorf("title = %q", fm["title"])
	}
	if fm["date"] != "2026-10-03" {
		t.Errorf("date = %q", fm["date"])
	}
	if body != "\nBody text\n" {
		t.Errorf("body = %q", body)
	}
	if got := fmStringSlice(fm, "tags"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("tags = %#v", got)
	}
	if !fmBool(fm, "draft") || fmBool(fm, "missing") {
		t.Error("fmBool")
	}
	if fmStringSlice(fm, "missing") != nil || fmStringSlice(map[string]string{"x": "[]"}, "x") != nil {
		t.Error("empty slices should be nil")
	}

	fm, body = parseFrontmatter("no frontmatter here")
	if len(fm) != 0 || body != "no frontmatter here" {
		t.Errorf("no frontmatter: fm=%v body=%q", fm, body)
	}
}

func TestParsePostPaths(t *testing.T) {
	src := "---\ntitle: T\ndate: 2026-10-03T10:00:00+02:00\ncategories: [croquis]\n---\nUn texto.\n"
	cases := []struct {
		file, section, path string
	}{
		{"content/posts/croquis/Mi texto.md", "croquis", "/posts/croquis/mi-texto/"},
		{"content/pockets/weekendreads.md", "pockets", "/pockets/weekendreads/"},
		{"content/manifesto.md", "", "/manifesto/"},
		{"content/contact.md", "", "/contact/"},
	}
	for _, tc := range cases {
		p, err := parsePost(src, tc.file, "content")
		if err != nil {
			t.Fatalf("%s: %v", tc.file, err)
		}
		if p.Section != tc.section || p.Path != tc.path {
			t.Errorf("%s: section=%q path=%q, want %q %q", tc.file, p.Section, p.Path, tc.section, tc.path)
		}
		if p.DateFormatted != "2026.10.03" {
			t.Errorf("%s: date %q", tc.file, p.DateFormatted)
		}
	}
	if _, err := parsePost("just text", "content/x.md", "content"); err == nil {
		t.Error("a file with neither title nor date should be rejected")
	}
}

func TestCacheControlFor(t *testing.T) {
	cases := map[string]string{
		"/fonts/jetbrains-mono/JetBrainsMono-Regular.woff2": "public, max-age=2592000",
		"/css/style.css":        "public, max-age=3600",
		"/js/photos.js":         "public, max-age=3600",
		"/favicon.ico":          "public, max-age=604800",
		"/photos/tallinn/a.JPG": "public, max-age=604800",
		"/photos/tallinn/":      "",
		"/croquis/202610031/":   "",
	}
	for path, want := range cases {
		if got := cacheControlFor(path); got != want {
			t.Errorf("%s: %q, want %q", path, got, want)
		}
	}
}

func TestWordCountsStopWords(t *testing.T) {
	got := wordCounts("Hoy tal vez el café, uno y otro café. The coffee is good")
	want := map[string]int{"café": 2, "coffee": 1, "good": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wordCounts = %v, want %v", got, want)
	}
}

func TestSortHelpers(t *testing.T) {
	if got := foldKey("Álvarez Ñu"); got != "alvarez nu" {
		t.Errorf("foldKey = %q", got)
	}
	if got := authorSortName("Gabriel García Márquez, Otro"); got != "marquez gabriel garcia marquez" {
		t.Errorf("authorSortName = %q", got)
	}
	if authorSortName("") != "~" {
		t.Error("empty author should sort last")
	}
	if plural(1, "book") != "1 book" || plural(3, "book") != "3 books" || plural(0, "book") != "0 books" {
		t.Error("plural")
	}
}

func TestComputePostIDsPinned(t *testing.T) {
	// a pins the slot b would get by filename order; b and c fill the free ones.
	a := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-05"), FilePath: "z-moved.md", Slug: "z", FixedID: "202610031"}
	b := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-03"), FilePath: "b.md", Slug: "b"}
	c := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-03"), FilePath: "c.md", Slug: "c"}
	if err := computePostIDs([]*Post{b, c, a}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		p    *Post
		path string
	}{{a, "/croquis/202610031/"}, {b, "/croquis/202610032/"}, {c, "/croquis/202610033/"}} {
		if tc.p.Path != tc.path {
			t.Errorf("%s: %s, want %s", tc.p.FilePath, tc.p.Path, tc.path)
		}
	}
	if a.OldPath != "/posts/croquis/z/" {
		t.Errorf("pinned post keeps its legacy path, got %q", a.OldPath)
	}

	dup := &Post{Section: "croquis", Categories: []string{"croquis"}, Date: day("2026-10-03"), FilePath: "d.md", FixedID: "202610031"}
	if err := computePostIDs([]*Post{a, dup}); err == nil {
		t.Error("two posts pinning the same ID should fail")
	}
	bad := &Post{Section: "croquis", Date: day("2026-10-03"), FilePath: "e.md", FixedID: "2026-10-03"}
	if err := computePostIDs([]*Post{bad}); err == nil {
		t.Error("a non-numeric id should fail")
	}
}

func TestParsePostFixedID(t *testing.T) {
	p, err := parsePost("---\ntitle: T\ndate: 2026-10-03\nid: 202610031\ncategories: [croquis]\n---\nx\n", "content/posts/croquis/T.md", "content")
	if err != nil || p.FixedID != "202610031" {
		t.Errorf("FixedID = %q, err %v", p.FixedID, err)
	}
}

func TestCacheControlVersioned(t *testing.T) {
	if got := cacheControlForRequest("/css/style.css", "adb4dda"); got != "public, max-age=31536000, immutable" {
		t.Errorf("versioned css: %q", got)
	}
	if got := cacheControlForRequest("/css/style.css", ""); got != "public, max-age=3600" {
		t.Errorf("unversioned css: %q", got)
	}
	if got := cacheControlForRequest("/croquis/202610071/", "x"); got != "" {
		t.Errorf("pages are never cached long: %q", got)
	}
}

func TestLegalDocs(t *testing.T) {
	p, err := parsePost("---\ntitle: Privacy\ndate: 2026-11-15\nsummary: What is collected.\n---\nx\n", "content/legaldocs/privacy.md", "content")
	if err != nil || p.Section != "legaldocs" || p.Path != "/legaldocs/privacy/" {
		t.Fatalf("legal doc parsed as section=%q path=%q err=%v", p.Section, p.Path, err)
	}
	other := &Post{Section: "legaldocs", Slug: "zz", Title: "Zz", Path: "/legaldocs/zz/"}
	cards := buildLegalData([]*Post{other, p, {Section: "croquis", Slug: "privacy"}})
	if len(cards) != 2 || cards[0].(map[string]any)["Href"] != "/legaldocs/privacy/" {
		t.Errorf("cards = %v", cards)
	}
}

func TestNewsletterIssues(t *testing.T) {
	p, err := parsePost("---\ntitle: Autumn notes\ndate: 2026-11-16\n---\n## what changed\n", "content/newsletter/007.md", "content")
	if err != nil || p.Section != "newsletter" || p.Path != "/newsletter/7/" || p.Slug != "7" {
		t.Fatalf("issue parsed as section=%q path=%q err=%v", p.Section, p.Path, err)
	}
	old := &Post{Section: "newsletter", Slug: "6", Title: "Earlier", Path: "/newsletter/6/", Date: day("2026-11-09")}
	draft := &Post{Section: "newsletter", Slug: "8", Draft: true}
	d := newsletterData([]*Post{old, p, draft, {Section: "croquis", Slug: "9"}}, 1)
	issues := d["issues"].([]any)
	if len(issues) != 1 || issues[0].(map[string]any)["No"] != "7" || d["issueCount"] != 2 || d["hasMore"] != true {
		t.Errorf("newsletter data = %v", d)
	}
	if w := issues[0].(map[string]any)["Week"]; w != "week 47" {
		t.Errorf("week = %v", w)
	}
}

func TestSearch(t *testing.T) {
	idx := []addonPost{
		{Title: "Café de mañana", Path: "/croquis/1/", Day: "2026-10-07", Section: "croquis"},
		{Title: "Lluvia", Path: "/recuerdos/2/", Day: "2023-01-02", Section: "recuerdos"},
	}
	texts := map[string]string{
		"/croquis/1/":   "Un café en Tallinn, temprano. <b>no html</b>",
		"/recuerdos/2/": "Llovía sobre Tallinn todo el día.",
	}
	s := buildSearchIndex(idx, texts, &library{Books: []*libBook{{Key: "k", Title: "Cien años de soledad", Author: "Gabriel García Márquez"}}})
	posts, books, total := s.search("tallinn", 10)
	if total != 2 || len(books) != 0 {
		t.Fatalf("tallinn: %d posts, %d books", total, len(books))
	}
	posts, _, total = s.search("cafe tallinn", 10)
	if total != 1 || posts[0].(map[string]any)["Path"] != "/croquis/1/" {
		t.Errorf("accents and all terms: %v", posts)
	}
	if snip := posts[0].(map[string]any)["Snippet"].(string); !strings.Contains(snip, "<mark>café</mark>") || strings.Contains(snip, "<b>") {
		t.Errorf("snippet = %q", snip)
	}
	_, _, total = s.search("llov", 10)
	if total != 1 {
		t.Errorf("prefix match: %d", total)
	}
	_, books, _ = s.search("garcia marquez", 10)
	if len(books) != 1 {
		t.Errorf("library: %v", books)
	}
	if d := searchData(s, "  "); d["hasQuery"] != false || d["none"] != false {
		t.Errorf("empty query: %v", d)
	}
}

func TestIssues(t *testing.T) {
	how := &Post{Section: "newsletter", Slug: "how-it-works", Title: "How the newsletter works"}
	one := &Post{Section: "newsletter", Slug: "1", Title: "A week of changes", Path: "/newsletter/1/", Date: day("2026-11-15")}
	if isIssue(how) || !isIssue(one) || isIssue(&Post{Section: "croquis", Slug: "1"}) {
		t.Errorf("isIssue wrong")
	}
	if issueLabel("1") != "issue 01" || issueLabel("12") != "issue 12" || issueLabel("x") != "" {
		t.Errorf("issueLabel wrong")
	}
	if pageTitleFor(one) != "issue 01: A week of changes | cptx" || pageTitleFor(how) != "How the newsletter works | cptx" {
		t.Errorf("pageTitleFor = %q", pageTitleFor(one))
	}
	words := map[string]int{"2026-11-10": 900, "2026-11-12": 50}
	a, b := coverArt(one, 1, words), coverArt(one, 1, words)
	if a != b || !strings.HasPrefix(a, "<svg") || !strings.HasSuffix(a, "</svg>") || !strings.Contains(a, ">01</text>") {
		t.Errorf("coverArt not stable or malformed: %s", a)
	}
	if strings.Count(a, `y="186`) != 7 {
		t.Errorf("week strip should have 7 days: %s", a)
	}
	d := newsletterData([]*Post{how, one}, 0)
	if d["issueCount"] != 1 || d["issueCountLabel"] != "1 issue" {
		t.Errorf("how-it-works counted as an issue: %v", d)
	}
}
