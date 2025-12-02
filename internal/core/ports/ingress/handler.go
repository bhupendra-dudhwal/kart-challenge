package ingress

type HandlerPorts interface {
	SetProductHandler(productServicePorts ProductServicePorts)
	SetOrderHandler(orderServicePorts OrderServicePorts)
}
