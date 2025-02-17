package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/internal/logger"
)

var DB *gorm.DB

// CacheDir returns the path to the favicon cache directory
func CacheDir() (string, error) {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %v", err)
		}
		cacheHome = filepath.Join(home, ".cache")
	}

	cacheDir := filepath.Join(cacheHome, "rofi-bookmarks")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %v", err)
	}

	return cacheDir, nil
}

func ConnectDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("bookmarks.db"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err == nil {
		err = DB.AutoMigrate(&Tag{}, &Favicon{}, &Bookmark{})
	}
	return err
}

func CloseDatabase() error {
	db, err := DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func GetBookmarks() (bookmark.Bookmarks, error) {
	var dbBookmarks []Bookmark
	err := DB.Model(&Bookmark{}).Preload("Tags").Find(&dbBookmarks).Error
	if err != nil {
		return nil, err
	}

	// Convert db.Bookmark to bookmark.Bookmark
	var bookmarks bookmark.Bookmarks
	for _, b := range dbBookmarks {
		bookmark := bookmark.Bookmark{
			Title:       b.Title,
			Path:        b.Path,
			Description: b.Description,
			URI:         b.URI,
			Domain:      b.Domain,
			Source:      b.Source,
		}

		// Try to get favicon path if URI exists
		if bookmark.URI != "" {
			if iconPath, err := GetIconPath(bookmark.URI); err == nil && iconPath != "" {
				bookmark.Icon = iconPath
			}
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func SaveBookmark(bookmark bookmark.Bookmark) error {
	dbBookmark := Bookmark{
		Title:       bookmark.Title,
		Path:        bookmark.Path,
		Description: bookmark.Description,
		URI:         bookmark.URI,
		Domain:      bookmark.Domain,
		Source:      bookmark.Source,
	}
	return DB.Save(&dbBookmark).Error
}

// UpdateBookmarks replaces all bookmarks in the database with new ones
func UpdateBookmarks(bookmarks bookmark.Bookmarks) error {
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Only delete bookmarks, preserve favicons
	if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Bookmark{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Convert bookmark.Bookmarks to db.Bookmark
	var dbBookmarks []Bookmark
	for _, b := range bookmarks {
		dbBookmark := Bookmark{
			Title:       b.Title,
			Path:        b.Path,
			Description: b.Description,
			URI:         b.URI,
			Domain:      b.Domain,
			Source:      b.Source,
		}
		dbBookmarks = append(dbBookmarks, dbBookmark)
	}

	// Save new bookmarks
	if err := tx.Create(&dbBookmarks).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetFaviconByDomain retrieves a favicon from the database by domain
func GetFaviconByDomain(domain string) (*Favicon, error) {
	log := logger.GetLogger()
	var favicon Favicon
	result := DB.Where("domain = ?", domain).First(&favicon)
	if result.Error == gorm.ErrRecordNotFound {
		log.Debug("No favicon found in database", "domain", domain)
		return nil, nil
	}
	if result.Error != nil {
		log.Debug("Error getting favicon from database", "domain", domain, "error", result.Error)
		return nil, result.Error
	}
	log.Debug("Found favicon in database", "domain", domain, "size", len(favicon.Data))
	return &favicon, nil
}

// SaveFavicon stores a favicon in the database
func SaveFavicon(data []byte, urlStr string) (*Favicon, error) {
	log := logger.GetLogger()

	// Parse URL to get domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Debug("Error parsing URL", "url", urlStr, "error", err)
		return nil, err
	}
	domain := parsedURL.Host

	// Check if favicon already exists
	existing, err := GetFaviconByDomain(domain)
	if err != nil {
		log.Debug("Error checking existing favicon", "domain", domain, "error", err)
		return nil, err
	}
	if existing != nil {
		log.Debug("Using existing favicon", "domain", domain)
		return existing, nil
	}

	// Create new favicon
	favicon := &Favicon{
		Data:   data,
		Domain: domain,
	}

	log.Debug("Saving new favicon to database", "domain", domain, "size", len(data))
	err = DB.Create(favicon).Error
	if err != nil {
		log.Debug("Error saving favicon to database", "domain", domain, "error", err)
		return nil, err
	}

	log.Debug("Successfully saved favicon to database", "domain", domain)
	return favicon, nil
}

// SaveAndCacheIcon stores the icon in both the database and filesystem cache
// Returns the path to the cached file for use with rofi
func SaveAndCacheIcon(iconData []byte, urlStr string) (string, error) {
	log := logger.GetLogger()

	if len(iconData) == 0 {
		log.Debug("No icon data provided", "url", urlStr)
		return "", fmt.Errorf("no icon data provided")
	}

	// Save to database first
	favicon, err := SaveFavicon(iconData, urlStr)
	if err != nil {
		log.Debug("Could not save favicon to database", "url", urlStr, "error", err)
		return "", fmt.Errorf("could not save favicon to database: %v", err)
	}

	// Get cache directory for filesystem storage
	cacheDir, err := CacheDir()
	if err != nil {
		log.Debug("Could not get cache directory", "error", err)
		return "", err
	}

	// Create hash of icon data for filename
	hash := sha256.Sum256(iconData)
	filename := hex.EncodeToString(hash[:])
	iconPath := filepath.Join(cacheDir, filename)

	// Check if icon already exists in filesystem cache
	if _, err := os.Stat(iconPath); err == nil {
		log.Debug("Icon already exists in cache", "path", iconPath)
		return iconPath, nil
	}

	// Write icon to filesystem cache
	log.Debug("Writing icon to cache", "path", iconPath, "size", len(favicon.Data))
	if err := os.WriteFile(iconPath, favicon.Data, 0644); err != nil {
		log.Debug("Could not write icon to cache", "error", err)
		return "", fmt.Errorf("could not write icon to cache: %v", err)
	}

	log.Debug("Successfully cached icon", "path", iconPath)
	return iconPath, nil
}

// GetIconPath returns the filesystem path for a favicon, fetching from database if needed
func GetIconPath(urlStr string) (string, error) {
	log := logger.GetLogger()

	// Parse URL to get domain
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Debug("Error parsing URL", "url", urlStr, "error", err)
		return "", err
	}
	domain := parsedURL.Host

	// Try to get favicon from database
	favicon, err := GetFaviconByDomain(domain)
	if err != nil {
		log.Debug("Error getting favicon from database", "domain", domain, "error", err)
		return "", err
	}
	if favicon == nil {
		log.Debug("No favicon found in database", "domain", domain)
		return "", nil
	}

	// Cache the icon data to filesystem if it exists in database
	iconPath, err := SaveAndCacheIcon(favicon.Data, urlStr)
	if err != nil {
		log.Debug("Error caching favicon", "domain", domain, "error", err)
		return "", err
	}

	log.Debug("Successfully got icon path", "domain", domain, "path", iconPath)
	return iconPath, nil
}
