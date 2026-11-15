// Content: loading markdown posts, frontmatter parsing, post IDs and URLs (B-11).

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/yuin/goldmark"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
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
	OldPath       string
	FilePath      string
	DateFormatted string
	DateYear      string
	DateShort     string
	ID            string
	FixedID       string // B-10: frontmatter `id:` pins the URL ID
	Lang          string // F-01: frontmatter `lang` or detected (es/en)
	Description   string // F-03: summary, or an excerpt of the body
}

// computePostIDs assigns IDs to posts in the recuerdos/croquis sections.
// IDs are YYYYMMDD + 1-based index within that (category, date) group.
// Posts within the same date are sorted by filename alphabetically.
//
// B-10: a post can pin its ID with `id: <digits>` in its frontmatter (e.g. to
// keep its URL after a date or filename change). Pinned IDs are taken first;
// the other posts of a group then get the lowest free index, in filename order.
// Two posts pinning the same ID in one category is an error.
func computePostIDs(posts []*Post) error {
	type postKey struct {
		category string
		dateStr  string
	}
	groups := map[postKey][]*Post{}
	var keys []postKey
	taken := map[string]string{} // category/id -> file that pinned it

	for _, p := range posts {
		if p.Section != "recuerdos" && p.Section != "croquis" {
			continue
		}
		// Use first category as the URL category; fall back to section
		category := p.Section
		if len(p.Categories) > 0 {
			category = p.Categories[0]
		}
		p.OldPath = "/posts/" + p.Section + "/" + p.Slug + "/"
		if p.FixedID != "" {
			if !fixedIDRe.MatchString(p.FixedID) {
				return fmt.Errorf("%s: id %q must be digits (YYYYMMDD + n)", p.FilePath, p.FixedID)
			}
			k := category + "/" + p.FixedID
			if other, dup := taken[k]; dup {
				return fmt.Errorf("%s: id %s is already pinned by %s", p.FilePath, p.FixedID, other)
			}
			taken[k] = p.FilePath
			p.ID = p.FixedID
			p.Path = "/" + category + "/" + p.ID + "/"
			continue
		}
		key := postKey{category: category, dateStr: p.Date.Format("20060102")}
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], p)
	}

	// Sort each group by filename and assign the lowest free index
	for _, key := range keys {
		group := groups[key]
		sort.Slice(group, func(i, j int) bool {
			return filepath.Base(group[i].FilePath) < filepath.Base(group[j].FilePath)
		})
		n := 1
		for _, p := range group {
			for {
				if _, pinned := taken[key.category+"/"+fmt.Sprintf("%s%d", key.dateStr, n)]; !pinned {
					break
				}
				n++
			}
			p.ID = fmt.Sprintf("%s%d", key.dateStr, n)
			p.Path = "/" + key.category + "/" + p.ID + "/"
			n++
		}
	}
	return nil
}

var fixedIDRe = regexp.MustCompile(`^[0-9]{9,}$`)

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
		FixedID:  fmString(fm, "id"),
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
	} else if strings.HasPrefix(relPath, "newsletter/") {
		// issues: content/newsletter/007.md → /newsletter/7/
		post.Section = "newsletter"
		post.Slug = strings.TrimLeft(strings.TrimSuffix(filepath.Base(relPath), ".md"), "0")
		post.Path = "/newsletter/" + post.Slug + "/"
	} else if strings.HasPrefix(relPath, "legaldocs/") {
		post.Section = "legaldocs"
		post.Slug = slugify(strings.TrimSuffix(filepath.Base(relPath), ".md"))
		post.Path = "/legaldocs/" + post.Slug + "/"
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
	post.Lang = postLang(fmString(fm, "lang"), body)
	post.Description = post.Summary
	if post.Description == "" {
		post.Description = excerpt(body, 160)
	}

	return post, nil
}

var md = goldmark.New(
	goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
)

func renderMarkdown(source string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		log.Printf("markdown render error: %v", err)
		return source
	}
	return buf.String()
}

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

// legacyParam decodes a percent-encoded path parameter and NFC-normalises it.
// On a malformed escape it returns the raw value, which simply won't match.
func legacyParam(raw string) string {
	v, err := url.PathUnescape(raw)
	if err != nil {
		return raw
	}
	return norm.NFC.String(v)
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
