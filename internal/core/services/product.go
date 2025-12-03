package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	ingressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/ingress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type productService struct {
	config            *models.Config
	logger            ports.LoggerPorts
	productRepository egressPorts.ProductRepository
}

func NewProductService(config *models.Config, logger ports.LoggerPorts, productRepository egressPorts.ProductRepository) ingressPorts.ProductServicePorts {
	return &productService{
		config:            config,
		logger:            logger,
		productRepository: productRepository,
	}
}

func (p *productService) ListProducts(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := p.logger.With(zap.Namespace("ListProducts"), zap.String(constants.CtxRequestID.String(), requestId))

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	products, err := p.productRepository.ListProducts(dbCtx)
	if err != nil {
		if errors.Is(err, utils.ErrNoData) {
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString(`[]`)
			return
		}
		logger.Error("failed to list products", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"internal server error"}`)
		return
	}

	response, err := json.Marshal(products)
	if err != nil {
		logger.Error("failed to marshal products response", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"internal server error"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(response)
}

func (p *productService) GetProduct(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := p.logger.With(zap.Namespace("GetProduct"), zap.String(constants.CtxRequestID.String(), requestId))

	productId, found := utils.PathParamValue[int64](ctx, "productId")
	if !found || productId <= 0 {
		logger.Error("invalid productId", zap.Any("productId", ctx.UserValue("productId")))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"productId must be a valid positive integer"}`)
		return
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	product, err := p.productRepository.GetProduct(dbCtx, productId)
	if err != nil {
		if errors.Is(err, utils.ErrNoData) {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetBodyString(`{"error":"product not found"}`)
			return
		}
		logger.Error("failed to get product", zap.Error(err), zap.Int64("productId", productId))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"internal server error"}`)
		return
	}

	respBody, err := json.Marshal(product)
	if err != nil {
		logger.Error("failed to marshal product response", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error":"internal server error"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(respBody)
}
