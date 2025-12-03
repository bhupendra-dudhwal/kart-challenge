package dto

import (
	"fmt"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type OrderReq struct {
	Items      []ItemReq `json:"items"`
	CouponCode string    `json:"couponCode"`
}

func (o *OrderReq) Sanitize(step constants.ProcesssStep) {
	o.CouponCode = utils.Sanitize(o.CouponCode)
}

func (or OrderReq) Validate(cfg *models.CouponValidator) error {
	return validation.ValidateStruct(&or,
		validation.Field(&or.Items, validation.Required, validation.Length(1, 0)),
		validation.Field(&or.CouponCode, validation.By(func(value interface{}) error {
			code := value.(string)
			if code != "" {
				if !utils.ValidateCode(code, cfg) {
					return fmt.Errorf("invalid coupon code")
				}
			}
			return nil
		})),
	)
}
