package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mo/internal/database"
)

// ExportSite creates a ZIP containing all posts as .md files, uploads/, and a sanitized config.
func ExportSite(uploadsDir string, writer io.Writer) error {
	zw := zip.NewWriter(writer)
	defer zw.Close()

	// Fetch all non-deleted posts
	rows, err := database.DB.Query(`
		SELECT title, slug, content, summary, category, tags, is_draft, is_pinned, is_private,
			published_at, created_at
		FROM posts WHERE deleted_at IS NULL ORDER BY published_at DESC`)
	if err != nil {
		return fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var title, slug, content, summary, category, tags string
		var isDraft, isPinned, isPrivate bool
		var publishedAt, createdAt *string

		if err := rows.Scan(&title, &slug, &content, &summary, &category, &tags, &isDraft, &isPinned, &isPrivate, &publishedAt, &createdAt); err != nil {
			continue
		}

		// Build YAML Front Matter
		fm := buildFrontMatter(title, slug, summary, category, tags, isDraft, isPinned, isPrivate, publishedAt, createdAt)
		fileName := slug + ".md"
		w, _ := zw.Create("posts/" + fileName)
		w.Write([]byte(fm))
		w.Write([]byte("\n"))
		w.Write([]byte(content))
	}

	// Include uploads directory
	_ = filepath.Walk(uploadsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(uploadsDir, path)
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		w, _ := zw.Create("uploads/" + filepath.ToSlash(rel))
		io.Copy(w, f)
		return nil
	})

	// Include sanitized settings as JSON
	settings, _ := GetAllSettings()
	sanitized := make(map[string]string)
	for k, v := range settings {
		// Skip internal/sensitive keys
		if strings.HasPrefix(k, "page:") {
			continue
		}
		sanitized[k] = v
	}
	data, _ := json.MarshalIndent(sanitized, "", "  ")
	w, _ := zw.Create("settings.json")
	w.Write(data)

	return nil
}

func buildFrontMatter(title, slug, summary, category, tags string, isDraft, isPinned, isPrivate bool, publishedAt, createdAt *string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("title: %q\n", title))
	b.WriteString(fmt.Sprintf("slug: %q\n", slug))
	if summary != "" {
		b.WriteString(fmt.Sprintf("summary: %q\n", summary))
	}
	b.WriteString(fmt.Sprintf("category: %q\n", category))
	b.WriteString(fmt.Sprintf("tags: %s\n", tags))
	b.WriteString(fmt.Sprintf("is_draft: %t\n", isDraft))
	b.WriteString(fmt.Sprintf("is_pinned: %t\n", isPinned))
	b.WriteString(fmt.Sprintf("is_private: %t\n", isPrivate))
	if publishedAt != nil {
		b.WriteString(fmt.Sprintf("published_at: %q\n", *publishedAt))
	}
	if createdAt != nil {
		b.WriteString(fmt.Sprintf("created_at: %q\n", *createdAt))
	}
	b.WriteString("---\n")
	return b.String()
}

// --- Auto Backup ---

type backupJob struct {
	enabled  bool
	schedule string // "daily 03:00" etc.
	dir      string
	uploadsDir string
	stopCh   chan struct{}
}

var activeBackup *backupJob

// StartBackupScheduler starts the automatic backup loop if enabled.
func StartBackupScheduler(enabled bool, schedule, backupDir, uploadsDir string) {
	if activeBackup != nil {
		activeBackup.stopCh <- struct{}{}
	}

	if !enabled {
		return
	}

	activeBackup = &backupJob{
		enabled:    true,
		schedule:   schedule,
		dir:        backupDir,
		uploadsDir: uploadsDir,
		stopCh:     make(chan struct{}),
	}

	go func() {
		for {
			next := nextBackupTime(activeBackup.schedule)
			select {
			case <-time.After(time.Until(next)):
				_, _ = createBackup(activeBackup.dir, activeBackup.uploadsDir)
			case <-activeBackup.stopCh:
				return
			}
		}
	}()
}

// TriggerBackup creates a backup immediately and returns the file path.
func TriggerBackup(backupDir, uploadsDir string) (string, error) {
	return createBackup(backupDir, uploadsDir)
}

func createBackup(backupDir, uploadsDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("backup-%s.zip", time.Now().Format("2006-01-02-150405"))
	path := filepath.Join(backupDir, name)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := ExportSite(uploadsDir, f); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}

	// Rotate: keep only last 7 backups
	rotateBackups(backupDir, 7)

	return path, nil
}

func rotateBackups(dir string, keep int) {
	entries, _ := os.ReadDir(dir)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "backup-") && strings.HasSuffix(e.Name(), ".zip") {
			files = append(files, e.Name())
		}
	}
	if len(files) <= keep {
		return
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	for _, f := range files[keep:] {
		os.Remove(filepath.Join(dir, f))
	}
}

func nextBackupTime(schedule string) time.Time {
	now := time.Now()
	// Simple: parse "daily HH:MM" format
	schedule = strings.TrimPrefix(schedule, "daily ")
	target, err := time.Parse("15:04", schedule)
	if err != nil {
		return now.Add(24 * time.Hour)
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), target.Hour(), target.Minute(), 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// --- Integrity Check ---

// CheckIntegrity runs PRAGMA integrity_check on the database.
func CheckIntegrity() (bool, string) {
	var result string
	err := database.DB.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return false, err.Error()
	}
	return result == "ok", result
}
