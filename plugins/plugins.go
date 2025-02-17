package plugins

import (
	"encoding/json"
	"log/slog"

	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/config"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/chrome"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/firefox"
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/interfaces"
)

type Plugins []interfaces.Plugin

var log *slog.Logger

// Init initializes all plugins
func Init() Plugins {
	log = logger.GetLogger()
	log.Info("Initializing plugins")
	plugins := Plugins{}

	// Initialize Firefox plugin
	ffConfig := &firefox.FirefoxConfig{}
	if pluginConfig, ok := config.GlobalConfig.Plugins["firefox"]; ok {
		jsonData, err := json.Marshal(pluginConfig)
		if err != nil {
			log.Error("Error marshaling Firefox config", "error", err)
		} else {
			if err := json.Unmarshal(jsonData, ffConfig); err != nil {
				log.Error("Error unmarshaling Firefox config", "error", err)
			} else {
				log.Debug("Loaded Firefox config", "profile_path", ffConfig.ProfilePath)
				// Only add Firefox plugin if we have a valid config
				plugins = append(plugins, &firefox.FirefoxPlugin{Config: ffConfig})
				log.Info("Added Firefox plugin")
			}
		}
	} else {
		log.Debug("No Firefox config found")
	}

	// Initialize Chrome plugin
	chromeConfig := &chrome.ChromeConfig{}
	if pluginConfig, ok := config.GlobalConfig.Plugins["chrome"]; ok {
		jsonData, err := json.Marshal(pluginConfig)
		if err != nil {
			log.Error("Error marshaling Chrome config", "error", err)
		} else {
			if err := json.Unmarshal(jsonData, chromeConfig); err != nil {
				log.Error("Error unmarshaling Chrome config", "error", err)
			} else {
				log.Debug("Loaded Chrome config", "profile_path", chromeConfig.ProfilePath)
				plugins = append(plugins, &chrome.ChromePlugin{Config: chromeConfig})
				log.Info("Added Chrome plugin")
			}
		}
	} else {
		// Add Chrome plugin with empty config to allow auto-detection
		log.Debug("No Chrome config found, using auto-detection")
		plugins = append(plugins, &chrome.ChromePlugin{Config: chromeConfig})
		log.Info("Added Chrome plugin with auto-detection")
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

func (p Plugins) GetBookmarsByPlugin(pluginName string) bookmark.Bookmarks {
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
