package ingress

type Item struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`

	ProductID int64 `json:"productId" gorm:"not null"`
	Quantity  int   `json:"quantity" gorm:"not null"`

	OrderID int64 `json:"orderId" gorm:"not null;index"`
	Order   Order `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
