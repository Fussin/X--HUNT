package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTestConfig returns a valid Config for reuse in tests
func createTestConfig() *Config {
	cfg := &Config{
		Server: &ServerConfig{},
	}
	cfg.SetDefaults()
	cfg.Listeners = []ListenerConfig{
		{Name: "test-listener", Bind: "127.0.0.1:8080"},
	}
	return cfg
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := createTestConfig()
	err := cfg.Validate()
	assert.NoError(t, err, "Expected no validation errors for a valid config")
}

func TestValidate_MissingListeners(t *testing.T) {
	cfg := createTestConfig()
	cfg.Listeners = []ListenerConfig{} // No listeners
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Listeners")
}

func TestValidate_ListenerMissingName(t *testing.T) {
	cfg := createTestConfig()
	cfg.Listeners[0].Name = "" // Missing name
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestValidate_ListenerMissingBind(t *testing.T) {
	cfg := createTestConfig()
	cfg.Listeners[0].Bind = "" // Missing bind
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bind")
}

func TestValidate_InvalidConnPool(t *testing.T) {
	cfg := createTestConfig()
	cfg.ConnectionPool.MaxGlobal = 100
	cfg.ConnectionPool.PerListenerMax = 200 // Invalid
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection_pool.max_global must be >= per_listener_max")
}

func TestValidate_InvalidBackpressure(t *testing.T) {
	cfg := createTestConfig()
	cfg.Backpressure.HighWatermarkPct = 60
	cfg.Backpressure.LowWatermarkPct = 70 // Invalid
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backpressure.high_watermark_pct must be > low_watermark_pct")
}

func TestValidate_InvalidHotReload(t *testing.T) {
	cfg := createTestConfig()
	cfg.HotReload.DebounceMS = 100 // Invalid
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hot_reload.debounce_ms must be >= 200")
}
