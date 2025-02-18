package chrome

import (
	"encoding/json"

	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins/interfaces"
	"github.com/zwo-bot/marks/plugins/registry"
)

func init() {
	registry.Register("chrome", createChromePlugin)
}

func createChromePlugin(config interface{}) (interfaces.Plugin, error) {
	log := logger.GetLogger()
	chromeConfig := &ChromeConfig{}

	// If config is provided, unmarshal it
	if config != nil {
		jsonData, err := json.Marshal(config)
		if err != nil {
			log.Error("Error marshaling Chrome config", "error", err)
			return nil, err
		}

		if err := json.Unmarshal(jsonData, chromeConfig); err != nil {
			log.Error("Error unmarshaling Chrome config", "error", err)
			return nil, err
		}

		log.Debug("Loaded Chrome config", "profile_path", chromeConfig.ProfilePath)
	} else {
		log.Debug("No Chrome config found, using auto-detection")
	}

	return &ChromePlugin{Config: chromeConfig}, nil
}
