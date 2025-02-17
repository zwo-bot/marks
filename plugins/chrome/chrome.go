package chrome

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	neturl "net/url"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/internal/favicon"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins/interfaces"
)

type ChromePlugin struct {
	Config interfaces.PluginConfig
}

type ChromeBookmark struct {
	DateAdded    string            `json:"date_added"`
	GUID         string            `json:"guid"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	URL          string            `json:"url,omitempty"`
	Children     []ChromeBookmark  `json:"children,omitempty"`
}

type ChromeBookmarks struct {
	Checksum string                    `json:"checksum"`
	Roots    map[string]ChromeBookmark `json:"roots"`
}

func (c *ChromePlugin) GetName() string {
	return "Chrome"
}

func (c *ChromePlugin) GetConfig() interfaces.PluginConfig {
	return c.Config
}

func (c *ChromePlugin) SetConfig(cc interfaces.PluginConfig) {
	c.Config = cc
}

func (c *ChromePlugin) GetBookmarks() bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks
	log := logger.GetLogger()
	log.With("plugin", c.GetName())

	chromeConfig, ok := c.Config.(*ChromeConfig)
	if !ok {
		log.Error("Configuration is not of type *ChromeConfig")
		return bookmark.Bookmarks{}
	}

	err := chromeConfig.Load()
	if err != nil {
		log.Error("Error loading configuration", "error", err)
		return bookmark.Bookmarks{}
	}

	profileDir := filepath.Dir(chromeConfig.ProfilePath) // Get the profile directory
	bookmarksPath := filepath.Join(profileDir, "Bookmarks")
	log.Debug("Reading Chrome bookmarks file", "path", bookmarksPath)

	// Read bookmarks file
	data, err := os.ReadFile(bookmarksPath)
	if err != nil {
		log.Error("Chrome bookmarks file not found", "path", bookmarksPath, "error", err)
		return bookmark.Bookmarks{}
	}

	var chromeBookmarks ChromeBookmarks
	if err := json.Unmarshal(data, &chromeBookmarks); err != nil {
		log.Error("Error parsing bookmarks file", "error", err)
		return bookmark.Bookmarks{}
	}

	// Process each root folder
	for folder, root := range chromeBookmarks.Roots {
		bookmarks = append(bookmarks, processBookmarks(root, folder, profileDir, log)...)
	}

	return bookmarks
}

func processBookmarks(node ChromeBookmark, path string, profilePath string, log *slog.Logger) bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks

	// If it's a URL bookmark, add it
	if node.Type == "url" {
		bookmark := bookmark.Bookmark{
			Title:  node.Name,
			URI:    node.URL,
			Path:   path,
			Source: "Chrome",
		}

		// Parse URL to get domain
		if parsedURL, err := neturl.Parse(node.URL); err == nil {
			bookmark.Domain = parsedURL.Host

			// Try to get favicon from cache
			if iconPath, err := favicon.GetIconPath(node.URL); err == nil && iconPath != "" {
			bookmark.Icon = iconPath
			log.Debug("Got favicon from cache", "title", bookmark.Title, "icon_path", bookmark.Icon)
		} else {
			// Try to get favicon from Chrome's database
			iconData, err := getFaviconFromChrome(profilePath, node.URL, log)
			if err != nil {
				log.Debug("Error getting favicon from Chrome", "error", err)
			} else if len(iconData) > 0 {
				iconPath, err := favicon.SaveAndCacheIcon(iconData, node.URL)
				if err != nil {
					log.Debug("Could not cache favicon", "error", err)
				} else {
					bookmark.Icon = iconPath
					log.Debug("Got and cached favicon from Chrome", "title", bookmark.Title, "icon_path", bookmark.Icon)
				}
			} else {
				log.Debug("No favicon found in Chrome", "title", bookmark.Title, "url", node.URL)
			}
			}
		}

		bookmarks = append(bookmarks, bookmark)
	}

	// Process children recursively
	for _, child := range node.Children {
		newPath := path
		if child.Type == "folder" {
			if newPath == "" {
				newPath = child.Name
			} else {
				newPath = filepath.Join(newPath, child.Name)
			}
		}
		bookmarks = append(bookmarks, processBookmarks(child, newPath, profilePath, log)...)
	}

	return bookmarks
}

func copyAndOpenDB(sourcePath string, prefix string) (*sql.DB, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer source.Close()

	dst, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(dst.Name())

	if _, err := io.Copy(dst, source); err != nil {
		return nil, fmt.Errorf("error copying database: %v", err)
	}

	// Open the database with SQLite
	db, err := sql.Open("sqlite3", dst.Name())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Configure SQLite connection for better handling
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting journal mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting synchronous mode: %v", err)
	}

	// Verify the database schema
	var exists bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name IN ('favicons', 'favicon_bitmaps', 'icon_mapping')").Scan(&exists)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error checking schema: %v", err)
	}
	if !exists {
		db.Close()
		return nil, fmt.Errorf("required tables not found in favicons database")
	}

	return db, nil
}

func getFaviconFromChrome(profilePath string, url string, log *slog.Logger) ([]byte, error) {
	// Chrome's Favicons database is in the profile directory
	faviconDBPath := filepath.Join(profilePath, "Favicons")
	log.Debug("Opening Chrome favicons database", "path", faviconDBPath)
	
	// Create a temporary copy of the database since Chrome might have it locked
	db, err := copyAndOpenDB(faviconDBPath, "chrome_favicons")
	if err != nil {
		return nil, fmt.Errorf("error opening favicons database: %v", err)
	}
	defer db.Close()

	// First try the direct join query
	var iconData []byte
	err = db.QueryRow(`
		SELECT fb.image_data
		FROM favicon_bitmaps fb
		JOIN icon_mapping im ON fb.icon_id = im.icon_id
		WHERE im.page_url = ?
		ORDER BY fb.width DESC
		LIMIT 1
	`, url).Scan(&iconData)

	if err == nil {
		return iconData, nil
	}

	if err != sql.ErrNoRows {
		log.Debug("Error in direct favicon query", "error", err)
	}

	// Fallback to a more lenient query that matches URL patterns
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	// Try to match the domain
	err = db.QueryRow(`
		SELECT fb.image_data
		FROM favicon_bitmaps fb
		JOIN icon_mapping im ON fb.icon_id = im.icon_id
		WHERE im.page_url LIKE ?
		ORDER BY fb.width DESC
		LIMIT 1
	`, "%"+parsedURL.Host+"%").Scan(&iconData)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return iconData, err
}
