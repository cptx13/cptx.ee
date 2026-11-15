// Photos: folders on disk and thumbnail generation (B-11).

package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/draw"
)

// Photo represents a single image in a photo folder.
type Photo struct {
	ID       string // e.g. "2016_Madrid-001"
	Folder   string // e.g. "2016_Madrid"
	File     string // e.g. "001.jpg"
	URL      string // e.g. "/photos/2016_Madrid/001.jpg"
	ThumbURL string // e.g. "/photos/2016_Madrid/thumb_001.jpg"
	Index    int    // 1-based
	W, H     int    // original pixel size (0 when unknown)
}

// PhotoFolder represents a directory of photos.
type PhotoFolder struct {
	Name     string // directory name e.g. "2016_Madrid"
	Title    string // display title e.g. "2016 Madrid"
	Location string // e.g. "Madrid"
	Year     string // e.g. "2016"
	Photos   []Photo
	Path     string // e.g. "/photos/2016_Madrid/"
	Count    int
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
			w, h := imageSize(filepath.Join(folderPath, file))
			photos = append(photos, Photo{
				ID:       fmt.Sprintf("%s-%03d", folderName, idx),
				Folder:   folderName,
				File:     file,
				URL:      fmt.Sprintf("/photos/%s/%s", folderName, file),
				ThumbURL: fmt.Sprintf("/photos/%s/thumb_%s", folderName, file),
				Index:    idx,
				W:        w,
				H:        h,
			})
		}

		title := strings.ReplaceAll(folderName, "_", " ")
		// Split "2016 Madrid" into year="2016" location="Madrid"
		var year, location string
		if parts := strings.SplitN(title, " ", 2); len(parts) == 2 {
			year = parts[0]
			location = parts[1]
		} else {
			location = title
		}
		folders = append(folders, PhotoFolder{
			Name:     folderName,
			Title:    title,
			Location: location,
			Year:     year,
			Photos:   photos,
			Path:     "/photos/" + folderName + "/",
			Count:    len(photos),
		})
	}

	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})

	return folders, nil
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
