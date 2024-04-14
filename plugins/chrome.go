package plugins

import (
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/chrome"
)

func init() {
	plugin := &chrome.ChromePlugin{
		URL: "test",
	}

	initers = append(initers, plugin)
}