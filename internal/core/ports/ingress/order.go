package ingress

import "github.com/valyala/fasthttp"

type OrderServicePorts interface {
	CreateOrder(ctx *fasthttp.RequestCtx)
}
