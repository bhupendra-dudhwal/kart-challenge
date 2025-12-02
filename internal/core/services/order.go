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

type orderService struct {
	config *models.Config
	logger ports.LoggerPorts
}

func NewOrderService(config *models.Config, logger ports.LoggerPorts) ingressPorts.OrderServicePorts {
	return &orderService{
		config: config,
		logger: logger,
	}
}

func (o *orderService) CreateOrder(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := o.logger.With(zap.Namespace("CreateOrder"), zap.String(constants.CtxRequestID.String(), requestId))
	logger.Info("CreateOrder")
}

func (o *orderService) ListOrders(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := o.logger.With(zap.Namespace("ListOrders"), zap.String(constants.CtxRequestID.String(), requestId))
	logger.Info("ListOrders")
}

func (o *orderService) GetOrder(ctx *fasthttp.RequestCtx) {
	requestId, _ := utils.CtxValue[string](ctx, constants.CtxRequestID)
	logger := o.logger.With(zap.Namespace("GetOrder"), zap.String(constants.CtxRequestID.String(), requestId))
	logger.Info("GetOrder")
}
