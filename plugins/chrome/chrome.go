package chrome

import (
	"fmt"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/interfaces"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
)

type ChromePlugin struct {
	Config *interfaces.PluginConfig
	URL string
}

func (c *ChromePlugin) GetName() string {
	return "Chrome"
}

func (c *ChromePlugin) GetConfig() interfaces.PluginConfig {
	return *c.Config
}

func (c *ChromePlugin) SetConfig(cc interfaces.PluginConfig) {
	c.Config = &cc
}

func (c *ChromePlugin) GetBookmarks() bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks
	log := logger.GetLogger()
	log.Info("Getting bookmarks")
	bookmark := bookmark.Bookmark{
		Title: "My second Bookmark",
	}
	bookmarks = append(bookmarks, bookmark)

	fmt.Println("Chrome get bookmarks")
	return bookmarks
}
