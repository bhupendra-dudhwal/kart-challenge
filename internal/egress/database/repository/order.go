package repository

import (
	"context"
	"errors"

	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type orderRepository struct {
	client *gorm.DB
}

func NewOrderRepository(client *gorm.DB) egressPorts.OrderRepository {
	return &orderRepository{
		client: client,
	}
}

func (m *orderRepository) CreateOrder(ctx context.Context, payload *ingressModels.Order) error {
	if err := m.client.WithContext(ctx).Create(payload).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return utils.ErrDuplicateKey
		}
		return err
	}
	return nil
}
