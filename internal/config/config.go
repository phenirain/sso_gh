package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env              string     `mapstructure:"env"`
	ConnectionString string     `mapstructure:"connection_string"`
	AllowedOrigins   []string   `mapstructure:"allowed_origins"`
	Secret           string     `mapstructure:"secret"`
	APIAddress       string     `mapstructure:"api_address"`
	HTTP             HTTPConfig `mapstructure:"http"`
	GRPC             GRPCConfig `mapstructure:"grpc"`
}

type HTTPConfig struct {
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type GRPCConfig struct {
	Admin string `mapstructure:"admin"`
	Client string `mapstructure:"client"`
	Manager string `mapstructure:"manager"`
}

func LoadConfig() (*Config, error) {

	viper.SetConfigFile("./config/config.yaml")

	var cfg Config
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling config file: %w", err)
	}
	return &cfg, nil
}
