package main

import (
	"net/url"
	"testing"
	"time"
)

func TestParseReads(t *testing.T) {
	src := "# comment\n\n" +
		"2026-05\tOld one\tA\thttps://a.example/x\t\t\n" +
		"2026-11-20\tNew one\t\thttps://www.b.example/y/\t2019-01-16\tgood\n"
	rs, err := parseReads(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 || rs[0].Title != "New one" || rs[1].Month() != "2026-05" {
		t.Fatalf("got %+v", rs)
	}
	for _, bad := range []string{
		"2026/11/20\tT\t\thttps://x.example/\t\t\n",
		"2026-11-20\t\t\thttps://x.example/\t\t\n",
		"2026-11-20\tT\t\tnot a url\t\t\n",
		"2026-13-40\tT\t\thttps://x.example/\t\t\n",
	} {
		if _, err := parseReads(bad); err == nil {
			t.Errorf("no error for %q", bad)
		}
	}
}

func TestBuildReadsData(t *testing.T) {
	rs, _ := parseReads("2026-11-20\tKept\tA\thttps://www.b.example/y/\t\t\n2026-10-01\tOlder\t\thttps://c.example/\t\t\n")
	cache := &FeedCache{LastRefreshed: time.Now(), Items: []FeedItem{
		{Title: "Kept", Link: "https://b.example/y", Source: "B"},
		{Title: "Other", Link: "https://b.example/z", Source: "B"},
	}}
	d := buildReadsData(rs, cache, url.Values{})
	if d["view"] != "kept" || len(d["months"].([]any)) != 2 {
		t.Fatalf("kept view: %v", d)
	}
	d = buildReadsData(rs, cache, url.Values{"view": {"latest"}})
	latest := d["latest"].([]any)
	if d["view"] != "latest" || !latest[0].(map[string]any)["Kept"].(bool) || latest[1].(map[string]any)["Kept"].(bool) {
		t.Fatalf("latest view: %v", latest)
	}
	if got := readsSummary(rs); got != "2 articles · 2026.11" {
		t.Errorf("summary %q", got)
	}
	if err := func() error { _, err := loadReads(); return err }(); err != nil {
		t.Errorf("data/reads.tsv: %v", err)
	}
}
