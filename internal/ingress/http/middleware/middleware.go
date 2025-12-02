package middleware

import (
	"kart-challenge/internal/constants"
	"kart-challenge/internal/core/models"
	"kart-challenge/internal/core/ports"
	ingressPorts "kart-challenge/internal/core/ports/ingress"
	"kart-challenge/internal/utils"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type middleware struct {
	config *models.Config
	logger ports.LoggerPorts
}

func NewMiddleware(config *models.Config, logger ports.LoggerPorts) ingressPorts.MiddlewarePorts {
	return &middleware{
		config: config,
		logger: logger,
	}
}

func (m *middleware) EnsureJSON(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetContentType(constants.JSON.String())
		next(ctx)
	}
}

func (m *middleware) Authorization(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if m.config.ApiKey.Enabled {
			apiKey, ok := utils.HeaderValue[string](ctx, constants.API_KEY.String())
			if !ok || apiKey == "" {
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				ctx.SetBodyString(`{"error": "API key missing"}`)
				return
			}

			if _, allowed := m.config.ApiKey.AllowedApiKeys[apiKey]; !allowed {
				ctx.SetStatusCode(fasthttp.StatusForbidden)
				ctx.SetBodyString(`{"error": "Invalid API key"}`)
				return
			}
		}
		next(ctx)
	}
}

func (m *middleware) RequestId(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		requestId := uuid.NewString()
		ctx.SetUserValue(constants.CtxRequestID, requestId)
		ctx.Response.Header.Set(constants.CtxRequestID.String(), requestId)

		next(ctx)
	}
}

func (m *middleware) PanicRecover(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic", zap.Any("err", err))

				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBodyString("internal error")
			}
		}()

		next(ctx)
	}
}
