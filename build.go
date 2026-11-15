// Static build: renders every page into dist/ (B-11).

package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhamidi/htmlc"
)

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
		return os.WriteFile(outPath, []byte(withDoctype(html)), 0o644)
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
		"posts":        toPostList(allPostsForPool),
		"postCount":    postCount,
		"onThisDay":    []any{},
		"hasOnThisDay": false,
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

	// Contact page
	if err := writePage("newsletter/index.html", "NewsletterPage", map[string]any{}); err != nil {
		return err
	}
	if err := writePage("contact/index.html", "ContactPage", map[string]any{}); err != nil {
		return err
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
		if name == "pockets" {
			// Write a redirect page instead of a category listing
			redirectHTML := `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="0;url=/pockets/"><link rel="canonical" href="https://cptx.ee/pockets/"></head><body><a href="/pockets/">Redirecting to /pockets/</a></body></html>`
			redirectPath := filepath.Join(distDir, "categories", "pockets", "index.html")
			if err := os.MkdirAll(filepath.Dir(redirectPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(redirectPath, []byte(redirectHTML), 0644); err != nil {
				return err
			}
			continue
		}
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

	// Legal docs hub (the pages themselves are written with the posts above)
	if err := writePage("legaldocs/index.html", "LegalPage", map[string]any{
		"docs": buildLegalData(published),
	}); err != nil {
		return err
	}

	// Pockets index (accordion page)
	pocketsData := buildPocketsData(pocketPosts)

	if err := writePage("pockets/index.html", "PocketsPage", map[string]any{
		"pockets":     pocketsData,
		"activeIndex": 0,
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
			// Generate thumbnails (600px wide, quality 80) — skip if already copied from dev cache
			for _, photo := range folder.Photos {
				thumbPath := filepath.Join(dstDir, "thumb_"+photo.File)
				if _, err := os.Stat(thumbPath); err == nil {
					continue // already exists (copied from content/photos/)
				}
				srcPath := filepath.Join(srcDir, photo.File)
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
			"total":     len(allPhotos),
			"years":     photoYears(photoFolders),
			"places":    len(photoFolders),
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
