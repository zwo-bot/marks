package firefox

import (
	"encoding/json"

	"github.com/zwo-bot/marks/internal/logger"
	"github.com/zwo-bot/marks/plugins/interfaces"
	"github.com/zwo-bot/marks/plugins/registry"
)

func init() {
	registry.Register("firefox", createFirefoxPlugin)
}

func createFirefoxPlugin(config interface{}) (interfaces.Plugin, error) {
	log := logger.GetLogger()
	ffConfig := &FirefoxConfig{}

	// If config is provided, unmarshal it
	if config != nil {
		jsonData, err := json.Marshal(config)
		if err != nil {
			log.Error("Error marshaling Firefox config", "error", err)
			return nil, err
		}

		if err := json.Unmarshal(jsonData, ffConfig); err != nil {
			log.Error("Error unmarshaling Firefox config", "error", err)
			return nil, err
		}

		log.Debug("Loaded Firefox config", "profile_path", ffConfig.ProfilePath)
	} else {
		log.Debug("No Firefox config found")
	}

	return &FirefoxPlugin{Config: ffConfig}, nil
}
