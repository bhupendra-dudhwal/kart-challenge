package ingress

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Product struct {
	ID       int64         `json:"id" gorm:"primaryKey"`
	Name     string        `json:"name"`
	Price    float64       `json:"price"`
	Category string        `json:"category"`
	Image    *ProductImage `json:"image,omitempty" gorm:"type:jsonb"`
}

type ProductImage struct {
	Thumbnail string `json:"thumbnail,omitempty"`
	Mobile    string `json:"mobile,omitempty"`
	Tablet    string `json:"tablet,omitempty"`
	Desktop   string `json:"desktop,omitempty"`
}

func (pi *ProductImage) Value() (driver.Value, error) {
	if pi == nil {
		return nil, nil
	}
	return json.Marshal(pi)
}

func (pi *ProductImage) Scan(value interface{}) error {
	if value == nil {
		*pi = ProductImage{}
		return nil
	}

	var bytes []byte

	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to scan JSONB: unexpected type %T", value)
	}

	if err := json.Unmarshal(bytes, pi); err != nil {
		return fmt.Errorf("failed to unmarshal JSONB: %w", err)
	}

	return nil
}
