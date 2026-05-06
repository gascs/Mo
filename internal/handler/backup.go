package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mo/internal/config"
	"mo/internal/service"

	"github.com/gin-gonic/gin"
)

type BackupHandler struct {
	cfg *config.Config
}

func NewBackupHandler(cfg *config.Config) *BackupHandler {
	return &BackupHandler{cfg: cfg}
}

// Export handles full-site export as ZIP download.
func (h *BackupHandler) Export(c *gin.Context) {
	tmpDir := os.TempDir()
	fileName := fmt.Sprintf("mo-export-%s.zip", time.Now().Format("2006-01-02-150405"))
	tmpPath := filepath.Join(tmpDir, fileName)

	f, err := os.Create(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	defer f.Close()
	defer os.Remove(tmpPath)

	if err := service.ExportSite(h.cfg.Uploads.Dir, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	f.Close()

	c.FileAttachment(tmpPath, fileName)
}

// Import handles ZIP upload and import of posts and media.
func (h *BackupHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "No file uploaded"}})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Only .zip files are supported"}})
		return
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("mo-import-%d.zip", time.Now().UnixNano()))
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	defer os.Remove(tmpPath)

	result, err := service.ImportZIP(tmpPath, h.cfg.Uploads.Dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TriggerBackup creates a backup ZIP and saves it to the backup directory.
func (h *BackupHandler) TriggerBackup(c *gin.Context) {
	backupDir := "backups"

	path, err := service.TriggerBackup(backupDir, h.cfg.Uploads.Dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Backup created",
		"file":     filepath.Base(path),
		"location": path,
	})
}

// IntegrityCheck runs PRAGMA integrity_check and returns the result.
func (h *BackupHandler) IntegrityCheck(c *gin.Context) {
	ok, result := service.CheckIntegrity()
	status := "healthy"
	if !ok {
		status = "unhealthy"
	}
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"result": result,
	})
}
