package interfaces

import (
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
)


type PluginConfig interface {
	Load() error
	Save() error
}

type Plugin interface {
	GetName() string
	GetBookmarks() bookmark.Bookmarks
	GetConfig() PluginConfig
	SetConfig(PluginConfig)
}