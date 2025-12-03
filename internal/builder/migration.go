package builder

import (
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"gorm.io/gorm"
)

func (a *appBuilder) GetDbClient() *gorm.DB {
	return a.dbClient
}

func (a *appBuilder) GetConfig() *models.Config {
	return a.config
}
