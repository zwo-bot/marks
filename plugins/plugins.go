package plugins

import (
	"github.com/zwo-bot/marks/bookmark"
	"github.com/zwo-bot/marks/internal/config"
	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins/interfaces"
	reg "github.com/zwo-bot/marks/plugins/registry"
)

type Plugins []interfaces.Plugin

// Init initializes all plugins
func Init() Plugins {
	log := logger.GetLogger()
	log.Info("Initializing plugins")
	plugins := Plugins{}

	// Initialize registered plugins
	for _, name := range reg.ListPlugins() {
		log.Debug("Initializing plugin", "name", name)
		
		// Get plugin config if available
		var pluginConfig interface{}
		if cfg, ok := config.GlobalConfig.Plugins[name]; ok {
			pluginConfig = cfg
		}

		// Create plugin instance
		plugin, err := reg.Create(name, pluginConfig)
		if err != nil {
			log.Error("Failed to initialize plugin", "name", name, "error", err)
			continue
		}

		plugins = append(plugins, plugin)
		log.Info("Added plugin", "name", name)
	}

	return plugins
}

func (p Plugins) GetBookmarks() bookmark.Bookmarks {
	log := logger.GetLogger()
	var bookmarks bookmark.Bookmarks

	for _, plugin := range p {
		name := plugin.GetName()
		log.Info("Getting bookmarks from plugin", "plugin", name)
		pluginBookmarks := plugin.GetBookmarks()
		log.Debug("Got bookmarks from plugin", "plugin", name, "count", len(pluginBookmarks))
		bookmarks = append(bookmarks, pluginBookmarks...)
	}
	return bookmarks.RemoveDuplicates()
}

func (p Plugins) GetBookmarksByPlugin(pluginName string) bookmark.Bookmarks {
	log := logger.GetLogger()
	var bookmarks bookmark.Bookmarks

	for _, plugin := range p {
		if plugin.GetName() == pluginName {
			log.With("plugin", plugin.GetName())
			log.Info("Getting bookmarks")
			bookmarks = append(bookmarks, plugin.GetBookmarks()...)
		}
	}
	return bookmarks.RemoveDuplicates()
}

func (p Plugins) ListPlugins() []string {
	var plugins []string

	for _, plugin := range p {
		plugins = append(plugins, plugin.GetName())
	}
	return plugins
}
