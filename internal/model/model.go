package model

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// User represents the blog owner/admin.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	TOTPSecret   string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Post represents a blog post or treehole entry.
type Post struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Slug                 string    `json:"slug"`
	Content              string    `json:"content"`
	ContentHTML          string    `json:"content_html"`
	Summary              string    `json:"summary"`
	Category             string    `json:"category"`
	Tags                 string    `json:"tags"` // JSON array string
	IsPinned             bool      `json:"is_pinned"`
	IsDraft              bool      `json:"is_draft"`
	IsPrivate            bool      `json:"is_private"`
	PrivatePasswordHash  string    `json:"-"`
	PublishedAt          *time.Time `json:"published_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at"`
}

// Comment represents a reader comment.
type Comment struct {
	ID          string    `json:"id"`
	PostID      string    `json:"post_id"`
	ParentID    string    `json:"parent_id"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"-"`
	AuthorURL   string    `json:"author_url"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	UserAgent   string    `json:"-"`
	IPAddress   string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

// Media represents an uploaded file.
type Media struct {
	ID           string    `json:"id"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	CreatedAt    time.Time `json:"created_at"`
}

// Setting is a key-value pair for runtime configuration overrides.
type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// LoginRequest is the payload for POST /api/v1/auth/login.
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is the response for a successful login.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

// SetupRequest is the payload for the initialization wizard.
type SetupRequest struct {
	Username     string `json:"username" binding:"required"`
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
	SiteTitle    string `json:"site_title" binding:"required"`
	SiteSubtitle string `json:"site_subtitle"`
	SiteDesc     string `json:"site_desc"`
}

// SetupStatusResponse indicates whether setup has been completed.
type SetupStatusResponse struct {
	SetupRequired bool `json:"setup_required"`
}

// ChangePasswordRequest is the payload for password change.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// APIError represents a unified error response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewULID generates a new ULID string.
func NewULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// --- Post request/response types ---

type PostListParams struct {
	Page     int    `form:"page"`
	PerPage  int    `form:"per_page"`
	Category string `form:"category"`
	Status   string `form:"status"` // published, draft, trashed
	Tag      string `form:"tag"`
	Search   string `form:"search"`
}

type PostListResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	PerPage int  `json:"per_page"`
}

type CreatePostRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Tags     string `json:"tags"` // JSON array string
	IsDraft  bool   `json:"is_draft"`
	Summary  string `json:"summary"`
	Slug     string `json:"slug"`
}

type UpdatePostRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Category   string `json:"category"`
	Tags       string `json:"tags"`
	IsDraft    *bool  `json:"is_draft"`
	IsPinned   *bool  `json:"is_pinned"`
	IsPrivate  *bool  `json:"is_private"`
	Summary    string `json:"summary"`
	Slug       string `json:"slug"`
}

type MediaListResponse struct {
	Media   []Media `json:"media"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
}

type TagItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DashboardStats struct {
	TotalPosts      int `json:"total_posts"`
	PublishedPosts  int `json:"published_posts"`
	DraftPosts      int `json:"draft_posts"`
	TreeholePosts   int `json:"treehole_posts"`
}

// --- Comment types ---

type CommentListParams struct {
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
	Status  string `form:"status"`
	PostID  string `form:"post_id"`
}

type CommentListResponse struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}
