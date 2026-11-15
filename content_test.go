package main

// B-12: tests for URL/ID generation, frontmatter parsing and small helpers.

import (
	"reflect"
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
