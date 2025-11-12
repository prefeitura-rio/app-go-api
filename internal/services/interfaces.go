package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// RMIClientInterface defines the interface for RMI API client
type RMIClientInterface interface {
	GetUserLegalEntities(ctx context.Context, authToken string, cpf string) ([]models.LegalEntity, error)
}

// LegalEntitiesCacheInterface defines the interface for legal entities cache
type LegalEntitiesCacheInterface interface {
	Get(ctx context.Context, cpf string) ([]models.LegalEntity, error)
	Set(ctx context.Context, cpf string, entities []models.LegalEntity) error
}
