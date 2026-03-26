// Package config loads and provides application configuration.
package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"formify/server/internal/logger"
)

type Config struct {
	Port                        string `mapstructure:"PORT"`
	Env                         string `mapstructure:"ENV"`
	DatabaseURL                 string `mapstructure:"DATABASE_URL"`
	JWTSecret                   string `mapstructure:"JWT_SECRET"`
	SessionSecret               string `mapstructure:"SESSION_SECRET"`
	GoogleClientID              string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret          string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleCallbackURL           string `mapstructure:"GOOGLE_CALLBACK_URL"`
	GoogleServiceAccountKeyPath string `mapstructure:"GOOGLE_SERVICE_ACCOUNT_KEY_PATH"`
	GoogleServiceAccountKey     string `mapstructure:"GOOGLE_SERVICE_ACCOUNT_KEY"`
	FrontendURL                 string `mapstructure:"FRONTEND_URL"`
	CORSOrigins                 string `mapstructure:"CORS_ORIGINS"`
	CloudinaryCloudName         string `mapstructure:"CLOUDINARY_CLOUD_NAME"`
	CloudinaryAPIKey            string `mapstructure:"CLOUDINARY_API_KEY"`
	CloudinaryAPISecret         string `mapstructure:"CLOUDINARY_API_SECRET"`
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) GetCORSOrigins() []string {
	if c.CORSOrigins != "" {
		return strings.Split(c.CORSOrigins, ",")
	}
	return []string{c.FrontendURL}
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "1323")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SESSION_SECRET", "formify-session-secret")
	viper.SetDefault("GOOGLE_CALLBACK_URL", "http://localhost:1323/api/auth/google/callback")
	viper.SetDefault("FRONTEND_URL", "http://localhost:5173")

	if err := viper.ReadInConfig(); err != nil {
		logger.GetLogger().Warn("No .env file found, using environment variables", zap.Error(err))
	}

	for _, key := range []string{
		"PORT", "ENV", "DATABASE_URL", "JWT_SECRET", "SESSION_SECRET",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_CALLBACK_URL",
		"GOOGLE_SERVICE_ACCOUNT_KEY_PATH", "GOOGLE_SERVICE_ACCOUNT_KEY",
		"FRONTEND_URL", "CORS_ORIGINS",
		"CLOUDINARY_CLOUD_NAME", "CLOUDINARY_API_KEY", "CLOUDINARY_API_SECRET",
	} {
		_ = viper.BindEnv(key, strings.ToUpper(key))
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		logger.GetLogger().Fatal("Failed to unmarshal config", zap.Error(err))
	}
	return cfg
}
