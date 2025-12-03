package repository

import (
	"context"
	"errors"

	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"
	"gorm.io/gorm"
)

type productRepository struct {
	client *gorm.DB
}

func NewProductRepository(client *gorm.DB) egressPorts.ProductRepository {
	return &productRepository{
		client: client,
	}
}

func (m *productRepository) ListProducts(ctx context.Context) ([]ingressModels.Product, error) {
	var products []ingressModels.Product
	if err := m.client.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (m *productRepository) GetProduct(ctx context.Context, id int64) (*ingressModels.Product, error) {
	var product ingressModels.Product
	if err := m.client.WithContext(ctx).First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNoData
		}
		return nil, err
	}
	return &product, nil
}

func (m *productRepository) ListProductsByIds(ctx context.Context, productIds []int64) ([]ingressModels.Product, error) {
	var products []ingressModels.Product
	if err := m.client.WithContext(ctx).Find(&products, productIds).Error; err != nil {
		return nil, err
	}

	return products, nil
}
