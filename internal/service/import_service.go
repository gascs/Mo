package service

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mo/internal/database"
	"mo/internal/model"
)

// ImportResult holds the count of imported posts.
type ImportResult struct {
	PostsImported int      `json:"posts_imported"`
	MediaImported int      `json:"media_imported"`
	Errors        []string `json:"errors"`
}

// parsedPost holds data parsed from a markdown file's front matter.
type parsedPost struct {
	Title       string
	Slug        string
	Summary     string
	Category    string
	Tags        string
	IsDraft     bool
	IsPinned    bool
	IsPrivate   bool
	PublishedAt string
	CreatedAt   string
}

// ImportZIP reads a ZIP archive and imports posts and media.
func ImportZIP(zipPath, uploadsDir string) (*ImportResult, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	result := &ImportResult{}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		cleanName := filepath.ToSlash(f.Name)

		// Handle posts (*.md files in posts/ or root)
		if (strings.HasPrefix(cleanName, "posts/") || !strings.Contains(cleanName, "/")) && strings.HasSuffix(cleanName, ".md") {
			rc, err := f.Open()
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("open %s: %v", f.Name, err))
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("read %s: %v", f.Name, err))
				continue
			}

			post, mdContent, err := parseFrontMatter(string(data))
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("parse %s: %v", f.Name, err))
				continue
			}

			// Render markdown to HTML
			html, err := RenderMarkdown(mdContent)
			if err != nil {
				html = mdContent
			}

			if post.Summary == "" {
				post.Summary = ExtractSummary(html, 200)
			}

			// Parse time values
			now := time.Now()
			var publishedAt sql.NullTime
			if post.PublishedAt != "" {
				if t, err := time.Parse("2006-01-02", post.PublishedAt); err == nil {
					publishedAt = sql.NullTime{Time: t, Valid: true}
				} else if t, err := time.Parse(time.RFC3339, post.PublishedAt); err == nil {
					publishedAt = sql.NullTime{Time: t, Valid: true}
				}
			}
			createdAt := now
			if post.CreatedAt != "" {
				if t, err := time.Parse("2006-01-02", post.CreatedAt); err == nil {
					createdAt = t
				} else if t, err := time.Parse(time.RFC3339, post.CreatedAt); err == nil {
					createdAt = t
				}
			}

			// Generate slug if empty
			slug := post.Slug
			if slug == "" {
				slug = GenerateSlug(post.Title)
			}

			// Insert into database
			_, err = database.DB.Exec(`
				INSERT OR REPLACE INTO posts (id, title, slug, content, content_html, summary, category, tags,
					is_pinned, is_draft, is_private, published_at, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				model.NewULID(), post.Title, slug, mdContent, html, post.Summary,
				post.Category, post.Tags, post.IsPinned, post.IsDraft, post.IsPrivate,
				nullTimeOrNil(publishedAt), createdAt, now,
			)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("insert %s: %v", f.Name, err))
				continue
			}
			result.PostsImported++
		}

		// Handle media files (uploads/ directory)
		if strings.HasPrefix(cleanName, "uploads/") && !strings.HasSuffix(cleanName, ".md") {
			rel := strings.TrimPrefix(cleanName, "uploads/")
			dest := filepath.Join(uploadsDir, rel)
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("mkdir %s: %v", dest, err))
				continue
			}
			rc, err := f.Open()
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("open media %s: %v", f.Name, err))
				continue
			}
			dstFile, err := os.Create(dest)
			if err != nil {
				rc.Close()
				result.Errors = append(result.Errors, fmt.Sprintf("create media %s: %v", dest, err))
				continue
			}
			_, err = io.Copy(dstFile, rc)
			rc.Close()
			dstFile.Close()
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("write media %s: %v", dest, err))
				continue
			}

			// Insert into media table
			info, _ := os.Stat(dest)
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			database.DB.Exec(`
				INSERT OR REPLACE INTO media (id, file_name, original_name, file_path, file_size, mime_type)
				VALUES (?, ?, ?, ?, ?, ?)`,
				model.NewULID(), filepath.Base(rel), filepath.Base(rel), filepath.ToSlash(rel), size, "application/octet-stream",
			)
			result.MediaImported++
		}
	}

	return result, nil
}

// parseFrontMatter extracts YAML front matter and markdown content from raw text.
func parseFrontMatter(raw string) (*parsedPost, string, error) {
	post := &parsedPost{
		Tags:     "[]",
		Category: "tech",
	}

	if !strings.HasPrefix(strings.TrimSpace(raw), "---") {
		return post, raw, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Scan() // skip first "---"

	var fmLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		fmLines = append(fmLines, line)
	}

	for _, line := range fmLines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")

		switch key {
		case "title":
			post.Title = value
		case "slug":
			post.Slug = value
		case "summary":
			post.Summary = value
		case "category":
			post.Category = value
		case "tags":
			post.Tags = value
		case "is_draft":
			post.IsDraft = value == "true"
		case "is_pinned":
			post.IsPinned = value == "true"
		case "is_private":
			post.IsPrivate = value == "true"
		case "published_at":
			post.PublishedAt = value
		case "created_at":
			post.CreatedAt = value
		}
	}

	// Get content after front matter
	idx := strings.Index(raw, "---\n")
	if idx >= 0 {
		idx2 := strings.Index(raw[idx+4:], "---\n")
		if idx2 >= 0 {
			content := raw[idx+4+idx2+4:]
			return post, strings.TrimSpace(content), nil
		}
	}
	return post, raw, nil
}

func nullTimeOrNil(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time
	}
	return nil
}
