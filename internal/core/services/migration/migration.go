package migration

import (
	"context"
	"math/rand"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	migrationPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/ingress/migration"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type migrationService struct {
	config *models.Config
	logger ports.LoggerPorts
	client *gorm.DB
}

func NewMigrationService(config *models.Config, logger ports.LoggerPorts, client *gorm.DB) migrationPorts.MigrationPorts {
	return &migrationService{
		config: config,
		logger: logger,
		client: client,
	}
}

func (m *migrationService) Migrate() {
	m.client.AutoMigrate(
		&ingressModels.Product{},
		&ingressModels.Order{},
		&ingressModels.Item{},
	)
}

func (m *migrationService) Seed() {
	m.seedProducts()
}

func (m *migrationService) seedProducts() {
	var count int64
	if err := m.client.Model(&ingressModels.Product{}).Count(&count).Error; err != nil {
		m.logger.Error("count check failed", zap.Error(err))
		return
	}
	if count > 0 {
		m.logger.Info("products already seeded")
		return
	}

	rand.Seed(time.Now().UnixNano())
	products := []ingressModels.Product{
		{
			Name:     "Orange Juice",
			Price:    5.99,
			Category: "Beverages",
			Image: &ingressModels.ProductImage{
				Thumbnail: randomImage(),
				Mobile:    randomImage(),
				Tablet:    randomImage(),
				Desktop:   randomImage(),
			},
		},
		{
			Name:     "Chips",
			Price:    2.49,
			Category: "Snacks",
			Image: &ingressModels.ProductImage{
				Thumbnail: randomImage(),
				Mobile:    randomImage(),
				Tablet:    randomImage(),
				Desktop:   randomImage(),
			},
		},
		{
			Name:     "Banana",
			Price:    2.49,
			Category: "fruits",
			Image: &ingressModels.ProductImage{
				Thumbnail: randomImage(),
				Mobile:    randomImage(),
				Tablet:    randomImage(),
				Desktop:   randomImage(),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.client.WithContext(ctx).Create(&products).Error; err != nil {
		m.logger.Error("product seeding failed", zap.Error(err))
		return
	}
	m.logger.Info("product seeding completed")
}

func randomImage() string {
	images := []string{
		"https://picsum.photos/200/200?random=1",
		"https://picsum.photos/200/200?random=2",
		"https://picsum.photos/200/200?random=3",
		"https://picsum.photos/200/200?random=4",
		"https://picsum.photos/200/200?random=5",
	}
	return images[rand.Intn(len(images))]
}
