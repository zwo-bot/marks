package firefox

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
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
	log.With("plugin", fp.GetName())

	fpConfig := fp.GetConfig()

	firefoxConfig, ok := fpConfig.(*FirefoxConfig)
	if !ok {
		// Handle the error or unexpected type
		log.Error("Configuration is not of type *FirefoxConfig")
		return bookmark.Bookmarks{}
	}

	err := firefoxConfig.Load()
	if err != nil {
		log.Error("Error loading configuration", "error", err)
		return bookmark.Bookmarks{}
	}

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
	db, err := sql.Open("sqlite3", dst.Name())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	sqlStmt := `SELECT moz_bookmarks.id, moz_bookmarks.parent, moz_bookmarks.type, moz_bookmarks.title, moz_places.url, moz_places.description
	FROM moz_bookmarks LEFT JOIN moz_places ON moz_bookmarks.fk=moz_places.id`

	rows, err := db.Query(sqlStmt)

	if err != nil {
		log.Debug("%q: %s\n", err)
	}

	for rows.Next() {
		var row mozBookmark

		err = rows.Scan(&row.Id, &row.Parent, &row.Typ, &row.Title, &row.Url, &row.Description)

		if err != nil {
			log.Error("%v", err)
			continue
		}
		mozBookmarks = append(mozBookmarks, row)
		log.Debug("Data", "row", row)
	}
	defer rows.Close()

	return mozBookmarks, nil
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
