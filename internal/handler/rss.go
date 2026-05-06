package handler

import (
	"database/sql"
	"encoding/xml"
	"net/http"
	"time"

	"mo/internal/config"
	"mo/internal/database"

	"github.com/gin-gonic/gin"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

func RSSFeed(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := cfg.Server.Domain
		if domain == "" {
			domain = c.Request.Host
		}
		baseURL := "http://" + domain
		if c.Request.TLS != nil {
			baseURL = "https://" + domain
		}

		channel := Channel{
			Title:       cfg.Site.Title,
			Link:        baseURL,
			Description: cfg.Site.Description,
			Language:    cfg.Site.Language,
		}

		query := `SELECT title, slug, summary, content_html, published_at, created_at
			FROM posts WHERE deleted_at IS NULL AND is_draft = 0`
		var args []interface{}
		if !cfg.RSS.FullContent {
			query += " AND category != 'treehole'"
		}
		query += " ORDER BY published_at DESC LIMIT 20"

		rows, err := database.DB.Query(query, args...)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error generating RSS")
			return
		}
		defer rows.Close()

		for rows.Next() {
			var title, slug, summary, contentHTML string
			var publishedAt, createdAt sql.NullTime
			if err := rows.Scan(&title, &slug, &summary, &contentHTML, &publishedAt, &createdAt); err != nil {
				continue
			}

			pubDate := time.Now()
			if publishedAt.Valid {
				pubDate = publishedAt.Time
			} else if createdAt.Valid {
				pubDate = createdAt.Time
			}

			desc := summary
			if cfg.RSS.FullContent {
				desc = contentHTML
			}

			channel.Items = append(channel.Items, Item{
				Title:       title,
				Link:        baseURL + "/post/" + slug,
				Description: desc,
				PubDate:     pubDate.Format(time.RFC1123Z),
				GUID:        baseURL + "/post/" + slug,
			})
		}

		rss := RSS{Version: "2.0", Channel: channel}
		data, _ := xml.MarshalIndent(rss, "", "  ")
		c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", append([]byte(xml.Header), data...))
	}
}
