package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkbench(t *testing.T) {
	ps, err := loadWorkbench(workbenchDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) == 0 || ps[0].Order > ps[len(ps)-1].Order {
		t.Fatalf("projects not loaded in order: %+v", ps)
	}
	seen := map[string]bool{}
	for i, p := range ps {
		if seen[p.Slug] {
			t.Errorf("duplicate slug %s", p.Slug)
		}
		seen[p.Slug] = true
		d := buildProjectData(ps, i)
		if d["Href"] != "/pockets/workbench/"+p.Slug+"/" || d["StatusLabel"] == "" {
			t.Errorf("%s: %v", p.Slug, d)
		}
	}
	dir := t.TempDir()
	for name, src := range map[string]string{
		"bad status":  "---\ntitle: X\nstatus: broken\nhardware: a\nsoftware: b\n---\n\n## guide\n\n## notes\n",
		"no hardware": "---\ntitle: X\nstatus: running\nsoftware: b\n---\n\n## guide\n\n## notes\n",
		"no notes":    "---\ntitle: X\nstatus: running\nhardware: a\nsoftware: b\n---\n\n## guide\n",
		"wrong order": "---\ntitle: X\nstatus: running\nhardware: a\nsoftware: b\n---\n\n## notes\n\n## guide\n",
	} {
		os.WriteFile(filepath.Join(dir, "x.md"), []byte(src), 0o644)
		if _, err := loadWorkbench(dir); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
