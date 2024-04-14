package firefox

import (
    "github.com/zwo-bot/go-rofi-bookmarks/internal/config"
)

type FirefoxConfig struct {
    ProfilePath string `json:"profilePath"`
}

func (fc *FirefoxConfig) Load() error {
    filename := "firefox_config.json" // Configuration filename
    return config.LoadConfig(fc, filename)
}

func (fc *FirefoxConfig) Save() error {
    filename := "firefox_config.json" // Configuration filename
    return config.SaveConfig(fc, filename)
}
