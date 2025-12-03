package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	ingressModels "github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models/ingress/dto"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	ingressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/ingress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type orderService struct {
	config            *models.Config
	logger            ports.LoggerPorts
	orderRepository   egressPorts.OrderRepository
	cacheRepository   egressPorts.CacheRepository
	productRepository egressPorts.ProductRepository
}

func NewOrderService(config *models.Config, logger ports.LoggerPorts, orderRepository egressPorts.OrderRepository, cacheRepository egressPorts.CacheRepository, productRepository egressPorts.ProductRepository) ingressPorts.OrderServicePorts {
	return &orderService{
		config:            config,
		logger:            logger,
		orderRepository:   orderRepository,
		cacheRepository:   cacheRepository,
		productRepository: productRepository,
	}
}

func (o *orderService) CreateOrder(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := o.logger.With(zap.Namespace("CreateOrder"), zap.String(constants.CtxRequestID.String(), requestId))

	var payload dto.OrderReq
	if err := json.Unmarshal(ctx.PostBody(), &payload); err != nil {
		logger.Error("invalid request payload", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"invalid request payload"}`)
		return
	}

	payload.Sanitize(constants.ADD)
	if err := payload.Validate(o.config.CouponConfig.Validation); err != nil {
		logger.Error("validation failed", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return
	}

	var discountPercentage float32
	if payload.CouponCode != "" {
		ctxCache, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		exists, err := o.cacheRepository.BFExists(ctxCache, o.config.CouponConfig.BloomKey, payload.CouponCode)
		if err != nil {
			logger.Error("coupon validation failed", zap.Error(err))
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBodyString(`{"error":"failed to validate coupon"}`)
			return
		}
		if !exists {
			logger.Warn("coupon does not exist", zap.String("coupon", payload.CouponCode))
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetBodyString(`{"error":"coupon code does not exist"}`)
			return
		}

		// flat 20% discount if coupon exists
		discountPercentage = 20
	}

	productIds := make([]int64, 0, len(payload.Items))
	for _, item := range payload.Items {
		productIds = append(productIds, item.ProductID)
	}

	pctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	products, err := o.productRepository.ListProductsByIds(pctx, productIds)
	if err != nil {
		logger.Error("failed to fetch products", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"failed to fetch products"}`)
		return
	}

	if len(products) != len(productIds) {
		logger.Error("some products not found in DB")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"some products not found"}`)
		return
	}

	orderPayload, err := o.buildOrderFromRequest(discountPercentage, products, &payload)
	if err != nil {
		logger.Error("failed to build order payload", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return
	}

	dbCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	if err := o.orderRepository.CreateOrder(dbCtx, orderPayload); err != nil {
		switch {
		case errors.Is(err, utils.ErrDuplicateKey):
			logger.Warn("order already exists")
			ctx.SetStatusCode(fasthttp.StatusConflict)
			ctx.SetBodyString(`{"error":"order already exists"}`)
		default:
			logger.Error("failed to create order", zap.Error(err))
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBodyString(`{"error":"internal server error"}`)
		}
		return
	}

	responseBody, _ := json.Marshal(map[string]any{
		"message": "order created successfully",
		"orderId": orderPayload.Id,
	})
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody(responseBody)
}

func (o *orderService) buildOrderFromRequest(discountPercentage float32, products []ingressModels.Product, orderReq *dto.OrderReq) (*ingressModels.Order, error) {
	if len(products) != len(orderReq.Items) {
		return nil, fmt.Errorf("number of products does not match order items")
	}

	productMap := make(map[int64]ingressModels.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}

	order := &ingressModels.Order{
		CouponCode: orderReq.CouponCode,
		Items:      []ingressModels.Item{},
	}

	var totalPrice float64
	for _, item := range orderReq.Items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return nil, fmt.Errorf("product with ID %d not found", item.ProductID)
		}

		totalPrice += product.Price * float64(item.Quantity)
		order.Items = append(order.Items, ingressModels.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order.Total = totalPrice

	if discountPercentage > 0 {
		discountAmount := utils.RoundFloat64((totalPrice * float64(discountPercentage) / 100), 2)
		order.Discounts = &discountAmount
		order.Total -= discountAmount
	}

	order.Total = utils.RoundFloat64(order.Total, 2)

	return order, nil
}
