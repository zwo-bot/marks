package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// FaviconPathResolver is a function type that resolves favicon paths.
// This avoids circular imports with the favicon package.
var FaviconPathResolver func(urlStr string) (string, error)

var DB *gorm.DB

// DataDir returns the path to the application data directory.
// On Linux: ~/.local/share/marks/
// On macOS: ~/Library/Application Support/marks/
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %v", err)
	}

	var dataDir string
	switch runtime.GOOS {
	case "darwin":
		dataDir = filepath.Join(home, "Library", "Application Support", "marks")
	default: // linux and others
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			xdgData = filepath.Join(home, ".local", "share")
		}
		dataDir = filepath.Join(xdgData, "marks")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("could not create data directory: %v", err)
	}

	return dataDir, nil
}

func ConnectDatabase() error {
	dataDir, err := DataDir()
	if err != nil {
		return fmt.Errorf("could not get data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, "bookmarks.db")

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err == nil {
		err = DB.AutoMigrate(&Tag{}, &Favicon{}, &Bookmark{})
	}
	return err
}

func CloseDatabase() error {
	if DB == nil {
		return nil
	}
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

	var bookmarks bookmark.Bookmarks
	for _, b := range dbBookmarks {
		bm := bookmark.Bookmark{
			Title:       b.Title,
			Path:        b.Path,
			Description: b.Description,
			URI:         b.URI,
			Domain:      b.Domain,
			Source:      b.Source,
			Tags:        make([]string, len(b.Tags)),
		}

		for i, tag := range b.Tags {
			bm.Tags[i] = tag.Name
		}

		// Try to get favicon path if URI exists and resolver is set
		if bm.URI != "" && FaviconPathResolver != nil {
			if iconPath, err := FaviconPathResolver(bm.URI); err == nil && iconPath != "" {
				bm.Icon = iconPath
			}
		}

		bookmarks = append(bookmarks, bm)
	}

	return bookmarks, nil
}

func SaveBookmark(bm bookmark.Bookmark) error {
	dbBookmark := Bookmark{
		Title:       bm.Title,
		Path:        bm.Path,
		Description: bm.Description,
		URI:         bm.URI,
		Domain:      bm.Domain,
		Source:      bm.Source,
	}
	return DB.Save(&dbBookmark).Error
}

// UpdateBookmarks replaces all bookmarks in the database with new ones
func UpdateBookmarks(bms bookmark.Bookmarks) error {
	log := logger.GetLogger()

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Bookmark{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var dbBookmarks []Bookmark
	for _, b := range bms {
		dbBookmark := Bookmark{
			Title:       b.Title,
			Path:        b.Path,
			Description: b.Description,
			URI:         b.URI,
			Domain:      b.Domain,
			Source:      b.Source,
			Tags:        make([]Tag, 0, len(b.Tags)),
		}

		for _, tagName := range b.Tags {
			var tag Tag
			result := tx.FirstOrCreate(&tag, Tag{Name: tagName})
			if result.Error != nil {
				tx.Rollback()
				return result.Error
			}
			dbBookmark.Tags = append(dbBookmark.Tags, tag)
		}

		dbBookmarks = append(dbBookmarks, dbBookmark)
	}

	if err := tx.Create(&dbBookmarks).Error; err != nil {
		tx.Rollback()
		return err
	}

	log.Debug("Created bookmarks with tags", "bookmark_count", len(dbBookmarks))

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

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Debug("Error parsing URL", "url", urlStr, "error", err)
		return nil, err
	}
	domain := parsedURL.Host

	existing, err := GetFaviconByDomain(domain)
	if err != nil {
		log.Debug("Error checking existing favicon", "domain", domain, "error", err)
		return nil, err
	}
	if existing != nil {
		log.Debug("Using existing favicon", "domain", domain)
		return existing, nil
	}

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
