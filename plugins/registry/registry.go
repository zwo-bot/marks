package registry

import (
	"fmt"
	"github.com/zwo-bot/marks/plugins/interfaces"
)

// PluginFactory is a function that creates a new plugin instance
type PluginFactory func(config interface{}) (interfaces.Plugin, error)

// registry holds all registered plugins
var registry = make(map[string]PluginFactory)

// Register registers a new plugin with the given name and factory function
func Register(name string, factory PluginFactory) {
	if _, exists := registry[name]; exists {
		return
	}
	registry[name] = factory
}

// Create creates a new plugin instance using the registered factory
func Create(name string, config interface{}) (interfaces.Plugin, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not registered", name)
	}
	return factory(config)
}

// ListPlugins returns a list of registered plugin names
func ListPlugins() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
