package plugins

import (
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
	"log/slog"
	"os"
)

type Plugin interface {
	GetName() string
	GetBookmarks() bookmark.Bookmarks
}

type Plugins []Plugin

var initers = []Plugin{}

var programLevel = new(slog.LevelVar)
var log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel}))

func Init() Plugins {
	programLevel.Set(slog.LevelDebug)

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
	var bookmarks bookmark.Bookmarks

	for _, plugin := range p {
		log.With("plugin", plugin.GetName())
		log.Info("Getting bookmarks")
		bookmarks = append(bookmarks, plugin.GetBookmarks()...)
	}
	return bookmarks
}
