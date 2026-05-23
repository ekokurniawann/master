package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	JWT      JWTConfig      `yaml:"jwt"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	HTTP     HTTPConfig     `yaml:"http"`
	Health   HealthConfig   `yaml:"health"`
	Mailer   MailerConfig   `yaml:"mailer"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env"`
	Debug   bool   `yaml:"debug"`
	BaseURL string `yaml:"base_url"`
}

type JWTConfig struct {
	Secret   string `yaml:"secret"`
	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes"`
}

type DatabaseConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type RedisConfig struct {
	URL          string `yaml:"url"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type HTTPConfig struct {
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	RateLimit        RateLimit     `yaml:"rate_limit"`
	CompressionLevel int           `yaml:"compression_level"`
	CORS             CORSConfig    `yaml:"cors"`
}

type RateLimit struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type HealthConfig struct {
	Path string `yaml:"path"`
}

type MailerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	SenderEmail string `yaml:"sender_email"`
	SenderName  string `yaml:"sender_name"`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	return cfg
}

func Load(configPath string) error {
	var err error

	once.Do(func() {
		fileBytes, errReadFile := os.ReadFile(configPath)
		if errReadFile != nil {
			err = fmt.Errorf("config.config.Load: %w", errReadFile)
			return
		}

		expandedContent := os.ExpandEnv(string(fileBytes))

		var target Config
		if errUnmarshal := yaml.Unmarshal([]byte(expandedContent), &target); errUnmarshal != nil {
			err = fmt.Errorf("config.config.Load: %w", errUnmarshal)
			return
		}

		cfg = target
	})

	return err
}
