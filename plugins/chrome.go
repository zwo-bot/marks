package plugins

import (
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/chrome"
)

func init() {
	config := &chrome.ChromeConfig{}
	plugin := &chrome.ChromePlugin{
		Config: config,
	}

	initers = append(initers, plugin)
}
