package ingress

type Product struct {
	Id       string        `json:"id"`
	Name     string        `json:"name"`
	Category string        `json:"category"`
	Price    float64       `json:"price"`
	Image    *ProductImage `json:"image,omitempty"`
}

type ProductImage struct {
	Thumbnail string `json:"thumbnail,omitempty"`
	Mobile    string `json:"mobile,omitempty"`
	Tablet    string `json:"tablet,omitempty"`
	Desktop   string `json:"desktop,omitempty"`
}

type PromoValidateReq struct {
	Code string `json:"code" validate:"required,min=8,max=10,alphanum"`
}
