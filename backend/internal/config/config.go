package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Monitor  MonitorConfig  `mapstructure:"monitor"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug | release
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	// TLS 配置（HTTPS）
	TLSEnable   bool   `mapstructure:"tls_enable"`
	TLSCertFile string `mapstructure:"tls_cert_file"`
	TLSKeyFile  string `mapstructure:"tls_key_file"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type JWTConfig struct {
	KeyDir          string `mapstructure:"key_dir"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`  // seconds
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"` // seconds
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type MonitorConfig struct {
	Interval int `mapstructure:"interval"` // seconds
}

var global *Config

// Load 加载配置，支持环境变量覆盖
func Load(path string) (*Config, error) {
	v := viper.New()

	// 默认值
	v.SetDefault("server.port", 8888)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 60)
	v.SetDefault("server.write_timeout", 60)
	v.SetDefault("server.tls_enable", false)
	v.SetDefault("server.tls_cert_file", "")
	v.SetDefault("server.tls_key_file", "")
	v.SetDefault("database.path", "./data/cloudpivot.db")
	v.SetDefault("jwt.key_dir", "./data/jwt")
	v.SetDefault("jwt.access_token_ttl", 7200)    // 2h
	v.SetDefault("jwt.refresh_token_ttl", 604800) // 7d
	v.SetDefault("log.level", "info")
	v.SetDefault("monitor.interval", 15)

	// 配置文件
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/opt/cloudpivot/agent")
	}

	// 环境变量覆盖: CLOUDPIVOT_SERVER_PORT 等
	v.SetEnvPrefix("CLOUDPIVOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// 配置文件不存在时使用默认值，不报错
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	global = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if global == nil {
		_, _ = Load("")
	}
	return global
}
