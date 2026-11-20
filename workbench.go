// Workbench: /pockets/workbench/ — Carlos's projects as cards, each with its
// own guide page /pockets/workbench/<slug>/. Source: data/workbench/*.md.

package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dhamidi/htmlc"
)

const workbenchDir = "data/workbench"

// Project is one workbench card + guide.
type Project struct {
	Slug, Title, Icon, Status, Summary, Link string
	Kit                                      []string
	Order                                    int
	Body                                     string // rendered HTML
}

var workbenchStatus = map[string]string{
	"running":   "running",
	"tinkering": "tinkering",
	"planned":   "planned",
	"revamp":    "to revamp",
}

// workbenchLoose are links that belong to no project yet.
var workbenchLoose = []struct{ Title, URL string }{
	{"Pop!_OS docs", "https://pop-os.github.io/docs/"},
	{"dArkOS", "https://github.com/christianhaitian/dArkOS/wiki"},
}

var siteProjects []Project

func loadWorkbench(dir string) ([]Project, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	var out []Project
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		fm, body := parseFrontmatter(string(b))
		p := Project{
			Slug: strings.TrimSuffix(filepath.Base(f), ".md"), Title: fm["title"], Icon: fm["icon"],
			Status: fm["status"], Summary: fm["summary"], Link: fm["link"], Body: renderMarkdown(body),
		}
		if p.Title == "" {
			return nil, fmt.Errorf("%s: no title", f)
		}
		if _, ok := workbenchStatus[p.Status]; !ok {
			return nil, fmt.Errorf("%s: status %q (want running, tinkering, planned or revamp)", f, p.Status)
		}
		p.Order, _ = strconv.Atoi(fm["order"])
		for _, k := range strings.Split(fm["kit"], ",") {
			if k = strings.TrimSpace(k); k != "" {
				p.Kit = append(p.Kit, k)
			}
		}
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out, nil
}

func projectCard(p Project) map[string]any {
	return map[string]any{
		"Slug": p.Slug, "Title": p.Title, "Icon": p.Icon, "Summary": p.Summary,
		"Status": p.Status, "StatusLabel": workbenchStatus[p.Status], "Href": "/pockets/workbench/" + p.Slug + "/",
	}
}

func buildWorkbenchData(ps []Project) map[string]any {
	var cards, loose []any
	for _, p := range ps {
		cards = append(cards, projectCard(p))
	}
	for _, l := range workbenchLoose {
		loose = append(loose, map[string]any{"Title": l.Title, "URL": l.URL})
	}
	return map[string]any{"projects": cards, "count": len(ps), "loose": loose}
}

func buildProjectData(ps []Project, i int) map[string]any {
	p := ps[i]
	d := projectCard(p)
	kit := make([]any, 0, len(p.Kit))
	for _, k := range p.Kit {
		kit = append(kit, k)
	}
	d["kit"] = kit
	d["hasKit"] = len(kit) > 0
	d["body"] = p.Body
	d["link"] = p.Link
	d["hasLink"] = p.Link != ""
	d["linkLabel"] = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(p.Link, "https://"), "www."), "/")
	return d
}

// workbenchSummary is the pockets hub card's meta line.
func workbenchSummary(ps []Project) string {
	if len(ps) == 0 {
		return ""
	}
	return strconv.Itoa(len(ps)) + " projects"
}

func registerWorkbench(mux *http.ServeMux, render func(http.ResponseWriter, string, map[string]any), notFound http.HandlerFunc) {
	mux.HandleFunc("/pockets/workbench/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pockets/workbench/" {
			render(w, "WorkbenchPage", buildWorkbenchData(siteProjects))
			return
		}
		for i, p := range siteProjects {
			if r.URL.Path == "/pockets/workbench/"+p.Slug+"/" {
				render(w, "ProjectPage", buildProjectData(siteProjects, i))
				return
			}
		}
		notFound(w, r)
	})
}

// writeWorkbenchPages puts the hub and every guide in dist/ (urls.txt, sitemap, check.sh).
func writeWorkbenchPages(engine *htmlc.Engine, distDir string, ps []Project) error {
	write := func(dir, comp string, data map[string]any) error {
		html, err := engine.RenderPageString(comp, data)
		if err != nil {
			return fmt.Errorf("%s: %w", comp, err)
		}
		out := filepath.Join(distDir, filepath.FromSlash(dir))
		if err := os.MkdirAll(out, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "index.html"), []byte(withDoctype(html)), 0o644)
	}
	if err := write("pockets/workbench", "WorkbenchPage", buildWorkbenchData(ps)); err != nil {
		return err
	}
	for i, p := range ps {
		if err := write("pockets/workbench/"+p.Slug, "ProjectPage", buildProjectData(ps, i)); err != nil {
			return err
		}
	}
	return nil
}
