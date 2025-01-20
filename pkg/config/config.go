package config

import "golang.org/x/exp/slog"
import "github.com/spf13/viper"

// Config app configuration
type Config struct {
	ServerPort                  string `mapstructure:"PORT"`
	DSN                   string `mapstructure:"DSN"`
	TokenSecretKey        string `mapstructure:"JWT_TOKEN_SECRET_KEY"`
	TokenExpOn            int    `mapstructure:"JWT_TOKEN_EXP_ON"`
	TokenIssuer           string `mapstructure:"JWT_TOKEN_ISSUER"`
	RefreshTokenExpOn     int    `mapstructure:"JWT_REFRESH_TOKEN_EXP_ON"`
	RefreshTokenSecretKey string `mapstructure:"JWT_REFRESH_TOKEN_SECRET_KEY"`
	LogLevel							slog.Level    `mapstructure:"LOG_LEVEL"`
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
