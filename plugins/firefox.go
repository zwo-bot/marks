package plugins

import (
	"github.com/zwo-bot/go-rofi-bookmarks/plugins/firefox"

)


func init() {

	plugin := &firefox.FirefoxPlugin{
		URL: "test",
	}
	initers = append(initers, plugin)
}
