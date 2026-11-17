package main

// A-01…A-04: country lookup, counting rules, persistence, readers page data.

import (
	"compress/gzip"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGeo(t *testing.T) {
	g, err := parseGeoCSV(strings.NewReader("1.0.0.0,1.0.0.255,AU\n1.0.1.0,1.0.3.255,CN\n1.0.4.0,1.0.7.255,AU\n2001:200::,2001:200:ffff:ffff:ffff:ffff:ffff:ffff,JP\n"))
	if err != nil {
		t.Fatal(err)
	}
	for ip, want := range map[string]string{"1.0.0.7": "AU", "1.0.2.1": "CN", "1.0.5.9": "AU", "1.0.8.1": "", "2001:200::1": "JP", "::ffff:1.0.2.1": "CN"} {
		if got := g.lookup(netip.MustParseAddr(ip)); got != want {
			t.Errorf("%s = %q want %q", ip, got, want)
		}
	}
}

func TestStats(t *testing.T) {
	g, _ := parseGeoCSV(strings.NewReader("10.0.0.0,10.255.255.255,FI\n"))
	k := &geoKeeper{db: g}
	dir := t.TempDir()
	s := newStatsStore(dir, map[string]bool{"/croquis/1/": true}, k)
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	s.countView("/", "10.1.1.1", "Mozilla", now)
	s.countView("/croquis/1/", "10.1.1.1", "Mozilla", now)
	s.countView("/", "8.8.8.8", "Mozilla", now)
	if !s.countRead("/croquis/1/", "10.1.1.1", "Mozilla", now) || s.countRead("/croquis/1/", "10.1.1.1", "Mozilla", now) {
		t.Error("read should count once per visitor per day")
	}
	if s.countRead("/nope/", "10.1.1.1", "Mozilla", now) {
		t.Error("unknown post")
	}
	ds := s.days["2026-11-15"]
	if ds.Visitors != 2 || ds.Visits["/"] != 2 || ds.Countries["FI"] != 1 || ds.Countries["ZZ"] != 1 {
		t.Errorf("day = %+v", ds)
	}
	s.save()
	s2 := newStatsStore(dir, nil, nil)
	if s2.readsOf("/croquis/1/") != 1 || s2.days["2026-11-15"].Visitors != 2 {
		t.Errorf("reload: %+v", s2.days)
	}
	d := buildReadersData(s2, []addonPost{{Title: "One", Path: "/croquis/1/", Day: "2026-10-01"}}, now)
	if d["visits"] != 3 || d["reads"] != 1 || !strings.Contains(d["svg"].(string), "Finland") {
		t.Errorf("readers data: %v %v", d["visits"], d["reads"])
	}
	if ms := d["months"].([]any); len(ms) != 1 || d["monthsNote"] != "since November 2026" {
		t.Errorf("months before the first data: %d %v", len(ms), d["monthsNote"])
	}
	if ms := buildReadersData(s2, nil, now.AddDate(1, 2, 0))["months"].([]any); len(ms) != 12 {
		t.Errorf("months capped at 12: %d", len(ms))
	}
	if readsLabel(0) != "" || readsLabel(1) != "read once" || readsLabel(7) != "read 7 times" {
		t.Error("labels")
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", "curl/8.0")
	if !isBot(r) {
		t.Error("curl is a bot")
	}
	r.Header.Set("User-Agent", "Mozilla/5.0 (iPhone)")
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if isBot(r) || clientIP(r) != "5.6.7.8" {
		t.Error("ua/ip")
	}
}

func TestLoadGeoGzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.part.gz")
	f, _ := os.Create(path)
	gz := gzip.NewWriter(f)
	gz.Write([]byte("0.0.0.0,0.255.255.255,ZZ\n1.0.0.0,1.0.0.255,AU\n"))
	gz.Close()
	f.Close()
	g, err := loadGeoFile(path)
	if err != nil || g.lookup(netip.MustParseAddr("1.0.0.1")) != "AU" {
		t.Fatalf("gzip load: %v", err)
	}
}

func TestPageLabel(t *testing.T) {
	for in, want := range map[string]string{"/": "home", "/addons/readers/": "addons / readers", "/categories/croquis/": "croquis", "/manifesto/": "manifesto"} {
		if got := pageLabel(in); got != want {
			t.Errorf("%s = %q, want %q", in, got, want)
		}
	}
}
