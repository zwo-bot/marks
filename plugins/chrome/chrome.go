package chrome

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/interfaces"
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

	// Read bookmarks file
	data, err := os.ReadFile(chromeConfig.ProfilePath)
	if err != nil {
		// Chrome is not installed or using a different path, return empty bookmarks without error
		log.Debug("Chrome bookmarks file not found", "path", chromeConfig.ProfilePath)
		return bookmark.Bookmarks{}
	}

	var chromeBookmarks ChromeBookmarks
	if err := json.Unmarshal(data, &chromeBookmarks); err != nil {
		log.Error("Error parsing bookmarks file", "error", err)
		return bookmark.Bookmarks{}
	}

	// Process each root folder
	for folder, root := range chromeBookmarks.Roots {
		bookmarks = append(bookmarks, processBookmarks(root, folder)...)
	}

	return bookmarks
}

func processBookmarks(node ChromeBookmark, path string) bookmark.Bookmarks {
	var bookmarks bookmark.Bookmarks

	// If it's a URL bookmark, add it
	if node.Type == "url" {
		bookmarks = append(bookmarks, bookmark.Bookmark{
			Title:  node.Name,
			URI:    node.URL,
			Path:   path,
			Source: "Chrome",
		})
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
		bookmarks = append(bookmarks, processBookmarks(child, newPath)...)
	}

	return bookmarks
}
