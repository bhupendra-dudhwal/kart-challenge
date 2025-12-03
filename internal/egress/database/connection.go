package database

import (
	"fmt"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type database struct {
	client *gorm.DB
	config *models.Database
	logger ports.LoggerPorts
}

func NewDatabase(config *models.Database, logger ports.LoggerPorts) egressPorts.DatabaseConnectionPorts {
	return &database{
		config: config,
		logger: logger,
	}
}

func (d *database) Connect() (*gorm.DB, error) {
	dsn := d.buildDSN()

	// Configure logging mode
	gormLogger := logger.Default.LogMode(logger.Silent)
	if d.config.Debug {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	var (
		db  *gorm.DB
		err error
	)

	for attempt := 1; attempt <= d.config.ConnectRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})

		if err == nil {
			sqlDB, _ := db.DB()
			if pingErr := sqlDB.Ping(); pingErr == nil {
				d.logger.Info("Successfully connected to PostgreSQL",
					zap.String("host", d.config.Host),
					zap.Int("port", d.config.Port),
					zap.String("db", d.config.Name),
				)
				d.client = db
				break
			}
		}

		d.logger.Error("Database connection failed",
			zap.Int("attempt", attempt),
			zap.Int("maxAttempts", d.config.ConnectRetries),
			zap.Error(err),
		)

		if attempt < d.config.ConnectRetries {
			sleep := d.retrySleep(attempt)
			time.Sleep(sleep)
		}
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to database after %d attempts: %w",
			d.config.ConnectRetries, err,
		)
	}

	if err := d.configurePool(); err != nil {
		return nil, err
	}

	return d.client, nil
}

func (d *database) buildDSN() string {
	return fmt.Sprintf(
		// "host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		d.config.Host,
		d.config.Username,
		d.config.Password,
		d.config.Name,
		d.config.Port,
		d.config.Sslmode,
		// d.config.Timezone,
	)
}

func (d *database) retrySleep(attempt int) time.Duration {
	if d.config.RetryInterval > 0 {
		return d.config.RetryInterval
	}
	return time.Duration(attempt) * time.Second
}

func (d *database) configurePool() error {
	sqlDB, err := d.client.DB()
	if err != nil {
		return fmt.Errorf("failed to obtain sql.DB from gorm: %w", err)
	}

	if d.config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(d.config.MaxIdleConns)
	}
	if d.config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(d.config.MaxOpenConns)
	}
	if d.config.ConnMaxLife > 0 {
		sqlDB.SetConnMaxLifetime(d.config.ConnMaxLife)
	}
	if d.config.ConnMaxIdle > 0 {
		sqlDB.SetConnMaxIdleTime(d.config.ConnMaxIdle)
	}

	d.logger.Info("PostgreSQL connection pool configured",
		zap.Int("maxIdle", d.config.MaxIdleConns),
		zap.Int("maxOpen", d.config.MaxOpenConns),
		zap.Duration("connMaxLife", d.config.ConnMaxLife),
		zap.Duration("connMaxIdle", d.config.ConnMaxIdle),
	)

	return nil
}

func (d *database) Close() error {
	if d.client == nil {
		return nil
	}

	sqlDB, err := d.client.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
