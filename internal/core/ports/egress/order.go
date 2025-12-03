package egress

import (
	"context"

	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, payload *ingressModels.Order) error
}
