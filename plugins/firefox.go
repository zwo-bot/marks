package plugins

import (
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/firefox"

)


func init() {

	config := &firefox.FirefoxConfig{}
	plugin := &firefox.FirefoxPlugin{
		Config: config,
	}
	initers = append(initers, plugin)
}
