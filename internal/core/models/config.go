package models

import (
	"fmt"
	"kart-challenge/internal/constants"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	App    *App    `yaml:"app"`
	Logger *Logger `yaml:"logger"`
	ApiKey *ApiKey `yaml:"apiKey"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.App, validation.Required, validation.NotNil),
		validation.Field(&c.Logger, validation.Required, validation.NotNil),
		// validation.Field(&c.AllowedApiKeys, validation.Required, validation.NotNil),
		validation.Field(&c.ApiKey, validation.Required, validation.NotNil),
	)
}

type App struct {
	Environment          constants.Environment `yaml:"environment"`
	Port                 int                   `yaml:"port"`
	GracefulShutdownTime time.Duration         `yaml:"gracefulShutdownTime"`
}

func (a App) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Environment, validation.Required, validation.By(func(value interface{}) error {
			env, _ := value.(constants.Environment)
			if !env.IsValid() {
				return fmt.Errorf("invalid environment: %s", env)
			}
			return nil
		})),
		validation.Field(&a.Port, validation.Required, validation.By(func(value interface{}) error {
			port, _ := value.(int)
			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid port number: %d", port)
			}
			return nil
		})),
		validation.Field(&a.GracefulShutdownTime, validation.Required, validation.Min(5*time.Second), validation.Max(5*time.Minute)),
	)
}

type Logger struct {
	Level constants.LogLevel `yaml:"level"`
}

func (l Logger) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.Level, validation.Required, validation.By(func(value interface{}) error {
			env, _ := value.(constants.LogLevel)
			if !env.IsValid() {
				return fmt.Errorf("invalid log level: %s", env)
			}
			return nil
		})),
	)
}

type ApiKey struct {
	Enabled        bool            `yaml:"enabled"`
	AllowedApiKeys map[string]bool `yaml:"allowedApiKeys"`
}

func (l ApiKey) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.AllowedApiKeys, validation.When(l.Enabled, validation.Required, validation.NotNil)),
	)
}
