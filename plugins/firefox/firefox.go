package firefox

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/db"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins/interfaces"
)

var mozBookmarks []mozBookmark

type FirefoxPlugin struct {
	Config interfaces.PluginConfig
}

type mozBookmark struct {
	Id          int
	Parent      int
	Title       sql.NullString
	Description sql.NullString
	Typ         int
	Url         sql.NullString
	IconPath    sql.NullString
	Tags        sql.NullString
}

func (fp *FirefoxPlugin) GetName() string {
	return "Firefox"
}

func (fp *FirefoxPlugin) GetConfig() interfaces.PluginConfig {
	return fp.Config
}

func (fp *FirefoxPlugin) SetConfig(fc interfaces.PluginConfig) {
	fp.Config = fc
}

func (fp *FirefoxPlugin) GetBookmarks() bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks

	log := logger.GetLogger()
	log.Debug("Starting Firefox bookmark retrieval")
	log.Debug("Getting Firefox bookmarks", "plugin", fp.GetName())

	fpConfig := fp.GetConfig()
	log.Debug("Got Firefox config", "config", fpConfig)
	log.Debug("Config type", "type", fmt.Sprintf("%T", fpConfig))

	firefoxConfig, ok := fpConfig.(*FirefoxConfig)
	if !ok {
		// Handle the error or unexpected type
		log.Error("Configuration is not of type *FirefoxConfig")
		return bookmark.Bookmarks{}
	}
	log.Debug("Firefox config before Load()", "profile_path", firefoxConfig.ProfilePath)

	err := firefoxConfig.Load()
	if err != nil {
		log.Error("Error loading configuration", "error", err)
		return bookmark.Bookmarks{}
	}
	log.Debug("Firefox config after Load()", "profile_path", firefoxConfig.ProfilePath)

	// Get bookmarks from the configured profile path
	moz_bookmarks, err := getMozBookmarks(firefoxConfig.ProfilePath)
	if err != nil {
		log.Debug("Could not get Firefox bookmarks", "error", err)
		log.Debug("Failed at getMozBookmarks", "profile_path", firefoxConfig.ProfilePath)
		return bookmark.Bookmarks{}
	}
	log.Debug("Retrieved Mozilla bookmarks", "count", len(moz_bookmarks))

	for _, mozBookmark := range moz_bookmarks {
		var bookmark bookmark.Bookmark
		has_url := false

		if mozBookmark.Url.Valid {
			bookmark.URI = mozBookmark.Url.String
			has_url = true

			// Parse URL to get domain
			if parsedURL, err := url.Parse(bookmark.URI); err == nil {
				bookmark.Domain = parsedURL.Host
			}
		}

		if mozBookmark.Title.Valid {
			bookmark.Title = mozBookmark.Title.String
		}

		if mozBookmark.Description.Valid {
			bookmark.Description = mozBookmark.Description.String
		}

		// Process tags if available
		if mozBookmark.Tags.Valid && mozBookmark.Tags.String != "" {
			// Split tags string into slice and trim whitespace
			tags := strings.Split(mozBookmark.Tags.String, ",")
			bookmark.Tags = make([]string, 0, len(tags))
			for _, tag := range tags {
				if trimmed := strings.TrimSpace(tag); trimmed != "" {
					bookmark.Tags = append(bookmark.Tags, trimmed)
				}
			}
		}

		bookmark.Path = getPath(mozBookmark)
		bookmark.Source = fp.GetName()

		if has_url {
			// Get favicon from Firefox's database and store it
			if mozBookmark.IconPath.Valid && mozBookmark.IconPath.String != "" {
				bookmark.Icon = mozBookmark.IconPath.String
				log.Debug("Setting favicon for bookmark", "title", bookmark.Title, "icon_path", bookmark.Icon)
			} else {
				// Try to get favicon from cache
				if iconPath, err := db.GetIconPath(bookmark.URI); err == nil && iconPath != "" {
					bookmark.Icon = iconPath
					log.Debug("Got favicon from cache", "title", bookmark.Title, "icon_path", bookmark.Icon)
				}
			}
			bookmarks = append(bookmarks, bookmark)
		}
	}

	return bookmarks
}

func getMozBookmarks(profile_path string) ([]mozBookmark, error) {
	log := logger.GetLogger()
	log.Debug("Starting getMozBookmarks", "profile_path", profile_path)

	// Check if places.sqlite exists
	placesPath := profile_path + "/places.sqlite"
	log.Debug("Opening places.sqlite", "path", placesPath)
	source, err := os.Open(placesPath)
	if err != nil {
		log.Debug("Failed to open places.sqlite", "error", err)
		return nil, err
	}
	defer source.Close()

	// Create temporary copy of places.sqlite
	dst, err := os.CreateTemp("", "ff_places")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(dst.Name())

	if _, err := io.Copy(dst, source); err != nil {
		return nil, fmt.Errorf("error copying database: %v", err)
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite3", dst.Name())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer sqlDB.Close()

	sqlStmt := `
WITH tags AS (
    -- Get all tag definitions (type=2, parent=4 for tags folder)
    SELECT id, title 
    FROM moz_bookmarks 
    WHERE type = 2 AND parent = 4
),
bookmark_tag_links AS (
    -- Get bookmark-to-tag relationships
    SELECT b.fk as place_id, 
           GROUP_CONCAT(t.title) as tags
    FROM moz_bookmarks b
    JOIN moz_bookmarks bt ON bt.fk = b.fk  -- Find tag links for same URL
    JOIN tags t ON bt.parent = t.id        -- Get tag names
    WHERE b.type = 1  -- Regular bookmarks
    GROUP BY b.fk
)
SELECT 
    b.id,
    b.parent,
    b.type,
    b.title,
    p.url,
    p.description,
    btl.tags
FROM moz_bookmarks b
LEFT JOIN moz_places p ON b.fk = p.id
LEFT JOIN bookmark_tag_links btl ON p.id = btl.place_id
WHERE b.type = 1  -- Only regular bookmarks
  AND b.title IS NOT NULL  -- Skip tag link entries`

	log.Debug("Executing SQL query", "query", sqlStmt)
	rows, err := sqlDB.Query(sqlStmt)
	if err != nil {
		log.Debug("SQL query failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var bookmarks []mozBookmark
	rowCount := 0
	for rows.Next() {
		rowCount++
		var row mozBookmark
		err = rows.Scan(&row.Id, &row.Parent, &row.Typ, &row.Title, &row.Url, &row.Description, &row.Tags)
		if err != nil {
			log.Error("Error scanning row", "error", err)
			continue
		}
		bookmarks = append(bookmarks, row)
		log.Debug("Processed bookmark row", "id", row.Id, "title", row.Title, "url", row.Url)
	}
	log.Debug("Finished processing rows", "total_rows", rowCount, "valid_bookmarks", len(bookmarks))

	// Get favicons
	faviconsDB, err := copyAndOpenDB(profile_path+"/favicons.sqlite", "ff_favicons")
	if err != nil {
		log.Error("Error opening favicons database", "error", err)
		return bookmarks, nil // Return bookmarks without icons
	}
	defer faviconsDB.Close()

	log.Debug("Successfully opened favicons.sqlite", "path", profile_path+"/favicons.sqlite")

	// Add icons to bookmarks
	for i, bookmark := range bookmarks {
		if bookmark.Url.Valid {
			log.Debug("Getting favicon for URL", "url", bookmark.Url.String)
			iconData, err := getFavicon(faviconsDB, bookmark.Url.String)
			if err != nil {
				log.Debug("Could not get favicon", "url", bookmark.Url.String, "error", err)
				continue
			}
			if len(iconData) > 0 {
				log.Debug("Found favicon data", "url", bookmark.Url.String, "size", len(iconData))
				iconPath, err := db.SaveAndCacheIcon(iconData, bookmark.Url.String)
				if err != nil {
					log.Debug("Could not cache favicon", "error", err)
					continue
				}
				log.Debug("Saved favicon", "url", bookmark.Url.String, "path", iconPath)
				bookmarks[i].IconPath = sql.NullString{String: iconPath, Valid: true}
			} else {
				log.Debug("No favicon data found", "url", bookmark.Url.String)
			}
		}
	}

	mozBookmarks = bookmarks
	return bookmarks, nil
}

func copyAndOpenDB(sourcePath string, prefix string) (*sql.DB, error) {
	log := logger.GetLogger()
	log.Debug("Starting database copy operation", "source", sourcePath)

	source, err := os.Open(sourcePath)
	if err != nil {
		log.Debug("Failed to open source database", "error", err)
		return nil, err
	}
	defer source.Close()

	dst, err := os.CreateTemp("", prefix)
	if err != nil {
		log.Debug("Failed to create temp file", "error", err)
		return nil, fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(dst.Name())

	if _, err := io.Copy(dst, source); err != nil {
		log.Debug("Failed to copy database", "error", err)
		return nil, fmt.Errorf("error copying database: %v", err)
	}

	// Ensure all writes are flushed to disk
	if err := dst.Sync(); err != nil {
		log.Debug("Failed to sync database file", "error", err)
		return nil, fmt.Errorf("error syncing database: %v", err)
	}

	// Close the file to ensure all writes are complete
	if err := dst.Close(); err != nil {
		log.Debug("Failed to close database file", "error", err)
		return nil, fmt.Errorf("error closing database: %v", err)
	}

	log.Debug("Opening copied database", "path", dst.Name())
	db, err := sql.Open("sqlite3", dst.Name())
	if err != nil {
		log.Debug("Failed to open database", "error", err)
		return nil, err
	}

	// Initialize the database connection with proper settings
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Debug("Failed to set pragma", "pragma", pragma, "error", err)
			db.Close()
			return nil, fmt.Errorf("error setting pragma %s: %v", pragma, err)
		}
	}

	// Verify tables exist
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='moz_pages_w_icons'").Scan(&tableCount)
	if err != nil {
		log.Debug("Failed to verify tables", "error", err)
		db.Close()
		return nil, fmt.Errorf("error verifying tables: %v", err)
	}
	if tableCount == 0 {
		log.Debug("Required table not found in copied database", "table", "moz_pages_w_icons")
		db.Close()
		return nil, fmt.Errorf("required table moz_pages_w_icons not found")
	}

	log.Debug("Successfully initialized database connection")
	return db, nil
}

func getFavicon(sqlDB *sql.DB, url string) ([]byte, error) {
	log := logger.GetLogger()
	log.Debug("Starting favicon retrieval", "url", url)

	var iconData []byte

	// First try to get the page ID and icon data in a single query
	query := `
WITH page AS (
SELECT id FROM moz_pages_w_icons WHERE page_url = ?
)
SELECT DISTINCT ic.data
FROM page
JOIN moz_icons_to_pages ip ON ip.page_id = page.id
JOIN moz_icons ic ON ic.id = ip.icon_id
ORDER BY ic.width DESC
LIMIT 1
`

	log.Debug("Executing favicon query", "query", query)
	err := sqlDB.QueryRow(query, url).Scan(&iconData)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("No favicon found for URL", "url", url)
			return nil, nil
		}
		// Try alternative query if the first one fails
		log.Debug("Primary query failed, trying alternative", "error", err)
		return getFaviconAlternative(sqlDB, url)
	}

	log.Debug("Found favicon data", "url", url, "size", len(iconData))
	return iconData, nil
}

// Alternative method to get favicon if the primary method fails
func getFaviconAlternative(sqlDB *sql.DB, url string) ([]byte, error) {
	log := logger.GetLogger()
	log.Debug("Trying alternative favicon retrieval", "url", url)

	// Try to get page ID first
	var pageID int64
	err := sqlDB.QueryRow("SELECT id FROM moz_pages_w_icons WHERE page_url = ?", url).Scan(&pageID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("No page found for URL", "url", url)
			return nil, nil
		}
		log.Debug("Error getting page ID", "error", err)
		return nil, err
	}
	log.Debug("Found page ID", "url", url, "page_id", pageID)

	// Get icon IDs for this page
	rows, err := sqlDB.Query("SELECT icon_id FROM moz_icons_to_pages WHERE page_id = ?", pageID)
	if err != nil {
		log.Debug("Error getting icon IDs", "error", err)
		return nil, err
	}
	defer rows.Close()

	var iconIDs []int64
	for rows.Next() {
		var iconID int64
		if err := rows.Scan(&iconID); err != nil {
			log.Debug("Error scanning icon ID", "error", err)
			continue
		}
		iconIDs = append(iconIDs, iconID)
		log.Debug("Found icon ID", "icon_id", iconID)
	}

	if len(iconIDs) == 0 {
		log.Debug("No icons found for page", "url", url)
		return nil, nil
	}

	// Try to get icon data for each icon ID
	var iconData []byte
	for _, iconID := range iconIDs {
		err = sqlDB.QueryRow("SELECT data FROM moz_icons WHERE id = ?", iconID).Scan(&iconData)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Debug("Error getting icon data", "error", err, "icon_id", iconID)
			}
			continue
		}
		if len(iconData) > 0 {
			log.Debug("Found icon data", "url", url, "icon_id", iconID, "size", len(iconData))
			return iconData, nil
		}
	}

	log.Debug("No valid icon data found", "url", url)
	return nil, nil
}

func getById(id int) mozBookmark {
	result := mozBookmark{}

	for _, b := range mozBookmarks {
		if id == b.Id {
			result = b
			break
		}
	}
	return result
}

func getPath(mb mozBookmark) string {
	rec := getById(mb.Id)
	parent_id := rec.Parent
	path := ""

	for parent_id > 1 {
		parent := getById(parent_id)
		parent_id = parent.Parent
		path = parent.Title.String + "/" + path
	}
	return path
}
