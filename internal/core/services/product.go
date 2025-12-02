package services

import (
	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	ingressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/ingress"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type productService struct {
	config *models.Config
	logger ports.LoggerPorts
}

func NewProductService(config *models.Config, logger ports.LoggerPorts) ingressPorts.ProductServicePorts {
	return &productService{
		config: config,
		logger: logger,
	}
}

func (p *productService) ListProducts(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := p.logger.With(zap.Namespace("ListProducts"), zap.String(constants.CtxRequestID.String(), requestId))
	logger.Info("ListProducts")
}

func (p *productService) GetProduct(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := p.logger.With(zap.Namespace("GetProduct"), zap.String(constants.CtxRequestID.String(), requestId))
	logger.Info("GetProduct")
}
