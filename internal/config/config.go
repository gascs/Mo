package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Site     SiteConfig     `yaml:"site"`
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"`
	Uploads  UploadsConfig  `yaml:"uploads"`
	Theme    ThemeConfig    `yaml:"theme"`
	Comment  CommentConfig  `yaml:"comment"`
	Backup   BackupConfig   `yaml:"backup"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Social   SocialConfig   `yaml:"social"`
	RSS      RSSConfig      `yaml:"rss"`
}

type SiteConfig struct {
	Title       string `yaml:"title"`
	Subtitle    string `yaml:"subtitle"`
	Description string `yaml:"description"`
	Language    string `yaml:"language"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	Domain    string `yaml:"domain"`
	AutoHTTPS bool   `yaml:"auto_https"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type UploadsConfig struct {
	Dir string `yaml:"dir"`
}

type ThemeConfig struct {
	Name        string `yaml:"name"`
	AccentColor string `yaml:"accent_color"`
	FontBody    string `yaml:"font_body"`
	FontCode    string `yaml:"font_code"`
}

type CommentConfig struct {
	Enabled         bool `yaml:"enabled"`
	RequireApproval bool `yaml:"require_approval"`
}

type BackupConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type SocialConfig struct {
	GitHub  string `yaml:"github"`
	Twitter string `yaml:"twitter"`
	Email   string `yaml:"email"`
}

type RSSConfig struct {
	FullContent bool `yaml:"full_content"`
}

func Default() *Config {
	return &Config{
		Site: SiteConfig{
			Title:       "My Blog",
			Subtitle:    "",
			Description: "A minimal personal blog",
			Language:    "zh-CN",
		},
		Server: ServerConfig{
			Port:      8080,
			Domain:    "",
			AutoHTTPS: false,
		},
		Auth: AuthConfig{
			JWTSecret: "",
		},
		Database: DatabaseConfig{
			Path: "data.db",
		},
		Uploads: UploadsConfig{
			Dir: "uploads",
		},
		Theme: ThemeConfig{
			Name:        "dark",
			AccentColor: "#58a6ff",
			FontBody:    "system",
			FontCode:    "jetbrains-mono",
		},
		Comment: CommentConfig{
			Enabled:         true,
			RequireApproval: true,
		},
		Backup: BackupConfig{
			Enabled:  false,
			Schedule: "daily 03:00",
		},
		RSS: RSSConfig{
			FullContent: false,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.Auth.JWTSecret == "" {
		secret, err := generateSecret(32)
		if err != nil {
			return nil, err
		}
		cfg.Auth.JWTSecret = secret
	}

	return cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func generateSecret(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
