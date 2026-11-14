package main

import (
	"compress/gzip"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Agosto está lleno de imposibles": "agosto-está-lleno-de-imposibles",
		"¿Y si nunca llegamos a bailar?":  "y-si-nunca-llegamos-a-bailar",
		"761 - Césped húmedo":             "761---césped-húmedo",
		"Café decomposed":                "café-decomposed", // NFD input -> NFC slug
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLegacyParam(t *testing.T) {
	if got := legacyParam("c%C3%A9sped-h%C3%BAmedo"); got != "césped-húmedo" {
		t.Errorf("encoded: %q", got)
	}
	if got := legacyParam("césped"); got != "césped" { // NFD -> NFC
		t.Errorf("nfd: %q", got)
	}
	if got := legacyParam("%zz"); got != "%zz" {
		t.Errorf("bad escape: %q", got)
	}
}

func TestPostLang(t *testing.T) {
	if l := postLang("", "No sé que tendrán las noches de verano para lograr que todo suene a una idea"); l != "es" {
		t.Errorf("spanish -> %s", l)
	}
	if l := postLang("", "I rarely write in English, not due to my lack of interest but rather a faithful commitment"); l != "en" {
		t.Errorf("english -> %s", l)
	}
	if l := postLang("EN", "texto en español"); l != "en" {
		t.Errorf("frontmatter override -> %s", l)
	}
}

func TestExcerpt(t *testing.T) {
	got := excerpt("## Title\n\nSome *bold* text with a [link](http://x) and more words here", 30)
	if strings.ContainsAny(got, "*#[]()") || !strings.HasSuffix(got, "…") || len([]rune(got)) > 31 {
		t.Errorf("excerpt = %q", got)
	}
	if got := excerpt("short", 30); got != "short" {
		t.Errorf("short = %q", got)
	}
}

func TestBuildRSS(t *testing.T) {
	d := func(s string) time.Time { t, _ := time.Parse("2006-01-02", s); return t }
	posts := []*Post{
		{Title: "old", Section: "croquis", Path: "/croquis/201901011/", Date: d("2019-01-01")},
		{Title: "new <&>", Section: "croquis", Path: "/croquis/202601011/", Date: d("2026-01-01")},
		{Title: "rec", Section: "recuerdos", Path: "/recuerdos/202201011/", Date: d("2022-01-01")},
	}
	var doc rssDoc
	if err := xml.Unmarshal(buildRSS(posts, 1), &doc); err != nil {
		t.Fatal(err)
	}
	if n := len(doc.Channel.Items); n != 1 || doc.Channel.Items[0].Title != "new <&>" {
		t.Errorf("items = %+v", doc.Channel.Items)
	}
}

func TestGzipAndHeaders(t *testing.T) {
	h := withGzip(withHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, strings.Repeat("hola ", 200))
	})))
	req := httptest.NewRequest("GET", "/css/style.css", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("not gzipped: %v", rec.Header())
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(zr)
	if len(b) != 1000 {
		t.Errorf("body len %d", len(b))
	}
	for _, k := range []string{"X-Content-Type-Options", "Content-Security-Policy", "Referrer-Policy"} {
		if rec.Header().Get(k) == "" {
			t.Errorf("missing %s", k)
		}
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Errorf("cache-control %q", cc)
	}
	// no gzip without Accept-Encoding
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest("GET", "/", nil))
	if rec2.Header().Get("Content-Encoding") != "" || rec2.Body.Len() != 1000 {
		t.Errorf("unexpected encoding without Accept-Encoding")
	}
}
