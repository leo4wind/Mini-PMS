package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Upload UploadConfig `mapstructure:"upload"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expireHours"`
}

type UploadConfig struct {
	Dir       string `mapstructure:"dir"`
	MaxSizeMB int64  `mapstructure:"maxSizeMB"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.JWT.ExpireHours == 0 {
		c.JWT.ExpireHours = 72
	}
	if c.Upload.Dir == "" {
		c.Upload.Dir = "./uploads"
	}
	if c.Upload.MaxSizeMB == 0 {
		c.Upload.MaxSizeMB = 20
	}
	return &c, nil
}
