package builder

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/services"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/egress/cache"
	cacheRepository "github.com/bhupendra-dudhwal/kart-challenge/internal/egress/cache/repository"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/egress/database"
	databaseRepository "github.com/bhupendra-dudhwal/kart-challenge/internal/egress/database/repository"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/ingress/http/handler"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/ingress/http/middleware"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"
	"github.com/bhupendra-dudhwal/kart-challenge/pkg/logger"
	"gorm.io/gorm"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	ingressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/ingress"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type appBuilder struct {
	ctx context.Context

	config *models.Config
	logger ports.LoggerPorts

	dbClient *gorm.DB

	handler fasthttp.RequestHandler
	server  *fasthttp.Server

	orderServicePorts   ingressPorts.OrderServicePorts
	productServicePorts ingressPorts.ProductServicePorts

	cacheRepository   egressPorts.CacheRepository
	orderRepository   egressPorts.OrderRepository
	productRepository egressPorts.ProductRepository
}

func NewAppBuilder(ctx context.Context) *appBuilder {
	return &appBuilder{
		ctx:    ctx,
		server: &fasthttp.Server{},
	}
}

func (a *appBuilder) LoadConfig(configFile string) error {
	start := time.Now()
	log.Println("Initializing config")
	// configFile = os.ExpandEnv(configFile)

	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute config path: %w", err)
	}

	if err := utils.FileExists(absPath); err != nil {
		return err
	}

	switch ext := filepath.Ext(absPath); ext {
	case ".yaml", ".yml":
	default:
		return fmt.Errorf("invalid file extension '%s', expected .yaml or .yml", ext)
	}

	configBytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg models.Config
	decoder := yaml.NewDecoder(bytes.NewReader(configBytes))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	a.config = &cfg

	log.Printf("Config loaded successfully from %s, duration(Micro Sec): %d\n", absPath, time.Since(start).Microseconds())
	return nil
}

func (a *appBuilder) SetLogger() (ports.LoggerPorts, error) {
	start := time.Now()
	log.Println("Initializing logger")
	loggerPorts, err := logger.NewLogger(a.config.Logger.Level, a.config.App.Environment)
	if err != nil {
		return nil, err
	}
	a.logger = loggerPorts
	loggerPorts.Info("Logger initialized successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))

	return loggerPorts, nil
}

func (a *appBuilder) SetRedisClientrepository() error {
	start := time.Now()
	a.logger.Info("Initializing SetRedisClient")

	redisClient, err := cache.NewCache(a.config.Cache, a.logger).Connect(a.ctx)
	if err != nil {
		return fmt.Errorf("cache connection err: %w", err)
	}

	a.cacheRepository = cacheRepository.NewRepository(redisClient)

	a.logger.Info("Redis initialized successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))
	return err
}

func (a *appBuilder) ProcessCouponData() error {
	start := time.Now()

	a.logger.Info("Initializing ProcessCouponData")
	if err := a.processCouponData(); err != nil {
		if !a.config.CouponConfig.IgnoreUnzipErrors {
			return err
		}
	}

	a.logger.Info("Coupon data initialized successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))

	return nil
}

func (a *appBuilder) SetDatabaseRepository() error {
	start := time.Now()

	a.logger.Info("Initializing SetDatabaseRepository")
	dbClient, err := database.NewDatabase(a.config.Database, a.logger).Connect()
	if err != nil {
		return err
	}
	a.dbClient = dbClient
	a.orderRepository = databaseRepository.NewOrderRepository(dbClient)
	a.productRepository = databaseRepository.NewProductRepository(dbClient)

	a.logger.Info("repository initialized successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))
	return nil
}

func (a *appBuilder) SetServices() error {
	start := time.Now()
	a.logger.Info("Initializing services", zap.String("component", "app_builder"), zap.String("step", "SetServices"))

	a.orderServicePorts = services.NewOrderService(a.config, a.logger, a.orderRepository, a.cacheRepository, a.productRepository)
	a.productServicePorts = services.NewProductService(a.config, a.logger, a.productRepository)

	a.logger.Info("Services initialized successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))
	return nil
}

func (a *appBuilder) SetHandlers() error {
	start := time.Now()
	a.logger.Info("Initializing handlers", zap.String("component", "app_builder"), zap.String("step", "SetHandlers"))

	middlewarePorts := middleware.NewMiddleware(a.config, a.logger)

	routes, handlerObj := handler.NewHandler(a.config, a.logger, middlewarePorts)
	handlerObj.SetProductHandler(a.productServicePorts)
	handlerObj.SetOrderHandler(a.orderServicePorts)

	a.handler = routes
	a.logger.Info("Handlers initialized successfully",
		zap.String("component", "app_builder"), zap.String("step", "SetHandlers"),
		zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()),
	)
	return nil
}

func (a *appBuilder) Build() (*fasthttp.Server, *models.App) {
	handler := a.handler
	if a.config.App.Server.Compression {
		handler = fasthttp.CompressHandlerLevel(
			a.handler,
			fasthttp.CompressBestSpeed,
		)
	}
	a.server.Handler = handler
	return a.server, a.config.App
}
