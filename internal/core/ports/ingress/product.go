package ingress

import "github.com/valyala/fasthttp"

type ProductServicePorts interface {
	ListProducts(ctx *fasthttp.RequestCtx)
	GetProduct(ctx *fasthttp.RequestCtx)
}
