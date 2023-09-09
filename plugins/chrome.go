package plugins

import (
	"fmt"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
)

func init() {
	plugin := &chromePlugin{
		URL: "test",
	}

	initers = append(initers, plugin)
}

type chromePlugin struct {
	URL string
}

func (c *chromePlugin) GetName() string {
	return "Chrome"
}

func (c *chromePlugin) GetBookmarks() bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks
	log.Info("Getting bookmarks")
	bookmark := bookmark.Bookmark{
		Title: "My second Bookmark",
	}
	bookmarks = append(bookmarks, bookmark)

	fmt.Println("Chrome get bookmarks")
	return bookmarks
}
