package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}
	Database struct {
		DSN             string        `mapstructure:"dsn"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	}
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.max_idle_conns", 10)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
