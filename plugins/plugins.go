package plugins

import (
	"log/slog"

	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"github.com/zwo-bot/go-rofi-bookmarks/internal/logger"
)

type Plugin interface {
	GetName() string
	GetBookmarks() bookmark.Bookmarks
}

type Plugins []Plugin

var initers = []Plugin{}
var log *slog.Logger

func Init() Plugins {
	log = logger.GetLogger()
	log.Info("Initializing plugins")
	plugins := Plugins{}
	for _, p := range initers {
		plugin := p
		if p != nil {
			plugins = append(plugins, plugin)
		}
	}
	return plugins
}

func (p Plugins) GetBookmarks() bookmark.Bookmarks {
	log := logger.GetLogger()
	var bookmarks bookmark.Bookmarks

	for _, plugin := range p {
		log.With("plugin", plugin.GetName())
		log.Info("Getting bookmarks")
		bookmarks = append(bookmarks, plugin.GetBookmarks()...)
	}
	return bookmarks
}

func (p Plugins) ListPlugins() []string {
	var plugins []string

	for _, plugin := range p {
		plugins = append(plugins, plugin.GetName())
	}
	return plugins
}