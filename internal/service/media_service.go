package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mo/internal/database"
	"mo/internal/model"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"image/svg+xml":   ".svg",
	"application/pdf": ".pdf",
}

var magicBytes = map[string][]byte{
	"image/jpeg":      {0xFF, 0xD8, 0xFF},
	"image/png":       {0x89, 0x50, 0x4E, 0x47},
	"image/gif":       {0x47, 0x49, 0x46},
	"image/webp":      {0x52, 0x49, 0x46, 0x46},
	"application/pdf": {0x25, 0x50, 0x44, 0x46},
}

func UploadFile(file *multipart.FileHeader, uploadDir string) (*model.Media, error) {
	// Validate size (10MB)
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file too large: max 10MB")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	// Read first bytes for magic number check
	header := make([]byte, 512)
	n, _ := src.Read(header)
	header = header[:n]

	// Detect MIME type
	mimeType := http.DetectContentType(header)

	ext, ok := allowedMimeTypes[mimeType]
	if !ok {
		// Also try by extension
		origExt := strings.ToLower(filepath.Ext(file.Filename))
		found := false
		for mt, e := range allowedMimeTypes {
			if e == origExt {
				mimeType = mt
				ext = origExt
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unsupported file type: %s", mimeType)
		}
	}

	// Magic number verification
	if expected, ok := magicBytes[mimeType]; ok {
		if n < len(expected) {
			return nil, fmt.Errorf("file too small to validate")
		}
		for i, b := range expected {
			if header[i] != b {
				return nil, fmt.Errorf("file magic number mismatch")
			}
		}
	}

	// Reset reader
	src.Seek(0, io.SeekStart)

	// Generate filename with random hash
	hash := model.NewULID()
	fileName := hash + ext

	// Create directory by year/month
	now := time.Now()
	subDir := filepath.Join(uploadDir, now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Save file
	filePath := filepath.Join(subDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("save file: %w", err)
	}

	// Relative path for URL
	relPath := filepath.ToSlash(filepath.Join(now.Format("2006"), now.Format("01"), fileName))

	media := &model.Media{
		ID:           model.NewULID(),
		FileName:     fileName,
		OriginalName: file.Filename,
		FilePath:     relPath,
		FileSize:     file.Size,
		MimeType:     mimeType,
	}

	_, err = database.DB.Exec(
		"INSERT INTO media (id, file_name, original_name, file_path, file_size, mime_type) VALUES (?, ?, ?, ?, ?, ?)",
		media.ID, media.FileName, media.OriginalName, media.FilePath, media.FileSize, media.MimeType,
	)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("save media record: %w", err)
	}

	return media, nil
}

func ListMedia(page, perPage int) (*model.MediaListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM media").Scan(&total)

	offset := (page - 1) * perPage
	rows, err := database.DB.Query(
		"SELECT id, file_name, original_name, file_path, file_size, mime_type, width, height, created_at FROM media ORDER BY created_at DESC LIMIT ? OFFSET ?",
		perPage, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(&m.ID, &m.FileName, &m.OriginalName, &m.FilePath, &m.FileSize, &m.MimeType, &m.Width, &m.Height, &m.CreatedAt); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, m)
	}

	if mediaList == nil {
		mediaList = []model.Media{}
	}

	return &model.MediaListResponse{
		Media:   mediaList,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

func DeleteMedia(id string) error {
	var filePath string
	err := database.DB.QueryRow("SELECT file_path FROM media WHERE id = ?", id).Scan(&filePath)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Delete from database first
	if _, err := database.DB.Exec("DELETE FROM media WHERE id = ?", id); err != nil {
		return err
	}

	// Remove file (non-fatal if fails)
	os.Remove(filePath)

	return nil
}
