package ingress

import "time"

type Order struct {
	Id         int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Total      float64  `json:"total" gorm:"not null"`
	Discounts  *float64 `json:"discounts,omitempty"`
	CouponCode string   `json:"couponCode,omitempty"`

	Items     []Item    `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
