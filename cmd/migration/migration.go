package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/builder"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/services/migration"
	"go.uber.org/zap"
)

func main() {
	var envPath string
	flag.StringVar(&envPath, "env", "config/config.yaml", "Path to environment config file")
	flag.Parse()

	ctx := context.Background()
	appBuilder := builder.NewAppBuilder(ctx)

	if err := appBuilder.LoadConfig(envPath); err != nil {
		log.Fatalf("failed to load config (%s): %+v", envPath, err)
	}

	logger, err := appBuilder.SetLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %+v", err)
	}
	defer logger.Close()

	if err := appBuilder.SetDatabaseRepository(); err != nil {
		logger.Error("failed to initialize database", zap.Error(err))
		os.Exit(1)
	}

	dbClient := appBuilder.GetDbClient()
	config := appBuilder.GetConfig()

	migrationPorts := migration.NewMigrationService(config, logger, dbClient)
	migrationPorts.Migrate()
	migrationPorts.Seed()
}
