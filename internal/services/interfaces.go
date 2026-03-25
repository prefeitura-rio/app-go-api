package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	modelsEmp "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// Suppress unused import - modelsEmp is used below for OportunidadeMEI interface
var _ = modelsEmp.StatusVagaEmEdicao

// EmpregoRepositoryInterface defines the interface for Emprego repository.
type EmpregoRepositoryInterface interface {
	Create(ctx context.Context, emprego *models.Emprego) (int, error)
	GetByID(ctx context.Context, id int) (*models.Emprego, error)
	Update(ctx context.Context, emprego *models.Emprego) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Emprego, int, error)
}

// EmpresaRepositoryInterface defines the interface for main Empresa repository.
type EmpresaRepositoryInterface interface {
	Create(ctx context.Context, empresa *models.Empresa) (int, error)
	GetByID(ctx context.Context, id int) (*models.Empresa, error)
	Update(ctx context.Context, empresa *models.Empresa) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error)
}

// InstituicaoRepositoryInterface defines the interface for InstituicaoEnsino repository.
type InstituicaoRepositoryInterface interface {
	Create(ctx context.Context, instituicao *models.InstituicaoEnsino) (int, error)
	GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error)
	Update(ctx context.Context, instituicao *models.InstituicaoEnsino) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.InstituicaoEnsino, int, error)
}

// JobRepositoryInterface defines the interface for Job repository.
type JobRepositoryInterface interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

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

// CategoriaRepositoryInterface defines the interface for Categoria repository
type CategoriaRepositoryInterface interface {
	Create(ctx context.Context, categoria *models.Categoria) (int, error)
	GetByID(ctx context.Context, id int) (*models.Categoria, error)
	Update(ctx context.Context, categoria *models.Categoria) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Categoria, int, error)
}

// AcessibilidadeRepositoryInterface defines the interface for Acessibilidade repository
type AcessibilidadeRepositoryInterface interface {
	Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Acessibilidade, error)
	Update(ctx context.Context, acessibilidade *models.Acessibilidade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Acessibilidade, int, error)
}

// EscolaridadeRepositoryInterface defines the interface for Escolaridade repository
type EscolaridadeRepositoryInterface interface {
	Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Escolaridade, error)
	Update(ctx context.Context, escolaridade *models.Escolaridade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Escolaridade, int, error)
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
	UpdateProposta(ctx context.Context, id uuid.UUID, oportunidadeID int, valorProposta *float64, prazoExecucao *string, aceitaCustosIntegrais *bool) error
	UpdateStatusCidadao(ctx context.Context, id uuid.UUID, status models.StatusPropostaCidadao) error
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, page, pageSize int) ([]*models.PropostaMEI, int, error)
	ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, page, pageSize int) ([]*models.PropostaMEI, int, error)
	ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, page, pageSize int) ([]*models.PropostaMEI, int, error)
	UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error)
}
