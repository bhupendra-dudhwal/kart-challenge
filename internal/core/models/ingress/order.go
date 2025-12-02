package ingress

import "time"

type Order struct {
	Id        string    `json:"id"`
	Total     float64   `json:"total"`
	Discounts float64   `json:"discounts,omitempty"`
	Products  []Product `json:"products,omitempty"`
	Items     []Item    `json:"items"`
	CreatedAt time.Time `json:"createdAt"`
}

type Item struct {
	ProductId string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	Products  []Product `json:"products"`
}

type OrderReq struct {
	Items      []Item `json:"items"`
	CouponCode string `json:"couponCode"`
}
