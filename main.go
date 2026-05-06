package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mo/internal/config"
	"mo/internal/database"
	"mo/internal/router"
	"mo/internal/service"

	"golang.org/x/crypto/acme/autocert"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	port := flag.Int("p", 0, "Server port (overrides config)")
	configPath := flag.String("c", "config.yaml", "Path to config file")
	migrateOnly := flag.Bool("m", false, "Run migrations and exit")
	showVersion := flag.Bool("v", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Mo Blog %s (commit %s) built %s\n", Version, GitCommit, BuildTime)
		return
	}

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}

	// Initialize database
	if err := database.Open(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if *migrateOnly {
		fmt.Println("Migrations completed successfully")
		return
	}

	// Load static files from embedded FS
	var staticFS fs.FS
	subFS, err := fs.Sub(frontend, "web/dist")
	if err == nil {
		staticFS = subFS
	}

	// Setup router
	r := router.Setup(cfg, staticFS)

	// Print startup banner
	setup, _ := database.IsSetup()
	if !setup {
		fmt.Printf("\n  ╔══════════════════════════════════════════╗\n")
		fmt.Printf("  ║       Welcome to Mo Blog!              ║\n")
		fmt.Printf("  ║                                        ║\n")
		fmt.Printf("  ║  Setup required.                       ║\n")
		fmt.Printf("  ║  Open http://localhost:%d/setup      ║\n", cfg.Server.Port)
		fmt.Printf("  ║  to complete initialization.           ║\n")
		fmt.Printf("  ╚══════════════════════════════════════════╝\n\n")
	}

	// Start backup scheduler
	service.StartBackupScheduler(cfg.Backup.Enabled, cfg.Backup.Schedule, "backups", cfg.Uploads.Dir)

	// Start HTTP/HTTPS server
	go func() {
		if cfg.Server.AutoHTTPS && cfg.Server.Domain != "" {
			runAutoHTTPS(cfg, r)
		} else {
			if setup {
				fmt.Printf("\n  Mo Blog is running at http://localhost:%d\n\n", cfg.Server.Port)
			}
			if err := runHTTP(cfg.Server.Port, r); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server if not using auto HTTPS
	httpSrv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
	}
	_ = httpSrv.Shutdown(ctx)

	fmt.Println("Server stopped")
}

func runHTTP(port int, handler http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}

func runAutoHTTPS(cfg *config.Config, handler http.Handler) {
	certDir := "certs"
	os.MkdirAll(certDir, 0700)

	m := &autocert.Manager{
		Cache:      autocert.DirCache(certDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Server.Domain),
	}

	tlsConfig := &tls.Config{
		GetCertificate: m.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}

	httpsSrv := &http.Server{
		Addr:         ":443",
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// HTTP server for ACME challenges and redirect to HTTPS
	httpSrv := &http.Server{
		Addr: ":80",
		Handler: m.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		fmt.Printf("  HTTPS server starting on :443 for %s\n", cfg.Server.Domain)
		if err := httpsSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			if strings.Contains(err.Error(), "permission denied") && cfg.Server.Port > 0 {
				log.Printf("Cannot bind :443 (need root). Falling back to HTTP on :%d", cfg.Server.Port)
				runHTTP(cfg.Server.Port, handler)
				return
			}
			log.Fatalf("HTTPS server error: %v", err)
		}
	}()

	go func() {
		fmt.Printf("  HTTP redirect server on :80\n")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if strings.Contains(err.Error(), "permission denied") && cfg.Server.Port > 0 {
				return // already handled above
			}
			log.Printf("HTTP server error: %v", err)
		}
	}()
}
