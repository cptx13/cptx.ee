package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	_ "image/png"
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
	"golang.org/x/image/draw"
	"golang.org/x/text/unicode/norm"
)

// Photo represents a single image in a photo folder.
type Photo struct {
	ID       string // e.g. "2016_Madrid-001"
	Folder   string // e.g. "2016_Madrid"
	File     string // e.g. "001.jpg"
	URL      string // e.g. "/photos/2016_Madrid/001.jpg"
	ThumbURL string // e.g. "/photos/2016_Madrid/thumb_001.jpg"
	Index    int    // 1-based
}

// PhotoFolder represents a directory of photos.
type PhotoFolder struct {
	Name   string  // directory name e.g. "2016_Madrid"
	Title  string  // display title e.g. "2016 Madrid"
	Photos []Photo
	Path   string  // e.g. "/photos/2016_Madrid/"
	Count  int
}

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
	OldPath       string
	FilePath      string
	DateFormatted string
	DateYear      string
	DateShort     string
	ID            string
}

func main() {
	posts, err := loadAllContent("content")
	if err != nil {
		log.Fatalf("loading content: %v", err)
	}

	// Filter out drafts
	var published []*Post
	for _, p := range posts {
		if !p.Draft {
			published = append(published, p)
		}
	}

	// Compute IDs for posts in recuerdos/croquis sections and assign new paths
	computePostIDs(published)

	// Load photos
	photoFolders, err := loadPhotos("content")
	if err != nil {
		log.Fatalf("loading photos: %v", err)
	}

	// Generate thumbnails for dev mode (into content/photos/)
	for _, folder := range photoFolders {
		for _, photo := range folder.Photos {
			srcPath := filepath.Join("content", "photos", folder.Name, photo.File)
			thumbPath := filepath.Join("content", "photos", folder.Name, "thumb_"+photo.File)
			if _, err := os.Stat(thumbPath); err == nil {
				continue // thumbnail already exists
			}
			if err := generateThumbnail(srcPath, thumbPath, 600, 80); err != nil {
				log.Printf("warning: dev thumbnail for %s: %v", photo.ID, err)
			}
		}
	}

	// Sort posts by date descending
	sort.Slice(published, func(i, j int) bool {
		return published[i].Date.After(published[j].Date)
	})

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
	postsByOldPath := map[string]*Post{}
	pocketPosts := []*Post{}
	var manifestoPost *Post
	var allPosts []*Post // all posts for shuffle pool (posts section only)

	for _, p := range published {
		postsByPath[p.Path] = p
		if p.OldPath != "" {
			postsByOldPath[p.OldPath] = p
		}
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
			"DateYear":      p.DateYear,
			"DateShort":     p.DateShort,
			"Content":       string(p.Content),
			"Slug":          p.Slug,
			"Section":       p.Section,
			"Path":          p.Path,
			"OldPath":       p.OldPath,
			"Summary":       p.Summary,
			"Tags":          tags,
			"ShowTags":      p.ShowTags,
			"TOC":           p.TOC,
			"Categories":    categories,
			"ID":            p.ID,
		}
	}

	toPostList := func(posts []*Post) []any {
		result := make([]any, len(posts))
		for i, p := range posts {
			result[i] = toPostMap(p)
		}
		return result
	}

	toPhotoMap := func(p Photo) map[string]any {
		return map[string]any{
			"ID":       p.ID,
			"Folder":   p.Folder,
			"File":     p.File,
			"URL":      p.URL,
			"ThumbURL": p.ThumbURL,
			"Index":    p.Index,
		}
	}

	toPhotoList := func(photos []Photo) []any {
		result := make([]any, len(photos))
		for i, p := range photos {
			result[i] = toPhotoMap(p)
		}
		return result
	}

	toFolderMap := func(f PhotoFolder) map[string]any {
		return map[string]any{
			"Name":   f.Name,
			"Title":  f.Title,
			"Path":   f.Path,
			"Count":  f.Count,
			"Photos": toPhotoList(f.Photos),
		}
	}

	toFolderList := func(folders []PhotoFolder) []any {
		result := make([]any, len(folders))
		for i, f := range folders {
			result[i] = toFolderMap(f)
		}
		return result
	}

	// Generate shuffle pool JSON (just paths)
	shufflePoolPaths := make([]string, len(allPostsForPool))
	for i, p := range allPostsForPool {
		shufflePoolPaths[i] = p.Path
	}
	shufflePoolJSON, _ := json.Marshal(shufflePoolPaths)

	postCount := len(allPostsForPool)

	// Home page
	must(router.GET("home", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"posts":     toPostList(allPostsForPool),
			"postCount": postCount,
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

	// Individual posts (new URL scheme: /{category}/{id}/)
	must(router.GET("post", "/{category}/{id}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		postPath := "/" + m.Params["category"] + "/" + m.Params["id"] + "/"
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

	// Old URL redirects: /posts/{section}/{slug}/ -> new path
	must(router.GET("old-post-redirect", "/posts/{section}/{slug}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		oldPath := "/posts/" + m.Params["section"] + "/" + m.Params["slug"] + "/"
		post := postsByOldPath[oldPath]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		http.Redirect(w, r, post.Path, http.StatusMovedPermanently)
	})))

	// Pockets index
	must(router.GET("pockets.index", "/pockets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Photos index
	must(router.GET("photos.index", "/photos/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var allPhotos []any
		for _, f := range photoFolders {
			allPhotos = append(allPhotos, toPhotoList(f.Photos)...)
		}
		data := map[string]any{
			"folders":   toFolderList(photoFolders),
			"allPhotos": allPhotos,
		}
		renderPage(w, engine, "PhotosPage", data)
	})))

	// Photo folder page
	must(router.GET("photos.folder", "/photos/{folder}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		folderName := m.Params["folder"]
		var folder *PhotoFolder
		for i := range photoFolders {
			if photoFolders[i].Name == folderName {
				folder = &photoFolders[i]
				break
			}
		}
		if folder == nil {
			renderNotFound(w, engine)
			return
		}
		data := map[string]any{
			"folder": toFolderMap(*folder),
		}
		renderPage(w, engine, "PhotoFolderPage", data)
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
	// Serve photo images from content/photos/ in dev mode
	photoFileServer := http.StripPrefix("/photos/", http.FileServer(http.Dir("content/photos")))
	mux.HandleFunc("/photos/", func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
			photoFileServer.ServeHTTP(w, r)
		default:
			router.ServeHTTP(w, r)
		}
	})
	mux.Handle("/", router)

	// Build static site
	if err := buildStaticSite(engine, published, allPostsForPool, postsByCategory, postsBySection, pocketPosts, manifestoPost, toPostMap, toPostList, shufflePoolJSON, postCount, photoFolders, toFolderMap, toFolderList, toPhotoList); err != nil {
		log.Fatalf("building static site: %v", err)
	}

	addr := ":8080"
	log.Printf("Serving on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// computePostIDs assigns IDs to posts in the recuerdos/croquis sections.
// IDs are YYYYMMDD + 1-based index within that (category, date) group.
// Posts within the same date are sorted by filename alphabetically.
func computePostIDs(posts []*Post) {
	// Collect posts that need IDs (recuerdos or croquis)
	type postKey struct {
		category string
		dateStr  string
	}
	groups := map[postKey][]*Post{}

	for _, p := range posts {
		if p.Section != "recuerdos" && p.Section != "croquis" {
			continue
		}
		// Use first category as the URL category; fall back to section
		category := p.Section
		if len(p.Categories) > 0 {
			category = p.Categories[0]
		}
		dateStr := p.Date.Format("20060102")
		key := postKey{category: category, dateStr: dateStr}
		groups[key] = append(groups[key], p)
	}

	// Sort each group by filename and assign IDs
	for key, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			return filepath.Base(group[i].FilePath) < filepath.Base(group[j].FilePath)
		})
		for i, p := range group {
			p.ID = fmt.Sprintf("%s%d", key.dateStr, i+1)
			p.OldPath = "/posts/" + p.Section + "/" + p.Slug + "/"
			p.Path = "/" + key.category + "/" + p.ID + "/"
		}
	}
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

func loadPhotos(root string) ([]PhotoFolder, error) {
	photosDir := filepath.Join(root, "photos")
	entries, err := os.ReadDir(photosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading photos dir: %w", err)
	}

	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}

	var folders []PhotoFolder
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folderName := entry.Name()
		folderPath := filepath.Join(photosDir, folderName)

		files, err := os.ReadDir(folderPath)
		if err != nil {
			return nil, fmt.Errorf("reading photo folder %s: %w", folderName, err)
		}

		var imageFiles []string
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			// Skip generated thumbnails
			if strings.HasPrefix(f.Name(), "thumb_") {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if imageExts[ext] {
				imageFiles = append(imageFiles, f.Name())
			}
		}

		sort.Strings(imageFiles)

		var photos []Photo
		for i, file := range imageFiles {
			idx := i + 1
			photos = append(photos, Photo{
				ID:       fmt.Sprintf("%s-%03d", folderName, idx),
				Folder:   folderName,
				File:     file,
				URL:      fmt.Sprintf("/photos/%s/%s", folderName, file),
				ThumbURL: fmt.Sprintf("/photos/%s/thumb_%s", folderName, file),
				Index:    idx,
			})
		}

		folders = append(folders, PhotoFolder{
			Name:   folderName,
			Title:  strings.ReplaceAll(folderName, "_", " "),
			Photos: photos,
			Path:   "/photos/" + folderName + "/",
			Count:  len(photos),
		})
	}

	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})

	return folders, nil
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
	post.DateYear = post.Date.Format("06")
	post.DateShort = post.Date.Format("06.01.02")

	// Parse categories and tags
	post.Categories = fmStringSlice(fm, "categories")
	post.Tags = fmStringSlice(fm, "tags")

	// Determine section and slug from file path
	relPath := strings.TrimPrefix(filePath, root+"/")

	// Determine the URL path based on content structure
	// Note: for posts in recuerdos/croquis, the Path will be overwritten by computePostIDs
	if strings.HasPrefix(relPath, "posts/") {
		parts := strings.SplitN(relPath, "/", 3)
		if len(parts) >= 3 {
			post.Section = parts[1]
			post.Slug = slugify(strings.TrimSuffix(parts[2], ".md"))
			post.Path = "/posts/" + post.Section + "/" + post.Slug + "/"
		}
	} else if strings.HasPrefix(relPath, "pockets/") {
		post.Section = "pockets"
		post.Slug = slugify(strings.TrimSuffix(filepath.Base(relPath), ".md"))
		post.Path = "/pockets/" + post.Slug + "/"
	} else if strings.TrimSuffix(filepath.Base(relPath), ".md") == "manifesto" {
		post.Section = ""
		post.Slug = "manifesto"
		post.Path = "/manifesto/"
	} else {
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
	s = norm.NFC.String(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	var buf strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' {
			buf.WriteRune(r)
		}
	}
	s = buf.String()
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
	postCount int,
	photoFolders []PhotoFolder,
	toFolderMap func(PhotoFolder) map[string]any,
	toFolderList func([]PhotoFolder) []any,
	toPhotoList func([]Photo) []any,
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

	// Helper to write a redirect page
	writeRedirect := func(oldPath, newPath string) error {
		redirectHTML := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta http-equiv="refresh" content="0; url=%s"><link rel="canonical" href="%s"><title>Redirecting...</title></head><body><p>Redirecting to <a href="%s">%s</a>...</p></body></html>`, newPath, newPath, newPath, newPath)
		outPath := filepath.Join(distDir, strings.TrimPrefix(oldPath, "/"), "index.html")
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outPath, []byte(redirectHTML), 0o644)
	}

	// Home page
	if err := writePage("index.html", "HomePage", map[string]any{
		"posts":     toPostList(allPostsForPool),
		"postCount": postCount,
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

	// Category pages
	allCategories := make(map[string]bool)
	for name := range postsByCategory {
		allCategories[name] = true
	}
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
			continue
		}

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

		// Write redirect from old path to new path
		if post.OldPath != "" && post.OldPath != post.Path {
			if err := writeRedirect(post.OldPath, post.Path); err != nil {
				return fmt.Errorf("writing redirect for %s: %w", post.OldPath, err)
			}
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

	// Photos section
	if len(photoFolders) > 0 {
		// Copy photo images and generate thumbnails
		for _, folder := range photoFolders {
			srcDir := filepath.Join("content", "photos", folder.Name)
			dstDir := filepath.Join(distDir, "photos", folder.Name)
			if err := copyDir(srcDir, dstDir); err != nil {
				return fmt.Errorf("copying photos for %s: %w", folder.Name, err)
			}
			// Generate thumbnails (600px wide, quality 80)
			for _, photo := range folder.Photos {
				srcPath := filepath.Join(srcDir, photo.File)
				thumbPath := filepath.Join(dstDir, "thumb_"+photo.File)
				if err := generateThumbnail(srcPath, thumbPath, 600, 80); err != nil {
					log.Printf("warning: thumbnail for %s: %v", photo.ID, err)
				}
			}
		}

		// Photos index page
		var allPhotos []any
		for _, f := range photoFolders {
			allPhotos = append(allPhotos, toPhotoList(f.Photos)...)
		}
		if err := writePage("photos/index.html", "PhotosPage", map[string]any{
			"folders":   toFolderList(photoFolders),
			"allPhotos": allPhotos,
		}); err != nil {
			return err
		}

		// Individual folder pages
		for _, folder := range photoFolders {
			if err := writePage(filepath.Join("photos", folder.Name, "index.html"), "PhotoFolderPage", map[string]any{
				"folder": toFolderMap(folder),
			}); err != nil {
				return err
			}
		}
	}

	// 404 page
	if err := writePage("404.html", "NotFoundPage", nil); err != nil {
		return err
	}

	log.Printf("Static site built to %s/", distDir)
	return nil
}

// generateThumbnail resizes an image to fit within maxWidth pixels wide,
// preserving aspect ratio, and writes JPEG at the given quality.
func generateThumbnail(srcPath, dstPath string, maxWidth int, quality int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decoding %s: %w", srcPath, err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Only downscale, never upscale
	newW := maxWidth
	newH := origH * maxWidth / origW
	if origW <= maxWidth {
		newW = origW
		newH = origH
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, dst, &jpeg.Options{Quality: quality})
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
