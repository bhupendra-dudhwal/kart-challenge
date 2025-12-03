package egress

import (
	"gorm.io/gorm"
)

type DatabaseConnectionPorts interface {
	Connect() (*gorm.DB, error)
	Close() error
}
