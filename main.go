package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dhamidi/dispatch"
	"github.com/dhamidi/htmlc"
)

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
	if err := computePostIDs(published); err != nil {
		log.Fatalf("post IDs: %v", err)
	}

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
	legalPosts := []*Post{}
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
		} else if p.Section == "legaldocs" {
			legalPosts = append(legalPosts, p)
		} else if p.Path == "/manifesto/" {
			manifestoPost = p
		}
	}

	allPostsForPool := allPosts

	postWords := map[string]int{} // filled once below; used by the read beacon (A-02)

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
			"Lang":          p.Lang,
			"ContentClass":  contentClass(p),
			"Description":   p.Description,
			"Tags":          tags,
			"ShowTags":      p.ShowTags,
			"TOC":           p.TOC,
			"Categories":    categories,
			"ID":            p.ID,
			"ReadsLabel":    readsLabel(siteStats.readsOf(p.Path)),
			"Words":         postWords[p.Path],
			"Readable":      p.Section == "croquis" || p.Section == "recuerdos" || isIssue(p) || isIssueES(p),
			"IsIssue":       isIssue(p) || isIssueES(p),
			"IssueES":       isIssueES(p),
			"IssueAltPath":  issueAltPath(p, postsByPath),
			"IssueLabel":    issueLabel(p.Slug),
			"PageTitle":     pageTitleFor(p),
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
			"W":        p.W,
			"H":        p.H,
			"HasSize":  p.W > 0 && p.H > 0,
			"Ratio":    photoRatio(p),
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
			"Name":     f.Name,
			"Title":    f.Title,
			"Location": f.Location,
			"Year":     f.Year,
			"Path":     f.Path,
			"Count":    f.Count,
			"Cover":    folderCover(f),
			"CoverW":   folderCoverSize(f, 0),
			"CoverH":   folderCoverSize(f, 1),
			"Photos":   toPhotoList(f.Photos),
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
	addonIdx := buildAddonIndex(published) // F-15
	addonTagList := addonTags(addonIdx)
	postTexts := map[string]string{}
	for _, p := range published {
		postTexts[p.Path] = plainText(string(p.Content))
		postWords[p.Path] = len(strings.Fields(postTexts[p.Path]))
	}
	// A-01…A-04: readers' counters (visits, reads, countries)
	readable := map[string]bool{}
	for _, p := range addonIdx {
		readable[p.Path] = true
	}
	// newsletter issues are read and counted like posts
	coverIdx = addonIdx
	readerExtras = nil
	for _, p := range published {
		if isIssue(p) || isIssueES(p) {
			readable[p.Path] = true
			title := issueLabel(p.Slug) + ": " + p.Title
			if isIssueES(p) {
				title += " (es)"
			}
			readerExtras = append(readerExtras, addonPost{Title: title, Path: p.Path, Section: "newsletter", Day: p.Date.Format("2006-01-02")})
		}
	}
	statsDir, geoFile := statsDirs()
	geo := &geoKeeper{path: geoFile}
	if geoFile != "" {
		geo.load()
		if statsDir != "" {
			go geo.refreshLoop()
		}
	}
	siteStats = newStatsStore(statsDir, readable, geo)
	if statsDir != "" {
		go siteStats.saveLoop()
	}
	go func() { // save the day's totals when systemd stops or restarts the server
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, os.Interrupt)
		<-sig
		siteStats.save()
		os.Exit(0)
	}()

	lib, err := loadLibrary(libraryDir(), addonIdx, postTexts) // L-01
	if err != nil {
		log.Printf("library: %v", err)
		lib = &library{ByKey: map[string]*libBook{}}
	}
	// F-06: site-wide search over posts' full text and the library
	searchIdx := buildSearchIndex(addonIdx, postTexts, lib)

	must(router.GET("home", "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otd, hasOTD := homeOnThisDay(addonIdx, time.Now())
		data := map[string]any{
			"posts":        toPostList(allPostsForPool),
			"postCount":    postCount,
			"onThisDay":    otd,
			"hasOnThisDay": hasOTD,
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

	// Contact page
	// About: hub for manifesto, contact and legal docs (same layout as addons)
	must(router.GET("about", "/about/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderPage(w, engine, "AboutPage", map[string]any{})
	})))

	// Newsletter (batch K): what the digest is, the sign-up form and the kiosk of past issues
	must(router.GET("newsletter", "/newsletter/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderPage(w, engine, "NewsletterPage", newsletterData(published, 3))
	})))
	// Spanish versions of the issues: /newsletter/<n>/es/
	must(router.GET("newsletter.es", "/newsletter/{slug}/es/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		post := postsByPath["/newsletter/"+m.Params["slug"]+"/es/"]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		renderPage(w, engine, "SinglePage", map[string]any{"post": toPostMap(post), "prevPost": false, "nextPost": false})
	})))
	// /newsletter/kiosk/ and the issues (/newsletter/<n>/) share one pattern
	must(router.GET("newsletter.page", "/newsletter/{slug}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		if m.Params["slug"] == "kiosk" {
			renderPage(w, engine, "KioskPage", newsletterData(published, 0))
			return
		}
		post := postsByPath["/newsletter/"+m.Params["slug"]+"/"]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		renderPage(w, engine, "SinglePage", map[string]any{"post": toPostMap(post), "prevPost": false, "nextPost": false})
	})))

	must(router.GET("contact", "/contact/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderPage(w, engine, "ContactPage", map[string]any{})
	})))

	// Category pages
	must(router.GET("category", "/categories/{name}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		name := m.Params["name"]
		if name == "pockets" {
			http.Redirect(w, r, "/pockets/", http.StatusMovedPermanently)
			return
		}
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
		// dispatch matches against the escaped RequestURI, so params arrive
		// percent-encoded (c%C3%A9sped). Decode and NFC-normalise them to match
		// the slugs built by slugify (B-16).
		oldPath := "/posts/" + legacyParam(m.Params["section"]) + "/" + legacyParam(m.Params["slug"]) + "/"
		post := postsByOldPath[oldPath]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		http.Redirect(w, r, post.Path, http.StatusMovedPermanently)
	})))

	// Build pockets data for the accordion page
	pocketsData := buildPocketsData(pocketPosts)

	// Pockets index
	must(router.GET("pockets.index", "/pockets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"pockets":     pocketsData,
			"activeIndex": 0,
		}
		renderPage(w, engine, "PocketsPage", data)
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
			"total":     len(allPhotos),
			"years":     photoYears(photoFolders),
			"places":    len(photoFolders),
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

	// Legal docs: hub + pages (privacy, cookies, how the site works, newsletter)
	legalData := buildLegalData(legalPosts)
	must(router.GET("legal.index", "/legaldocs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderPage(w, engine, "LegalPage", map[string]any{"docs": legalData})
	})))
	must(router.GET("legal", "/legaldocs/{slug}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m, _ := dispatch.MatchFromContext(r.Context())
		post := postsByPath["/legaldocs/"+m.Params["slug"]+"/"]
		if post == nil {
			renderNotFound(w, engine)
			return
		}
		renderPage(w, engine, "SinglePage", map[string]any{"post": toPostMap(post), "prevPost": false, "nextPost": false})
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

	// Feed page (server-side only, not static)
	// Note: registered on mux below since it's under /pockets/feed/ which would
	// otherwise match the pocket slug route.

	// Static files
	staticFS := http.FileServer(http.Dir("static"))
	mux := http.NewServeMux()
	mux.Handle("/css/", staticFS)
	// F-06: on the mux, not the dispatch router, because dispatch matches the
	// whole RequestURI and /search/?q=… would 404 (same as the addons).
	mux.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/" {
			renderNotFound(w, engine)
			return
		}
		renderPage(w, engine, "SearchPage", searchData(searchIdx, r.URL.Query().Get("q")))
	})
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
	// Feed page handlers (server-side only)
	startFeedAutoRefresh()
	mux.HandleFunc("/pockets/cheatsheet/feed/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !refreshFeeds(feedManualGap, "manual") {
			log.Println("Feed refresh skipped: running or refreshed recently")
		}
		http.Redirect(w, r, "/pockets/cheatsheet/feed/", http.StatusSeeOther)
	})
	// Redirect old /pockets/feed/ to new location
	mux.HandleFunc("/pockets/feed/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/pockets/cheatsheet/feed/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/pockets/cheatsheet/feed/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pockets/cheatsheet/feed/" {
			router.ServeHTTP(w, r)
			return
		}
		cache, err := loadFeedCache()
		noCache := err != nil

		var items []any
		var feedErrors []any
		var lastRefreshed string
		var hasItems, hasErrors bool
		var errorCount int

		if !noCache {
			lastRefreshed = cache.LastRefreshed.Format("2006.01.02 15:04 MST")
			for _, item := range cache.Items {
				items = append(items, map[string]any{
					"Title":         item.Title,
					"Link":          item.Link,
					"Source":        item.Source,
					"DateFormatted": item.DateFormatted,
				})
			}
			for _, e := range cache.Errors {
				feedErrors = append(feedErrors, e)
			}
			hasItems = len(items) > 0
			hasErrors = len(feedErrors) > 0
			errorCount = len(feedErrors)
		}

		data := map[string]any{
			"noCache":       noCache,
			"hasItems":      hasItems,
			"hasErrors":     hasErrors,
			"errorCount":    errorCount,
			"lastRefreshed": lastRefreshed,
			"items":         items,
			"errors":        feedErrors,
		}
		renderPage(w, engine, "FeedPage", data)
	})

	// Batch A: health, feed, sitemap, robots (B-05, B-06)
	started := time.Now()
	rssXML := buildRSS(published, 20)
	var sitemapXML []byte // filled after the static build below
	mux.HandleFunc("/healthz", healthHandler(postCount, started))
	mux.HandleFunc("/_read", readHandler) // A-02: read beacon
	registerAddons(mux, func(w http.ResponseWriter, c string, d map[string]any) { renderPage(w, engine, c, d) }, addonIdx, addonTagList)
	registerLibrary(mux, func(w http.ResponseWriter, c string, d map[string]any) { renderPage(w, engine, c, d) },
		func(w http.ResponseWriter, r *http.Request) { renderNotFound(w, engine) }, lib)
	mux.HandleFunc("/index.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write(rssXML)
	})
	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write(sitemapXML)
	})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(robotsTxt))
	})

	mux.Handle("/", router)

	// Build static site
	if err := buildStaticSite(engine, published, allPostsForPool, postsByCategory, postsBySection, pocketPosts, manifestoPost, toPostMap, toPostList, shufflePoolJSON, postCount, photoFolders, toFolderMap, toFolderList, toPhotoList); err != nil {
		log.Fatalf("building static site: %v", err)
	}
	if err := writeAddonPages(engine, "dist", addonIdx, addonTagList); err != nil {
		log.Printf("addons dist: %v", err)
	}
	if html, err := engine.RenderPageString("SearchPage", searchData(searchIdx, "")); err == nil {
		os.MkdirAll(filepath.Join("dist", "search"), 0o755)
		os.WriteFile(filepath.Join("dist", "search", "index.html"), []byte(withDoctype(html)), 0o644)
	} else {
		log.Printf("search dist: %v", err)
	}
	if err := writeLibraryPages(engine, "dist", lib); err != nil {
		log.Printf("library dist: %v", err)
	}
	if err := registerRetired(mux, "dist"); err != nil {
		log.Printf("retired URLs: %v", err)
	}
	if sitemapXML, err = buildSitemap("dist", published); err != nil {
		log.Printf("sitemap: %v", err)
	}

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" { // B-16: lets a test copy run beside the live server
		addr = ":" + p
	}
	log.Printf("Serving on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, withGzip(withHeaders(withStats(mux)))))
}

func renderPage(w http.ResponseWriter, engine *htmlc.Engine, component string, data map[string]any) {
	html, err := engine.RenderPageString(component, data)
	if err != nil {
		log.Printf("render error for %s: %v", component, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, withDoctype(html))
}

func renderNotFound(w http.ResponseWriter, engine *htmlc.Engine) {
	html, err := engine.RenderPageString("NotFoundPage", nil)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, withDoctype(html))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
