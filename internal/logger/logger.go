package logger

import (
	"errors"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger        //nolint:gochecknoglobals // Package-level logger singleton used across the app.
	Sugar  *zap.SugaredLogger //nolint:gochecknoglobals // Package-level sugared logger singleton used across the app.
)

const keyValuePairSize = 2
const timestampKey = "timestamp"

func Init(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = timestampKey
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.TimeKey = timestampKey
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.Encoding = "console"
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	Sugar = Logger.Sugar()

	return nil
}

func Close() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

func GetLogger() *zap.Logger {
	if Logger == nil {
		logger, _ := zap.NewDevelopment()
		return logger
	}
	return Logger
}

func GetSugaredLogger() *zap.SugaredLogger {
	if Sugar == nil {
		logger := GetLogger()
		return logger.Sugar()
	}
	return Sugar
}

func InitFromEnv() error {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}
	return Init(env)
}

func ToField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func ToFields(values ...interface{}) []zap.Field {
	if len(values)%keyValuePairSize != 0 {
		return []zap.Field{zap.Error(errors.New("odd number of arguments"))}
	}

	fields := make([]zap.Field, 0, len(values)/keyValuePairSize)
	for i := 0; i < len(values); i += keyValuePairSize {
		key := values[i].(string)
		fields = append(fields, zap.Any(key, values[i+1]))
	}
	return fields
}
