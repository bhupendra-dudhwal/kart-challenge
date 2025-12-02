package handler

import (
	"kart-challenge/internal/core/models"
	"kart-challenge/internal/core/ports"
	ingressPorts "kart-challenge/internal/core/ports/ingress"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type middlewareFuc func(fasthttp.RequestHandler) fasthttp.RequestHandler

type handler struct {
	config *models.Config
	logger ports.LoggerPorts
	route  *router.Router
}

func chainMiddleware(base fasthttp.RequestHandler, middlewares ...middlewareFuc) fasthttp.RequestHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		base = middlewares[i](base)
	}

	return base
}

func NewHandler(config *models.Config, logger ports.LoggerPorts, middlewarePorts ingressPorts.MiddlewarePorts) (fasthttp.RequestHandler, ingressPorts.HandlerPorts) {
	r := router.New()

	commonMiddlewareHandlers := chainMiddleware(
		r.Handler,
		middlewarePorts.RequestId,
		middlewarePorts.PanicRecover,
		middlewarePorts.EnsureJSON,
		middlewarePorts.Authorization,
	)

	return commonMiddlewareHandlers, &handler{
		config: config,
		logger: logger,
		route:  r,
	}
}

func (h *handler) SetProductHandler(productServicePorts ingressPorts.ProductServicePorts) {
	h.route.GET("/api/v1/products", productServicePorts.ListProducts)
	h.route.GET("/api/v1/products/{productId}", productServicePorts.GetProduct)
}

func (h *handler) SetOrderHandler(orderServicePorts ingressPorts.OrderServicePorts) {
	h.route.GET("/api/v1/orders", orderServicePorts.ListOrders)
	h.route.GET("/api/v1/orders/{orderId}", orderServicePorts.GetOrder)
	h.route.POST("/api/v1/orders", orderServicePorts.CreateOrder)
}
