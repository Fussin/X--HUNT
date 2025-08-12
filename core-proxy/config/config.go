package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config is the root configuration structure for core-proxy.
type Config struct {
	Version         string            `mapstructure:"version"`
	Server          *ServerConfig      `mapstructure:"server" validate:"required"`
	Listeners       []ListenerConfig  `mapstructure:"listeners" validate:"required,min=1,dive"`
	TLS             TLSConfig         `mapstructure:"tls" validate:"required"`
	Multiplexer     MultiplexerConfig `mapstructure:"multiplexer"`
	ConnectionPool  ConnPoolConfig    `mapstructure:"connection_pool"`
	RateLimiter     RateLimiterConfig `mapstructure:"rate_limiter"`
	Backpressure    BackpressureCfg   `mapstructure:"backpressure"`
	Graceful        GracefulCfg       `mapstructure:"graceful_shutdown"`
	Logging         LoggingConfig     `mapstructure:"logging"`
	Metrics         MetricsConfig     `mapstructure:"metrics"`
	Tracing         TracingConfig     `mapstructure:"tracing"`
	AdminAPI        AdminAPIConfig    `mapstructure:"admin_api"`
	HotReload       HotReloadConfig   `mapstructure:"hot_reload"`
	Security        SecurityConfig    `mapstructure:"security"`
	Additional      map[string]any    `mapstructure:",remain"` // capture extra fields
}

// ServerConfig global server options
type ServerConfig struct {
	PIDFile      string        `mapstructure:"pid_file"`
	RunAsUser    string        `mapstructure:"run_as_user"`
	MaxOpenFiles int           `mapstructure:"max_open_files" validate:"gte=0"`
	TCPKeepAlive Duration      `mapstructure:"tcp_keepalive"`
}

// ListenerConfig describes a single listener binding
type ListenerConfig struct {
	Name       string        `mapstructure:"name" validate:"required"`
	Bind       string        `mapstructure:"bind" validate:"required"` // host:port
	Protocol   string        `mapstructure:"protocol_hint"`            // http/https etc
	Backlog    int           `mapstructure:"backlog"`
	TCPNoDelay bool          `mapstructure:"tcp_nodelay"`
	TLS        ListenerTLS   `mapstructure:"tls"`
	ALPN       []string      `mapstructure:"alpn"`
	IdleTimeout Duration     `mapstructure:"idle_timeout"`
}

// ListenerTLS per-listener TLS settings
type ListenerTLS struct {
	Enabled bool   `mapstructure:"enabled"`
	Type    string `mapstructure:"type"` // "static" | "dynamic"
}

// TLSConfig top-level TLS management config
type TLSConfig struct {
	RootCA struct {
		CertPath                 string `mapstructure:"cert_path"`
		KeyPath                  string `mapstructure:"key_path"`
		PrivateKeyProtection     string `mapstructure:"private_key_protection"` // vault|filesystem
	} `mapstructure:"root_ca"`
	DynamicCertCache struct {
		Enabled       bool   `mapstructure:"enabled"`
		CacheDir      string `mapstructure:"cache_dir"`
		MaxItems      int    `mapstructure:"max_items"`
		EvictionPolicy string `mapstructure:"eviction_policy"`
	} `mapstructure:"dynamic_cert_cache"`
	MinVersion string `mapstructure:"min_version"`
	Ciphers    []string `mapstructure:"cipher_suites"`
}

// MultiplexerConfig to tune detection behavior
type MultiplexerConfig struct {
	ALPNPriority []string `mapstructure:"alpn_priority"`
	SniffBytes   int      `mapstructure:"sniff_bytes"`
	SniffTimeout Duration `mapstructure:"sniff_timeout"`
}

// ConnPoolConfig connection pool sizing
type ConnPoolConfig struct {
	MaxGlobal       int      `mapstructure:"max_global" validate:"gte=0"`
	PerListenerMax  int      `mapstructure:"per_listener_max" validate:"gte=0"`
	PerOriginPool   int      `mapstructure:"per_origin_pool_size" validate:"gte=0"`
	IdleTimeout     Duration `mapstructure:"idle_timeout"`
	ReadBuffer      int      `mapstructure:"read_buffer"`  // bytes
	WriteBuffer     int      `mapstructure:"write_buffer"` // bytes
}

// RateLimiterConfig token bucket params
type RateLimiterConfig struct {
	Default TokenBucket `mapstructure:"default"`
	PerIP   TokenBucket `mapstructure:"per_ip"`
}

// TokenBucket describes a simple token-bucket limiter
type TokenBucket struct {
	Type     string  `mapstructure:"type"`     // token_bucket
	FillRate float64 `mapstructure:"fill_rate"`// tokens/sec
	Capacity int     `mapstructure:"capacity"`
}

// BackpressureCfg queue sizes and watermarks
type BackpressureCfg struct {
	QueueSize         int `mapstructure:"queue_size"`
	HighWatermarkPct  int `mapstructure:"high_watermark_pct"`
	LowWatermarkPct   int `mapstructure:"low_watermark_pct"`
	Policy            string `mapstructure:"policy"` // reject|slow_down|queue
}

// GracefulCfg draining settings
type GracefulCfg struct {
	Enabled        bool     `mapstructure:"enabled"`
	DrainTimeout   Duration `mapstructure:"drain_timeout"`
	StopAcceptAfter Duration `mapstructure:"stop_accept_after"`
}

// LoggingConfig structured logging
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
	Rotate struct {
		Enabled    bool `mapstructure:"enabled"`
		MaxSizeMB  int  `mapstructure:"max_size_mb"`
		MaxBackups int  `mapstructure:"max_backups"`
		MaxAgeDays int  `mapstructure:"max_age_days"`
	} `mapstructure:"rotate"`
}

// MetricsConfig prometheus settings
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Bind      string `mapstructure:"bind"`
	Path      string `mapstructure:"path"`
	Namespace string `mapstructure:"namespace"`
}

// TracingConfig OpenTelemetry settings
type TracingConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
	Sampler      string `mapstructure:"sampler"`
}

// AdminAPIConfig internal admin endpoints
type AdminAPIConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Bind      string `mapstructure:"bind"`
	AuthToken string `mapstructure:"auth_token"`
}

// HotReloadConfig runtime config reload
type HotReloadConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	WatchPaths  []string `mapstructure:"watch_paths"`
	DebounceMS  int      `mapstructure:"debounce_ms"`
}

// SecurityConfig general limits and hardening
type SecurityConfig struct {
	MinTLSVersion string `mapstructure:"min_tls_version"`
	AllowedCiphers []string `mapstructure:"allowed_ciphers"`
	DropPrivileges bool `mapstructure:"drop_privileges"`
	Chroot         string `mapstructure:"chroot"`
}

// Duration is a thin wrapper so we can unmarshal strings like "60s"
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		d.Duration = 0
		return nil
	}
	t, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = t
	return nil
}

// MarshalText for completeness
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// ------------------- Loading & Validation -------------------

var validate = validator.New()

// SetDefaults applies sane defaults to a Config in-place
func (c *Config) SetDefaults() {
	// Server defaults
	if c.Server == nil {
		c.Server = &ServerConfig{}
	}
	if c.Server.MaxOpenFiles == 0 {
		c.Server.MaxOpenFiles = 65536
	}
	if c.Server.TCPKeepAlive.Duration == 0 {
		c.Server.TCPKeepAlive = Duration{Duration: 300 * time.Second}
	}

	// Listeners defaults
	for i := range c.Listeners {
		if c.Listeners[i].Backlog == 0 {
			c.Listeners[i].Backlog = 1024
		}
		if c.Listeners[i].IdleTimeout.Duration == 0 {
			c.Listeners[i].IdleTimeout = Duration{Duration: 60 * time.Second}
		}
	}

	// TLS defaults
	if c.TLS.DynamicCertCache.MaxItems == 0 {
		c.TLS.DynamicCertCache.MaxItems = 20000
	}
	if c.TLS.DynamicCertCache.EvictionPolicy == "" {
		c.TLS.DynamicCertCache.EvictionPolicy = "lru"
	}
	if c.TLS.MinVersion == "" {
		c.TLS.MinVersion = "1.2"
	}

	// Multiplexer defaults
	if c.Multiplexer.SniffBytes == 0 {
		c.Multiplexer.SniffBytes = 512
	}
	if c.Multiplexer.SniffTimeout.Duration == 0 {
		c.Multiplexer.SniffTimeout = Duration{Duration: 200 * time.Millisecond}
	}

	// Connection pool
	if c.ConnectionPool.MaxGlobal == 0 {
		c.ConnectionPool.MaxGlobal = 200000
	}
	if c.ConnectionPool.PerListenerMax == 0 {
		c.ConnectionPool.PerListenerMax = 100000
	}
	if c.ConnectionPool.PerOriginPool == 0 {
		c.ConnectionPool.PerOriginPool = 100
	}
	if c.ConnectionPool.IdleTimeout.Duration == 0 {
		c.ConnectionPool.IdleTimeout = Duration{Duration: 60 * time.Second}
	}
	if c.ConnectionPool.ReadBuffer == 0 {
		c.ConnectionPool.ReadBuffer = 32 * 1024
	}
	if c.ConnectionPool.WriteBuffer == 0 {
		c.ConnectionPool.WriteBuffer = 32 * 1024
	}

	// Backpressure
	if c.Backpressure.QueueSize == 0 {
		c.Backpressure.QueueSize = 10000
	}
	if c.Backpressure.HighWatermarkPct == 0 {
		c.Backpressure.HighWatermarkPct = 85
	}
	if c.Backpressure.LowWatermarkPct == 0 {
		c.Backpressure.LowWatermarkPct = 60
	}
	if c.Backpressure.Policy == "" {
		c.Backpressure.Policy = "slow_down"
	}

	// Graceful defaults
	if c.Graceful.DrainTimeout.Duration == 0 {
		c.Graceful.DrainTimeout = Duration{Duration: 90 * time.Second}
	}
	if c.Graceful.StopAcceptAfter.Duration == 0 {
		c.Graceful.StopAcceptAfter = Duration{Duration: 10 * time.Second}
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}
}

// Validate runs rules and returns an error if config invalid
func (c *Config) Validate() error {
	// basic validator for struct tags
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// custom checks
	if len(c.Listeners) == 0 {
		return errors.New("at least one listener must be configured")
	}
	for _, l := range c.Listeners {
		if strings.TrimSpace(l.Bind) == "" {
			return fmt.Errorf("listener %s has empty bind", l.Name)
		}
	}
	// logical checks
	if c.ConnectionPool.MaxGlobal < c.ConnectionPool.PerListenerMax {
		return errors.New("connection_pool.max_global must be >= per_listener_max")
	}
	if c.Backpressure.HighWatermarkPct <= c.Backpressure.LowWatermarkPct {
		return errors.New("backpressure.high_watermark_pct must be > low_watermark_pct")
	}
	if c.HotReload.DebounceMS != 0 && c.HotReload.DebounceMS < 200 {
		return errors.New("hot_reload.debounce_ms must be >= 200")
	}
	return nil
}

// LoadConfig loads config from file path (yaml/json/toml). It supports env overrides.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Allow env overrides: COREPROXY_ prefix
	v.SetEnvPrefix("COREPROXY")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if path == "" {
		// search default locations
		v.SetConfigName("config")
		v.AddConfigPath("/etc/sentinelx/")
		v.AddConfigPath(".")
	} else {
		v.SetConfigFile(path)
	}

	// Provide default keys so viper knows types when env-only
	v.SetDefault("version", "1")
	v.SetDefault("logging.level", "info")

	if err := v.ReadInConfig(); err != nil {
		// if config file is not found, allow env-only mode
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	decoderConfig := &mapstructure.DecoderConfig{
		TagName: "mapstructure",
		Result:  &cfg,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			StringToDurationHookFunc(), // convert "60s" -> Duration
		),
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(v.AllSettings()); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	// Apply defaults and validate
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

// StringToDurationHookFunc converts strings like "60s" to Duration
func StringToDurationHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		ft reflect.Type, tt reflect.Type, data any,
	) (any, error) {
		// we only care about converting to Duration
		if tt != reflect.TypeOf(Duration{}) {
			return data, nil
		}
		switch s := data.(type) {
		case string:
			var d Duration
			if err := d.UnmarshalText([]byte(s)); err != nil {
				return nil, err
			}
			return d, nil
		default:
			return data, nil
		}
	}
}
