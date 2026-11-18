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
	s := newStatsStore(dir, k)
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	// raw loads: scrapers included, never shown
	s.countView("/", "10.1.1.1", "Mozilla", now)
	s.countView("/croquis/1/", "10.1.1.1", "Mozilla", now)
	s.countView("/", "8.8.8.8", "Mozilla", now)
	s.countView("/", "9.9.9.9", "Mozilla", now) // a scraper: loads the HTML, never says hello
	// people: the hello beacon
	if !s.countHello("/", "10.1.1.1", "Mozilla", now) || s.countHello("/", "10.1.1.1", "Mozilla", now) {
		t.Error("a page counts once per person per day")
	}
	s.countHello("/croquis/1/", "10.1.1.1", "Mozilla", now)
	s.countHello("/", "8.8.8.8", "Mozilla", now)
	if s.countHello("/never-served/", "8.8.8.8", "Mozilla", now) {
		t.Error("hello for a page the server never served")
	}
	ds := s.days["2026-11-15"]
	if ds.Visitors != 2 || ds.Visits["/"] != 2 || ds.Visits["/croquis/1/"] != 1 || ds.Countries["FI"] != 1 || ds.Countries["ZZ"] != 1 {
		t.Errorf("day = %+v", ds)
	}
	if ds.Raw.Hits != 4 || ds.Raw.Visitors != 3 {
		t.Errorf("raw = %+v", ds.Raw)
	}
	s.save()
	s2 := newStatsStore(dir, nil)
	if s2.visitsOf("/croquis/1/") != 1 || s2.days["2026-11-15"].Visitors != 2 {
		t.Errorf("reload: %+v", s2.days)
	}
	d := buildReadersData(s2, []addonPost{{Title: "One", Path: "/croquis/1/", Day: "2026-10-01"}}, now)
	if d["visits"] != 3 || !strings.Contains(d["svg"].(string), "Finland") {
		t.Errorf("readers data: %v", d["visits"])
	}
	// a visit to a post is a read: most read lists it, other pages don't
	if mr := d["mostRead"].([]any); len(mr) != 1 || mr[0].(map[string]any)["Reads"] != "read once" {
		t.Errorf("most read: %v", d["mostRead"])
	}
	if mv := d["mostVisited"].([]any); len(mv) != 1 || mv[0].(map[string]any)["Label"] != "home" {
		t.Errorf("other pages: %v", d["mostVisited"])
	}
	// raw: 4 loads, 3 raw visitors; people: 3 visits, 2 visitors → 1 machine load, 1 machine visitor
	if d["machineHits"] != 1 || d["machineVisitors"] != 1 || d["machinePct"] != 25 || d["hasMachines"] != true {
		t.Errorf("machines: %v %v %v", d["machineHits"], d["machineVisitors"], d["machinePct"])
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

// Days saved before AD-07 counted every HTML load as a visit: they move to Raw.
func TestStatsOldDays(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "2026-11.json"), []byte(`{"2026-11-15":{"visitors":5,"visits":{"/":7,"/a/":3},"reads":{"/croquis/1/":2},"countries":{"US":5}}}`), 0o644)
	s := newStatsStore(dir, nil)
	ds := s.days["2026-11-15"]
	if ds.Visitors != 0 || len(ds.Visits) != 0 || len(ds.Countries) != 0 || ds.Raw.Hits != 10 || ds.Raw.Visitors != 5 || ds.V != statsVersion {
		t.Errorf("old day = %+v", ds)
	}
	if s.visitsOf("/croquis/1/") != 0 || s.days["2026-11-15"].Reads["/croquis/1/"] != 2 {
		t.Error("old timed reads are kept on disk but no longer shown")
	}
	s.save()
	if s2 := newStatsStore(dir, nil); s2.days["2026-11-15"].Raw.Hits != 10 {
		t.Errorf("migrated once: %+v", s2.days["2026-11-15"].Raw)
	}
}

// The list shows the top 10 countries; the map lights all of them.
func TestReadersTop10(t *testing.T) {
	s := newStatsStore("", nil)
	ds := newDay()
	ds.Visitors = 1
	for i, cc := range []string{"EE", "FI", "ES", "DE", "FR", "IT", "PT", "SE", "NO", "DK", "PL", "LV"} {
		ds.Countries[cc] = 20 - i
	}
	s.days["2026-11-18"] = ds
	d := buildReadersData(s, nil, time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC))
	if n := len(d["countries"].([]any)); n != 10 || d["moreCountries"] != 2 || d["countryCount"] != 12 {
		t.Errorf("list %d, more %v, count %v", n, d["moreCountries"], d["countryCount"])
	}
	if !strings.Contains(d["svg"].(string), "Latvia") {
		t.Error("the 12th country still lights the map")
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
