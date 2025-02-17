package firefox

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/db"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/interfaces"
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
	log.Debug("Getting Firefox bookmarks", "plugin", fp.GetName())

	fpConfig := fp.GetConfig()
	log.Debug("Got Firefox config", "config", fpConfig)

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
		return bookmark.Bookmarks{}
	}

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
	
	// Check if places.sqlite exists
	source, err := os.Open(profile_path + "/places.sqlite")
	if err != nil {
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

	sqlStmt := `SELECT moz_bookmarks.id, moz_bookmarks.parent, moz_bookmarks.type, moz_bookmarks.title, moz_places.url, moz_places.description
	FROM moz_bookmarks LEFT JOIN moz_places ON moz_bookmarks.fk=moz_places.id`

	rows, err := sqlDB.Query(sqlStmt)
	if err != nil {
		log.Debug("%q: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	var bookmarks []mozBookmark
	for rows.Next() {
		var row mozBookmark
		err = rows.Scan(&row.Id, &row.Parent, &row.Typ, &row.Title, &row.Url, &row.Description)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		bookmarks = append(bookmarks, row)
		log.Debug("Data", "row", row)
	}

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

	return sql.Open("sqlite3", dst.Name())
}

func getFavicon(sqlDB *sql.DB, url string) ([]byte, error) {
	log := logger.GetLogger()
	var iconData []byte

	// First check if the URL exists in moz_pages_w_icons
	var pageID int64
	err := sqlDB.QueryRow("SELECT id FROM moz_pages_w_icons WHERE page_url = ?", url).Scan(&pageID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Debug("Error checking page URL", "error", err)
		} else {
			log.Debug("No page found for URL", "url", url)
		}
		return nil, nil
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

	// Get the icon data using the correct joins
	query := `
		SELECT ic.data, ic.width
		FROM moz_icons ic
		WHERE ic.id = ?
		ORDER BY ic.width DESC
		LIMIT 1
	`

	var width int64
	for _, iconID := range iconIDs {
		err = sqlDB.QueryRow(query, iconID).Scan(&iconData, &width)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Debug("Error getting icon data", "error", err, "icon_id", iconID)
			}
			continue
		}
		log.Debug("Found icon data", "url", url, "icon_id", iconID, "width", width, "size", len(iconData))
		break
	}

	return iconData, nil
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
