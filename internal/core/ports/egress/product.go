package egress

import (
	"context"

	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
)

type ProductRepository interface {
	ListProducts(ctx context.Context) ([]ingressModels.Product, error)
	ListProductsByIds(ctx context.Context, productIds []int64) ([]ingressModels.Product, error)
	GetProduct(ctx context.Context, id int64) (*ingressModels.Product, error)
}
