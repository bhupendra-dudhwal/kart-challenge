package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	App          *App          `yaml:"app"`
	Logger       *Logger       `yaml:"logger"`
	ApiKey       *ApiKey       `yaml:"apiKey"`
	CouponConfig *CouponConfig `yaml:"couponConfig"`
	Cache        *Cache        `yaml:"cache"`
	Database     *Database     `yaml:"database"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.App, validation.Required, validation.NotNil),
		validation.Field(&c.Logger, validation.Required, validation.NotNil),
		validation.Field(&c.CouponConfig, validation.Required, validation.NotNil),
		validation.Field(&c.ApiKey, validation.Required, validation.NotNil),
		validation.Field(&c.Cache, validation.Required, validation.NotNil),
	)
}

type App struct {
	Environment constants.Environment `yaml:"environment"`
	Server      *Server               `yaml:"server"`
}

func (a App) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Server, validation.Required, validation.NotNil),
		validation.Field(&a.Environment, validation.Required, validation.By(func(value interface{}) error {
			env, _ := value.(constants.Environment)
			if !env.IsValid() {
				return fmt.Errorf("invalid environment: %s", env)
			}
			return nil
		})),
	)
}

type Server struct {
	Compression          bool          `yaml:"compression"`
	GracefulShutdownTime time.Duration `yaml:"gracefulShutdownTime"`
	Port                 int           `yaml:"port"`
}

func (s Server) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Port, validation.Required, validation.By(func(value interface{}) error {
			port, _ := value.(int)
			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid port number: %d", port)
			}
			return nil
		})),
		validation.Field(&s.GracefulShutdownTime, validation.Required, validation.Min(5*time.Second), validation.Max(5*time.Minute)),
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

type CouponConfig struct {
	IgnoreUnzipErrors bool             `yaml:"ignoreUnzipErrors"`
	BloomKey          string           `yaml:"bloomKey"`
	ExactSet          string           `yaml:"exactSet"`
	BatchSize         int              `yaml:"batchSize"`
	Files             []string         `yaml:"files"`
	Validation        *CouponValidator `yaml:"validation"`
}

func (c CouponConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Files, validation.Required, validation.Length(1, 3)),
		validation.Field(&c.Validation, validation.Required, validation.NotNil),
		validation.Field(&c.BloomKey, validation.Required),
		validation.Field(&c.ExactSet, validation.Required),
		validation.Field(&c.BatchSize, validation.Required, validation.Min(100), validation.Max(10000)),
	)
}

type CouponValidator struct {
	MinLength         int                         `yaml:"minLength"`
	MaxLength         int                         `yaml:"maxLength"`
	AllowedCharacters constants.AllowedCharacters `yaml:"allowedCharacters"`
}

func (c CouponValidator) Validate() error {
	err := validation.ValidateStruct(&c,
		validation.Field(&c.MinLength, validation.Required, validation.Min(1)),
		validation.Field(&c.MaxLength, validation.Required, validation.Min(1)),
		validation.Field(&c.AllowedCharacters, validation.Required, validation.By(func(value interface{}) error {
			v, _ := value.(constants.AllowedCharacters)
			if !v.IsValid() {
				return fmt.Errorf("invalid allowedCharacters value: %s", v)
			}
			return nil
		})),
	)
	if err != nil {
		return err
	}

	if c.MinLength > c.MaxLength {
		return fmt.Errorf("minLength (%d) cannot be greater than maxLength (%d)", c.MinLength, c.MaxLength)
	}

	return nil
}

type Cache struct {
	Name           int           `yaml:"name"`
	Host           string        `yaml:"host"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
	Port           int           `yaml:"port"`
	PoolSize       int           `yaml:"poolSize"`
	MinIdleConns   int           `yaml:"minIdleConns"`
	DialTimeout    time.Duration `yaml:"dialTimeout"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
	ConnectRetries int           `yaml:"connectRetries"`
	RetryInterval  time.Duration `yaml:"retryInterval"`
	TTL            time.Duration `yaml:"ttl"`
}

func (c Cache) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.Min(1)),
		validation.Field(&c.Port, validation.Required, validation.Min(1024), validation.Max(65535)),
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.PoolSize, validation.Required, validation.Min(5)),
		validation.Field(&c.MinIdleConns, validation.Required, validation.Min(0), validation.By(func(value interface{}) error {
			minIdle := value.(int)
			if c.PoolSize > 0 && minIdle > c.PoolSize {
				return errors.New("minIdleConns cannot be greater than poolSize")
			}
			return nil
		})),
		validation.Field(&c.DialTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.ReadTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.WriteTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.ConnectRetries, validation.Required, validation.Min(1)),
		validation.Field(&c.RetryInterval, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.TTL, validation.Required, validation.Min(time.Millisecond)),
	)
}

type Database struct {
	Debug          bool          `yaml:"debug"`
	Name           string        `yaml:"name"`
	Host           string        `yaml:"host"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
	Port           int           `yaml:"port"`
	Timezone       string        `yaml:"timezone"`
	Sslmode        string        `yaml:"sslmode"`
	MaxIdleConns   int           `yaml:"maxIdleConns"`
	MaxOpenConns   int           `yaml:"maxOpenConns"`
	ConnMaxLife    time.Duration `yaml:"connMaxLife"`
	ConnMaxIdle    time.Duration `yaml:"connMaxIdle"`
	ConnectRetries int           `yaml:"connectRetries"`
	RetryInterval  time.Duration `yaml:"retryInterval"`
}

func (d Database) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Name, validation.Required),
		validation.Field(&d.Port, validation.Required, validation.Min(1111)),
		validation.Field(&d.Host, validation.Required),
		validation.Field(&d.Username, validation.Required),
		validation.Field(&d.Password, validation.Required),
		validation.Field(&d.Timezone, validation.Required),
		validation.Field(&d.Sslmode, validation.Required),
		validation.Field(&d.MaxIdleConns, validation.Required, validation.Min(10)),
		validation.Field(&d.MaxOpenConns, validation.Required, validation.Min(5)),
		validation.Field(&d.ConnMaxLife, validation.Required),
		validation.Field(&d.ConnMaxIdle, validation.Required),
		validation.Field(&d.ConnectRetries, validation.Required, validation.Min(3)),
		validation.Field(&d.RetryInterval, validation.Required),
	)
}
