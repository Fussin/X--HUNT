package pluginsdk

// Loader is responsible for loading and managing plugins.
type Loader interface {
	// Load loads a plugin from a given path.
	Load(path string) (Plugin, error)
	// Unload unloads a previously loaded plugin.
	Unload(plugin Plugin) error
}
