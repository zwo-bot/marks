package plugins

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"gopkg.in/ini.v1"
)

var home, _ = os.UserHomeDir()
var ffDir = home + "/.mozilla/firefox"
var profile = "Default"
var mozBookmarks []mozBookmark

func init() {
	println(ffDir)
	plugin := &firefoxPlugin{
		URL: "test",
	}
	initers = append(initers, plugin)
}

type firefoxPlugin struct {
	URL string
}

type mozBookmark struct {
	Id          int
	Parent      int
	Title       sql.NullString
	Description sql.NullString
	Typ         int
	Url         sql.NullString
}

func (c *firefoxPlugin) GetName() string {
	return "Firefox"
}

func (c *firefoxPlugin) GetBookmarks() bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks
	log := log.With("plugin", c.GetName())
	profile_path := getProfilePath(profile)
	log.Debug("FF profile path", "Path", profile_path)

	moz_bookmarks, _ := getMozBookmarks(profile_path)
	log.Debug("Bookmarks", "content", moz_bookmarks)

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

		if has_url {
			bookmarks = append(bookmarks, bookmark)
		}

	}

	return bookmarks
}

func getProfilePath(profile string) string {
	var path string
	cfg, err := ini.Load(ffDir + "/installs.ini")

	if err != nil {
		log.Error("Fail to read file", "error", err)
		os.Exit(1)
	}

	for _, sec := range cfg.Sections() {
		if sec.HasKey("Default") {
			name, _ := sec.GetKey("Default")
			path = fmt.Sprintf("%s/%s", ffDir, name)
			break
		}
	}
	return path
}

func getMozBookmarks(profile_path string) ([]mozBookmark, error) {
	source, err := os.Open(profile_path + "/places.sqlite")

	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	defer source.Close()

	dst, err := os.CreateTemp("", "ff_places")
	if err != nil {
		log.Error("Error creating temp file", "error", err)
		os.Exit(1)
	}

	log.Debug("Tempporary DB", "name", dst.Name())

	defer os.Remove(dst.Name())

	nBytes, _ := io.Copy(dst, source)

	log.Debug("Bytes copied", "count", nBytes)

	db, err := sql.Open("sqlite3", dst.Name())

	if err != nil {
		log.Error("Error", "error", err)
		os.Exit(1)
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
