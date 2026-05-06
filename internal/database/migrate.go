package database

import (
	"fmt"
)

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{1, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			totp_secret TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			title TEXT DEFAULT '',
			slug TEXT NOT NULL UNIQUE,
			content TEXT NOT NULL DEFAULT '',
			content_html TEXT NOT NULL DEFAULT '',
			summary TEXT DEFAULT '',
			category TEXT DEFAULT 'tech' CHECK(category IN ('tech','life','treehole')),
			tags TEXT DEFAULT '[]',
			is_pinned INTEGER DEFAULT 0,
			is_draft INTEGER DEFAULT 1,
			is_private INTEGER DEFAULT 0,
			private_password_hash TEXT DEFAULT '',
			published_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		);

		CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
		CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category);
		CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at);

		CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(title, content, content='posts', content_rowid='rowid');

		CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			post_id TEXT NOT NULL REFERENCES posts(id),
			parent_id TEXT REFERENCES comments(id),
			author_name TEXT NOT NULL,
			author_email TEXT NOT NULL,
			author_url TEXT DEFAULT '',
			content TEXT NOT NULL,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending','approved','spam','trash')),
			user_agent TEXT DEFAULT '',
			ip_address TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);

		CREATE TABLE IF NOT EXISTS media (
			id TEXT PRIMARY KEY,
			file_name TEXT NOT NULL,
			original_name TEXT NOT NULL,
			file_path TEXT NOT NULL UNIQUE,
			file_size INTEGER DEFAULT 0,
			mime_type TEXT DEFAULT '',
			width INTEGER DEFAULT 0,
			height INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT DEFAULT ''
		);
	`},
}

func RunMigrations() error {
	// Create migrations tracking table
	if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS _migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create _migrations: %w", err)
	}

	for _, m := range migrations {
		var done bool
		err := DB.QueryRow("SELECT 1 FROM _migrations WHERE version = ?", m.version).Scan(&done)
		if err == nil {
			continue // already applied
		}

		// Run the migration in a transaction
		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec("INSERT INTO _migrations (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}

	return nil
}
