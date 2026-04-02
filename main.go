package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/dhamidi/dispatch"
	"github.com/dhamidi/htmlc"
	"github.com/yuin/goldmark"
	"golang.org/x/text/unicode/norm"
)

// Post represents a content entry parsed from markdown files.
type Post struct {
	Title         string
	Author        string
	Date          time.Time
	Draft         bool
	Categories    []string
	Tags          []string
	ShowTags      bool
	TOC           bool
	Summary       string
	Content       template.HTML
	Slug          string
	Section       string
	Path          string
	FilePath      string
	DateFormatted string
}

func main() {
	posts, err := loadAllContent("content")
	if err != nil {
		log.Fatalf("loading content: %v", err)
	}

	// Sort posts by date descending
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	// Filter out drafts
	var published []*Post
	for _, p := range posts {
		if !p.Draft {
			published = append(published, p)
		}
	}

	engine, err := htmlc.New(htmlc.Options{
		ComponentDir: ".",
		FS:           os.DirFS("."),
	})
	if err != nil {
		log.Fatalf("creating htmlc engine: %v", err)
	}

	router := dispatch.New()

	// Group posts by section and category
	postsBySection := map[string][]*Post{}
	postsByCategory := map[string][]*Post{}
	postsByPath := map[string]*Post{}
	pocketPosts := []*Post{}
	// allPostsForPool will be set below after grouping
	var manifestoPost *Post
	var allPosts []*Post // all posts for shuffle pool (posts section only)

	for _, p := range published {
		postsByPath[p.Path] = p
		for _, cat := range p.Categories {
			postsByCategory[cat] = append(postsByCategory[cat], p)
		}
		if p.Section == "recuerdos" || p.Section == "croquis" {
			postsBySection[p.Section] = append(postsBySection[p.Section], p)
			allPosts = append(allPosts, p)
		} else if p.Section == "pockets" {
			pocketPosts = append(pocketPosts, p)
		} else if p.Path == "/manifesto/" {
			manifestoPost = p
		}
	}

	// allPosts for shuffle pool = all posts in the "posts" collection
	allPostsForPool := allPosts

	// Post data as maps for templates
	toPostMap := func(p *Post) map[string]any {
		tags := make([]any, len(p.Tags))
		for i, t := range p.Tags {
			tags[i] = t
		}
		categories := make([]any, len(p.Categories))
		for i, c := range p.Categories {
			categories[i] = c
		}
		return map[string]any{
			"Title":         p.Title,
			"Author":        p.Author,
			"Date":          p.Date,
			"DateFormatted": p.DateFormatted,
			"Content":       string(p.Content),
			"Slug":          p.Slug,
			"Section":       p.Section,
			"Path":          p.Path,
			"Summary":       p.Summary,
			"Tags":          tags,
			"ShowTags":      p.ShowTags,
			"TOC":           p.TOC,
			"Categories":    categories,
		}
	}

	toPostList := func(posts []*Post) []any {
		result := make([]any, len(posts))
		for i, p := range posts {
			result[i] = toPostMap(p)
		}
		return result
	}

	// Generate shuffle pool JSON (just paths)
	shufflePoolPaths := make([]string, len(allPostsForPool))
	for i, p := range allPostsForPool {
		shufflePoolPaths[i] = p.Path
	}
	shufflePoolJSON, _ := json.Marshal(shufflePoolPaths)

	// Home page
	must(router.GET("home", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"posts": toPostList(allPostsForPool),
		}
		renderPage(w, engine, "HomePage", data)
	})))

	// Manifesto
	must(router.GET("manifesto", "/manifesto/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manifestoPost == nil {
			http.NotFound(w, r)
			return
		}
		data := map[string]any{
			"post":     toPostMap(manifestoPost),
			"prevPost": false,
			"nextPost": false,
		}
		renderPage(w, engine, "SinglePage", data)
	})))

	// Category pages
	must(router.GET("category", "/categories/{name}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		name := m.Params["name"]
		catPosts := postsByCategory[name]

		// Check if there's an _index.md for this category
		var contentHTML string
		indexPath := filepath.Join("content", "categories", name, "_index.md")
		if data, err := os.ReadFile(indexPath); err == nil {
			_, body := parseFrontmatter(string(data))
			contentHTML = renderMarkdown(body)
		}

		data := map[string]any{
			"title":        name,
			"description":  name + " - cptx",
			"canonicalURL": "https://cptx.ee/categories/" + name + "/",
			"posts":        toPostList(catPosts),
			"contentHTML":  contentHTML,
		}
		renderPage(w, engine, "ListPage", data)
	})))

	// Individual posts
	must(router.GET("post", "/posts/{section}/{slug}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		postPath := "/posts/" + m.Params["section"] + "/" + m.Params["slug"] + "/"
		post := postsByPath[postPath]
		if post == nil {
			renderNotFound(w, engine)
			return
		}

		// Find prev/next in same section
		sectionPosts := postsBySection[post.Section]
		var prevPost, nextPost any = false, false
		for i, p := range sectionPosts {
			if p.Path == post.Path {
				if i > 0 {
					nextPost = toPostMap(sectionPosts[i-1])
				}
				if i < len(sectionPosts)-1 {
					prevPost = toPostMap(sectionPosts[i+1])
				}
				break
			}
		}

		data := map[string]any{
			"post":     toPostMap(post),
			"prevPost": prevPost,
			"nextPost": nextPost,
		}
		renderPage(w, engine, "SinglePage", data)
	})))

	// Pockets index
	must(router.GET("pockets.index", "/pockets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for _index.md content
		var contentHTML string
		indexPath := filepath.Join("content", "pockets", "_index.md")
		if data, err := os.ReadFile(indexPath); err == nil {
			_, body := parseFrontmatter(string(data))
			contentHTML = renderMarkdown(body)
		}

		data := map[string]any{
			"title":        "pockets",
			"description":  "pockets - cptx",
			"canonicalURL": "https://cptx.ee/pockets/",
			"posts":        toPostList(pocketPosts),
			"contentHTML":  contentHTML,
		}
		renderPage(w, engine, "ListPage", data)
	})))

	// Individual pocket pages
	must(router.GET("pocket", "/pockets/{slug}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		postPath := "/pockets/" + m.Params["slug"] + "/"
		post := postsByPath[postPath]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		data := map[string]any{
			"post":     toPostMap(post),
			"prevPost": false,
			"nextPost": false,
		}
		renderPage(w, engine, "SinglePage", data)
	})))

	// Static files
	staticFS := http.FileServer(http.Dir("static"))
	mux := http.NewServeMux()
	mux.Handle("/css/", staticFS)
	mux.Handle("/fonts/", staticFS)
	mux.Handle("/js/", staticFS)
	mux.Handle("/favicon.ico", staticFS)
	mux.Handle("/favicon-16x16.png", staticFS)
	mux.Handle("/favicon-32x32.png", staticFS)
	mux.Handle("/android-chrome-192x192.png", staticFS)
	mux.Handle("/apple-touch-icon.png", staticFS)
	mux.HandleFunc("/shuffle-pool.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(shufflePoolJSON)
	})
	mux.Handle("/", router)

	// Build static site
	if err := buildStaticSite(engine, published, allPostsForPool, postsByCategory, postsBySection, pocketPosts, manifestoPost, toPostMap, toPostList, shufflePoolJSON); err != nil {
		log.Fatalf("building static site: %v", err)
	}

	addr := ":8080"
	log.Printf("Serving on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func renderPage(w http.ResponseWriter, engine *htmlc.Engine, component string, data map[string]any) {
	html, err := engine.RenderPageString(component, data)
	if err != nil {
		log.Printf("render error for %s: %v", component, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func renderNotFound(w http.ResponseWriter, engine *htmlc.Engine) {
	html, err := engine.RenderPageString("NotFoundPage", nil)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, html)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Content loading

func loadAllContent(root string) ([]*Post, error) {
	var posts []*Post

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		// Skip _index.md files
		if d.Name() == "_index.md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		post, err := parsePost(string(data), path, root)
		if err != nil {
			log.Printf("warning: skipping %s: %v", path, err)
			return nil
		}

		posts = append(posts, post)
		return nil
	})

	return posts, err
}

func parsePost(content, filePath, root string) (*Post, error) {
	fm, body := parseFrontmatter(content)

	title := fmString(fm, "title")
	if title == "" && fmString(fm, "date") == "" {
		return nil, fmt.Errorf("no title or date")
	}

	post := &Post{
		Title:    title,
		Author:   fmString(fm, "author"),
		Draft:    fmBool(fm, "draft"),
		ShowTags: fmBool(fm, "showTags"),
		TOC:      fmBool(fm, "toc"),
		Summary:  fmString(fm, "summary"),
		FilePath: filePath,
	}

	// Parse date
	if dateStr := fmString(fm, "date"); dateStr != "" {
		// Try various date formats
		for _, layout := range []string{
			"2006-01-02",
			"2006-01-02T15:04:05-07:00",
			"2006-01-02T15:04:05Z",
			time.RFC3339,
		} {
			if t, err := time.Parse(layout, dateStr); err == nil {
				post.Date = t
				break
			}
		}
	}
	post.DateFormatted = post.Date.Format("2006.01.02")

	// Parse categories and tags
	post.Categories = fmStringSlice(fm, "categories")
	post.Tags = fmStringSlice(fm, "tags")

	// Determine section and slug from file path
	relPath := strings.TrimPrefix(filePath, root+"/")

	// Determine the URL path based on content structure
	if strings.HasPrefix(relPath, "posts/") {
		// posts/section/filename.md -> /posts/section/slug/
		parts := strings.SplitN(relPath, "/", 3)
		if len(parts) >= 3 {
			post.Section = parts[1]
			post.Slug = slugify(strings.TrimSuffix(parts[2], ".md"))
			post.Path = "/posts/" + post.Section + "/" + post.Slug + "/"
		}
	} else if strings.HasPrefix(relPath, "pockets/") {
		// pockets/filename.md -> /pockets/slug/
		post.Section = "pockets"
		post.Slug = slugify(strings.TrimSuffix(filepath.Base(relPath), ".md"))
		post.Path = "/pockets/" + post.Slug + "/"
	} else if strings.TrimSuffix(filepath.Base(relPath), ".md") == "manifesto" {
		post.Section = ""
		post.Slug = "manifesto"
		post.Path = "/manifesto/"
	} else {
		// Other top-level content
		post.Slug = slugify(strings.TrimSuffix(filepath.Base(relPath), ".md"))
		post.Path = "/" + post.Slug + "/"
	}

	// Render markdown
	post.Content = template.HTML(renderMarkdown(body))

	return post, nil
}

func renderMarkdown(source string) string {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(source), &buf); err != nil {
		log.Printf("markdown render error: %v", err)
		return source
	}
	return buf.String()
}

// YAML frontmatter parsing (simple, no external dependency)

func parseFrontmatter(content string) (map[string]string, string) {
	fm := make(map[string]string)

	// Trim leading whitespace (some files have blank lines before frontmatter)
	content = strings.TrimLeft(content, " \t\n\r")

	if !strings.HasPrefix(content, "---") {
		return fm, content
	}

	end := strings.Index(content[3:], "---")
	if end < 0 {
		return fm, content
	}

	fmContent := content[3 : end+3]
	body := content[end+6:]

	for _, line := range strings.Split(fmContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Remove surrounding quotes
		val = strings.Trim(val, "\"'")
		fm[key] = val
	}

	return fm, body
}

func fmString(fm map[string]string, key string) string {
	return fm[key]
}

func fmBool(fm map[string]string, key string) bool {
	v := strings.ToLower(fm[key])
	return v == "true" || v == "yes"
}

func fmStringSlice(fm map[string]string, key string) []string {
	val := fm[key]
	if val == "" {
		return nil
	}
	// Handle YAML array syntax: [item1, item2] or ["item1", "item2"]
	val = strings.Trim(val, "[]")
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"'")
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Slugify converts a filename to a URL-safe slug, matching Hugo's behavior.
func slugify(s string) string {
	// Normalize unicode
	s = norm.NFC.String(s)

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with dashes
	s = strings.ReplaceAll(s, " ", "-")

	// Remove characters that aren't alphanumeric, dashes, or dots
	var buf strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' {
			buf.WriteRune(r)
		}
	}
	s = buf.String()

	// Trim leading/trailing dashes
	s = strings.Trim(s, "-")

	return s
}

// Static site builder

func buildStaticSite(
	engine *htmlc.Engine,
	published []*Post,
	allPostsForPool []*Post,
	postsByCategory map[string][]*Post,
	postsBySection map[string][]*Post,
	pocketPosts []*Post,
	manifestoPost *Post,
	toPostMap func(*Post) map[string]any,
	toPostList func([]*Post) []any,
	shufflePoolJSON []byte,
) error {
	distDir := "dist"
	os.RemoveAll(distDir)

	// Copy static files
	if err := copyDir("static", filepath.Join(distDir)); err != nil {
		return fmt.Errorf("copying static: %w", err)
	}

	// Write shuffle pool JSON
	if err := os.WriteFile(filepath.Join(distDir, "shuffle-pool.json"), shufflePoolJSON, 0o644); err != nil {
		return fmt.Errorf("writing shuffle-pool.json: %w", err)
	}

	// Helper to write a page
	writePage := func(path string, component string, data map[string]any) error {
		html, err := engine.RenderPageString(component, data)
		if err != nil {
			return fmt.Errorf("rendering %s for %s: %w", component, path, err)
		}
		outPath := filepath.Join(distDir, path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outPath, []byte(html), 0o644)
	}

	// Home page
	if err := writePage("index.html", "HomePage", map[string]any{
		"posts": toPostList(allPostsForPool),
	}); err != nil {
		return err
	}

	// Manifesto
	if manifestoPost != nil {
		if err := writePage("manifesto/index.html", "SinglePage", map[string]any{
			"post":     toPostMap(manifestoPost),
			"prevPost": nil,
			"nextPost": nil,
		}); err != nil {
			return err
		}
	}

	// Category pages — from posts and from _index.md files
	allCategories := make(map[string]bool)
	for name := range postsByCategory {
		allCategories[name] = true
	}
	// Also discover categories with _index.md but no posts
	if entries, err := os.ReadDir(filepath.Join("content", "categories")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				allCategories[e.Name()] = true
			}
		}
	}

	for name := range allCategories {
		catPosts := postsByCategory[name]
		var contentHTML string
		indexPath := filepath.Join("content", "categories", name, "_index.md")
		if data, err := os.ReadFile(indexPath); err == nil {
			_, body := parseFrontmatter(string(data))
			contentHTML = renderMarkdown(body)
		}

		if err := writePage(filepath.Join("categories", name, "index.html"), "ListPage", map[string]any{
			"title":        name,
			"description":  name + " - cptx",
			"canonicalURL": "https://cptx.ee/categories/" + name + "/",
			"posts":        toPostList(catPosts),
			"contentHTML":  contentHTML,
		}); err != nil {
			return err
		}
	}

	// Individual posts
	for _, post := range published {
		if post.Path == "/manifesto/" {
			continue // already handled
		}

		// Find prev/next for posts in sections
		var prevPost, nextPost any = false, false
		if post.Section == "recuerdos" || post.Section == "croquis" {
			sectionPosts := postsBySection[post.Section]
			for i, p := range sectionPosts {
				if p.Path == post.Path {
					if i > 0 {
						nextPost = toPostMap(sectionPosts[i-1])
					}
					if i < len(sectionPosts)-1 {
						prevPost = toPostMap(sectionPosts[i+1])
					}
					break
				}
			}
		}

		outPath := strings.TrimPrefix(post.Path, "/")
		if err := writePage(filepath.Join(outPath, "index.html"), "SinglePage", map[string]any{
			"post":     toPostMap(post),
			"prevPost": prevPost,
			"nextPost": nextPost,
		}); err != nil {
			return err
		}
	}

	// Pockets index
	var pocketsContentHTML string
	indexPath := filepath.Join("content", "pockets", "_index.md")
	if data, err := os.ReadFile(indexPath); err == nil {
		_, body := parseFrontmatter(string(data))
		pocketsContentHTML = renderMarkdown(body)
	}

	if err := writePage("pockets/index.html", "ListPage", map[string]any{
		"title":        "pockets",
		"description":  "pockets - cptx",
		"canonicalURL": "https://cptx.ee/pockets/",
		"posts":        toPostList(pocketPosts),
		"contentHTML":  pocketsContentHTML,
	}); err != nil {
		return err
	}

	// 404 page
	if err := writePage("404.html", "NotFoundPage", nil); err != nil {
		return err
	}

	log.Printf("Static site built to %s/", distDir)
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0o644)
	})
}
