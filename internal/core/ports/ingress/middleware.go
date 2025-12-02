package ingress

import "github.com/valyala/fasthttp"

type MiddlewarePorts interface {
	RequestId(next fasthttp.RequestHandler) fasthttp.RequestHandler
	Authorization(next fasthttp.RequestHandler) fasthttp.RequestHandler
	PanicRecover(next fasthttp.RequestHandler) fasthttp.RequestHandler
	EnsureJSON(next fasthttp.RequestHandler) fasthttp.RequestHandler
}
