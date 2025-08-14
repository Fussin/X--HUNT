package pluginsdk

// Plugin is the interface that all SentinelX plugins must implement.
type Plugin interface {
	// Name returns the name of the plugin.
	Name() string
	// Version returns the version of the plugin.
	Version() string
	// Init initializes the plugin with a given configuration.
	Init(config map[string]interface{}) error
	// Close cleans up any resources used by the plugin.
	Close() error
}
