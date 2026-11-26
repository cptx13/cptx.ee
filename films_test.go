package main

import (
	"image"
	"net/url"
	"testing"
)

func TestFrameOfIsTwoByThree(t *testing.T) {
	for _, sz := range [][2]int{{300, 450}, {300, 444}, {320, 450}, {400, 400}, {600, 300}, {200, 800}} {
		b := frameOf(image.NewRGBA(image.Rect(0, 0, sz[0], sz[1]))).Bounds()
		if b.Dx()*filmH != b.Dy()*filmW {
			t.Errorf("%v → %dx%d, want 2:3", sz, b.Dx(), b.Dy())
		}
	}
}

func TestFilmGroups(t *testing.T) {
	lib := &library{Films: &filmShelf{ByKey: map[string]*film{}}}
	for _, f := range []*film{
		{Key: "a", Title: "The Pianist", Year: 2002, Shelf: "drama", Director: "Roman Polanski"},
		{Key: "b", Title: "Alien", Year: 1979, Shelf: "fantasy-sf", Director: "Ridley Scott"},
		{Key: "c", Title: "Memento", Year: 2000, Shelf: "crime"},
	} {
		f.sortName = foldKey(articleless(f.Title))
		lib.Films.Films = append(lib.Films.Films, f)
		lib.Films.ByKey[f.Key] = f
	}
	if id, _, _ := filmGroupOf("title", lib.Films.ByKey["a"]); id != "p" {
		t.Errorf("The Pianist files under %q, want p", id)
	}
	if id, _, _ := filmGroupOf("director", lib.Films.ByKey["a"]); id != "p" {
		t.Errorf("Polanski files under %q, want p", id)
	}
	if id, _, _ := filmGroupOf("director", lib.Films.ByKey["c"]); id != "unknown" {
		t.Errorf("no director files under %q, want unknown", id)
	}
	d := buildFilmsData(lib, url.Values{"by": {"decade"}})
	if d["stats"] != "3 films · 3 decades" {
		t.Errorf("stats = %v", d["stats"])
	}
	fd := buildFilmData(lib, lib.Films.ByKey["b"])
	if fd["canonical"] != siteURL+"/archives/films/b/" {
		t.Errorf("canonical = %v", fd["canonical"])
	}
}
