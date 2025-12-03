package dto

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ItemReq struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}

func (i ItemReq) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ProductID, validation.Required),
		validation.Field(&i.Quantity, validation.Required, validation.Min(1)),
	)
}
