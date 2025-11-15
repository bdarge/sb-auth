package config

import (
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

// Config app configuration
type Config struct {
	ServerPort string `mapstructure:"PORT"`
	DSN string `mapstructure:"DSN"`
	PrivateKeyPath string `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	TokenExpOn int `mapstructure:"JWT_TOKEN_EXP_ON"`
	TokenIssuer string `mapstructure:"JWT_TOKEN_ISSUER"`
	RefreshTokenExpOn int `mapstructure:"JWT_REFRESH_TOKEN_EXP_ON"`
	LogLevel slog.Level `mapstructure:"LOG_LEVEL"`
}

// LoadConfig loads config from files
func LoadConfig(target string) (config Config, err error) {
	viper.AddConfigPath("./envs")
	viper.SetConfigName(target)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
