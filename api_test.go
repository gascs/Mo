package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"mo/internal/config"
	"mo/internal/database"
	"mo/internal/router"
)

const testDB = "test_integration.db"

var (
	baseURL string
	token   string
)

func setup() error {
	os.Remove(testDB)

	cfg := &config.Config{
		Site:     config.SiteConfig{Title: "Test"},
		Server:   config.ServerConfig{Port: 18888},
		Auth:     config.AuthConfig{JWTSecret: "test-jwt-secret"},
		Database: config.DatabaseConfig{Path: testDB},
		Uploads:  config.UploadsConfig{Dir: "test_uploads"},
		Theme:    config.ThemeConfig{Name: "dark", AccentColor: "#58a6ff"},
		Comment:  config.CommentConfig{Enabled: true, RequireApproval: true},
	}

	if err := database.Open(cfg.Database.Path); err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	if err := database.RunMigrations(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	// Start test server
	r := router.Setup(cfg, nil)
	go http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), r)
	time.Sleep(500 * time.Millisecond)

	baseURL = fmt.Sprintf("http://localhost:%d/api/v1", cfg.Server.Port)
	return nil
}

func teardown() {
	database.Close()
	os.Remove(testDB)
	os.RemoveAll("test_uploads")
	os.RemoveAll("backups")
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("setup failed: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func jsonPost(url string, body interface{}) (*http.Response, error) {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func jsonGet(url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func jsonPut(url string, body interface{}) (*http.Response, error) {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func readBody(r *http.Response) map[string]interface{} {
	defer r.Body.Close()
	var m map[string]interface{}
	json.NewDecoder(r.Body).Decode(&m)
	return m
}

// --- Tests ---

func TestSetupFlow(t *testing.T) {
	// Status before init
	resp, _ := http.Get(baseURL + "/setup/status")
	if resp.StatusCode != 200 {
		t.Fatal("setup status failed")
	}
	body := readBody(resp)
	if body["setup_required"] != true {
		t.Fatal("expected setup_required=true")
	}

	// Initialize
	resp, err := jsonPost(baseURL+"/setup/initialize", map[string]string{
		"username":   "admin",
		"email":      "admin@test.com",
		"password":   "admin123",
		"site_title": "Test Blog",
	})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	body = readBody(resp)
	if resp.StatusCode != 200 {
		t.Fatalf("init status: %d, body: %v", resp.StatusCode, body)
	}
	if body["access_token"] == nil {
		t.Fatal("expected access_token")
	}
	token = body["access_token"].(string)

	// Setup should now be complete
	resp, _ = http.Get(baseURL + "/setup/status")
	body = readBody(resp)
	if body["setup_required"] != false {
		t.Fatal("expected setup_required=false after init")
	}
}

func TestAuthFlow(t *testing.T) {
	if token == "" {
		t.Skip("setup not completed")
	}

	// Login
	resp, err := jsonPost(baseURL+"/auth/login", map[string]string{
		"login":    "admin",
		"password": "admin123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	body := readBody(resp)
	if body["access_token"] == nil {
		t.Fatal("login: expected access_token")
	}
	token = body["access_token"].(string)

	// Login with wrong password
	resp, _ = jsonPost(baseURL+"/auth/login", map[string]string{
		"login": "admin", "password": "wrong",
	})
	if resp.StatusCode != 401 {
		t.Fatal("expected 401 for bad password")
	}
}

func TestCRUDPost(t *testing.T) {
	if token == "" {
		t.Skip("not authenticated")
	}

	// Create
	resp, err := jsonPost(baseURL+"/admin/posts", map[string]interface{}{
		"title":    "Integration Test Post",
		"content":  "# Hello\n\nIntegration test content.",
		"category": "tech",
		"tags":     `["test"]`,
		"is_draft": false,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	body := readBody(resp)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("create post status: %d, body: %v", resp.StatusCode, body)
	}
	postID := body["id"].(string)
	slug := body["slug"].(string)

	// Get post by slug (public)
	resp, _ = http.Get(baseURL + "/posts/" + slug)
	if resp.StatusCode != 200 {
		t.Fatal("get post by slug failed")
	}

	// Update
	resp, err = jsonPut(baseURL+"/admin/posts/"+postID, map[string]string{
		"title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("update post: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("update post failed")
	}

	// Verify update
	resp, _ = http.Get(baseURL + "/posts/" + slug)
	body = readBody(resp)
	if body["title"] != "Updated Title" {
		t.Fatalf("expected 'Updated Title', got %v", body["title"])
	}

	// Delete (soft)
	resp, _ = jsonReq("DELETE", baseURL+"/admin/posts/"+postID, nil)
	if resp.StatusCode != 200 {
		t.Fatal("delete post failed")
	}
}

func TestSearchAndArchive(t *testing.T) {
	// Search
	resp, _ := http.Get(baseURL + "/posts/search?q=Integration")
	if resp.StatusCode != 200 {
		t.Fatal("search failed")
	}
	body := readBody(resp)
	posts := body["posts"].([]interface{})
	if len(posts) == 0 {
		t.Log("search returned 0 results (may be expected after delete)")
	}

	// Archive
	resp, _ = http.Get(baseURL + "/posts/archive")
	if resp.StatusCode != 200 {
		t.Fatal("archive failed")
	}
}

func TestComments(t *testing.T) {
	if token == "" {
		t.Skip("not authenticated")
	}

	// Create post first
	resp, _ := jsonPost(baseURL+"/admin/posts", map[string]interface{}{
		"title": "Post With Comments", "content": "Content.", "category": "tech",
		"tags": "[]", "is_draft": false,
	})
	body := readBody(resp)
	slug := body["slug"].(string)

	// Submit comment
	resp, _ = jsonPost(baseURL+"/posts/"+slug+"/comments", map[string]string{
		"author_name": "Reader", "author_email": "r@test.com", "content": "Nice post!",
	})
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body = readBody(resp)
		t.Fatalf("create comment failed: status=%d, body=%v", resp.StatusCode, body)
	}
	body = readBody(resp)
	commentID := body["id"].(string)

	// Admin list comments
	resp, _ = jsonGet(baseURL + "/admin/comments")
	body = readBody(resp)
	comments := body["comments"].([]interface{})
	if len(comments) == 0 {
		t.Fatal("expected comments in admin list")
	}

	// Approve
	resp, _ = jsonPut(baseURL+"/admin/comments/"+commentID, map[string]string{"status": "approved"})
	if resp.StatusCode != 200 {
		t.Fatal("approve comment failed")
	}

	// Public comments should now show it
	resp, _ = http.Get(baseURL + "/posts/" + slug + "/comments")
	body = readBody(resp)
	comments = body["comments"].([]interface{})
	if len(comments) == 0 {
		t.Fatal("expected approved comment in public list")
	}
}

func TestRSS(t *testing.T) {
	resp, _ := http.Get("http://localhost:18888/rss.xml")
	if resp.StatusCode != 200 {
		t.Fatal("RSS feed failed")
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(data), "<rss") {
		t.Fatal("RSS response missing <rss> tag")
	}
}

func TestSettings(t *testing.T) {
	if token == "" {
		t.Skip("not authenticated")
	}

	resp, _ := jsonGet(baseURL + "/admin/settings")
	if resp.StatusCode != 200 {
		t.Fatal("get settings failed")
	}

	resp, _ = jsonPut(baseURL+"/admin/settings", map[string]string{
		"site.title": "Updated Site",
	})
	if resp.StatusCode != 200 {
		t.Fatal("update settings failed")
	}

	resp, _ = http.Get(baseURL + "/settings/public")
	body := readBody(resp)
	site := body["site"].(map[string]interface{})
	if site["title"] != "Updated Site" {
		t.Fatalf("expected 'Updated Site', got %v", site["title"])
	}
}

func TestIntegrityCheck(t *testing.T) {
	if token == "" {
		t.Skip("not authenticated")
	}
	resp, _ := jsonGet(baseURL + "/admin/integrity")
	if resp.StatusCode != 200 {
		t.Fatal("integrity check failed")
	}
	body := readBody(resp)
	if body["status"] != "healthy" {
		t.Fatalf("expected healthy, got %v", body["status"])
	}
}

func jsonReq(method, url string, body interface{}) (*http.Response, error) {
	var req *http.Request
	if body != nil {
		data, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}
