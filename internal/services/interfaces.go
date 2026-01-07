package services

import (
	"context"

	"github.com/google/uuid"
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

// PropostaMEIRepositoryInterface defines the interface for PropostaMEI repository
type PropostaMEIRepositoryInterface interface {
	Create(ctx context.Context, proposta *models.PropostaMEI) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error)
	Update(ctx context.Context, proposta *models.PropostaMEI) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckExistingProposta(ctx context.Context, oportunidadeID int, meiEmpresaID string) (bool, error)
	ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, limit, offset int) ([]*models.PropostaMEI, int, error)
	ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, limit, offset int) ([]*models.PropostaMEI, int, error)
	ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, limit, offset int) ([]*models.PropostaMEI, int, error)
	UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error)
}

// OportunidadeMEIRepositoryInterface defines the interface for OportunidadeMEI repository
type OportunidadeMEIRepositoryInterface interface {
	GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error)
}

// CNAEValidationServiceInterface defines the interface for CNAE validation service
type CNAEValidationServiceInterface interface {
	ValidatePropostaForCNAE(ctx context.Context, authToken string, cnpj string, opportunityCNAEIDs []string) error
	CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error)
}

// PropostaMEIServiceInterface defines the interface for PropostaMEI service
type PropostaMEIServiceInterface interface {
	Create(ctx context.Context, proposta *models.PropostaMEI, authToken string) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error)
	Update(ctx context.Context, proposta *models.PropostaMEI) error
	UpdateProposta(ctx context.Context, id uuid.UUID, oportunidadeID int, valorProposta *float64) error
	UpdateStatusCidadao(ctx context.Context, id uuid.UUID, status models.StatusPropostaCidadao) error
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, page, pageSize int) ([]*models.PropostaMEI, int, error)
	ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, page, pageSize int) ([]*models.PropostaMEI, int, error)
	ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, page, pageSize int) ([]*models.PropostaMEI, int, error)
	UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error)
}
