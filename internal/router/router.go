package router

import (
	"io/fs"
	"net/http"
	"strings"

	"mo/internal/auth"
	"mo/internal/config"
	"mo/internal/database"
	"mo/internal/handler"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, staticFS fs.FS) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(auth.SecurityHeaders(cfg.Server.Domain))

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	authHandler := handler.NewAuthHandler(cfg.Auth.JWTSecret)
	setupHandler := handler.NewSetupHandler(cfg.Auth.JWTSecret, cfg)
	postHandler := handler.NewPostHandler()
	mediaHandler := handler.NewMediaHandler(cfg.Uploads.Dir)
	settingsHandler := handler.NewSettingsHandler(cfg)
	backupHandler := handler.NewBackupHandler(cfg)

	// --- RSS ---
	r.GET("/rss.xml", handler.RSSFeed(cfg))

	api := r.Group("/api/v1")
	api.Use(auth.GlobalRateLimit())
	{
		api.GET("/healthz", handler.HealthCheck)

		// Public auth
		authGroup := api.Group("/auth")
		authGroup.Use(auth.LoginRateLimit())
		{
			authGroup.POST("/login", authHandler.Login)
		}
		api.POST("/auth/refresh", authHandler.Refresh)
		api.POST("/auth/logout", authHandler.Logout)

		// Setup
		api.GET("/setup/status", setupHandler.Status)
		api.POST("/setup/initialize", setupHandler.Initialize)

		// Public settings (site info + theme)
		api.GET("/settings/public", settingsHandler.PublicSettings)

		// Public frontend
		api.GET("/posts", handler.PublicPostList)
		api.GET("/posts/search", handler.SearchPosts)
		api.GET("/posts/archive", handler.Archive)
		api.GET("/posts/:slug", handler.PublicPostBySlug)
		api.GET("/posts/:slug/comments", handler.GetComments)
		api.POST("/posts/:slug/comments", handler.CreateComment)

		// Public tags
		api.GET("/tags", handler.ListTags)

		// Public static pages
		api.GET("/pages/:slug", handler.GetPage)

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(auth.AuthMiddleware(cfg.Auth.JWTSecret))
		{
			admin.POST("/auth/change-password", authHandler.ChangePassword)
			admin.GET("/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"id": c.GetString("user_id"),
					"username": c.GetString("username"),
				})
			})

			admin.GET("/dashboard", postHandler.Dashboard)

			// Posts
			admin.GET("/posts", postHandler.List)
			admin.POST("/posts", postHandler.Create)
			admin.GET("/posts/:id", postHandler.Get)
			admin.PUT("/posts/:id", postHandler.Update)
			admin.DELETE("/posts/:id", postHandler.Delete)
			admin.DELETE("/posts/:id/permanent", postHandler.HardDelete)
			admin.PUT("/posts/:id/restore", postHandler.Restore)
			admin.PUT("/posts/:id/publish", postHandler.Publish)
			admin.PUT("/posts/:id/pin", postHandler.Pin)

			// Media
			admin.POST("/upload", mediaHandler.Upload)
			admin.GET("/media", mediaHandler.List)
			admin.DELETE("/media/:id", mediaHandler.Delete)

			// Comments
			admin.GET("/comments", handler.AdminListComments)
			admin.PUT("/comments/:id", handler.AdminUpdateCommentStatus)
			admin.DELETE("/comments/:id", handler.AdminDeleteComment)

			// Settings
			admin.GET("/settings", settingsHandler.AdminGetSettings)
			admin.PUT("/settings", settingsHandler.AdminUpdateSettings)

			// Backup & Tools
			admin.GET("/export", backupHandler.Export)
			admin.POST("/import", backupHandler.Import)
			admin.POST("/backup", backupHandler.TriggerBackup)
			admin.GET("/integrity", backupHandler.IntegrityCheck)

			// Pages
			admin.PUT("/pages/:slug", handler.UpdatePage)
		}
	}

	r.GET("/healthz", handler.HealthCheck)
	r.Static("/uploads", cfg.Uploads.Dir)

	if staticFS != nil {
		httpFS := http.FS(staticFS)
		r.GET("/assets/*filepath", func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, httpFS)
		})

		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/") || path == "/healthz" || path == "/rss.xml" {
				c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Not found"}})
				return
			}

			setup, err := database.IsSetup()
			if err == nil && !setup && path == "/" {
				c.Redirect(http.StatusFound, "/setup")
				return
			}

			data, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>Mo Blog</title>
<style>body{background:#0d1117;color:#c9d1d9;font-family:system-ui,-apple-system,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0}div{text-align:center}h1{font-size:2em;color:#58a6ff}p{color:#8b949e}</style>
</head><body><div><h1>Mo</h1><p>A minimal personal blog</p></div></body></html>`)
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}

	return r
}
