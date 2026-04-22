package services

import (
	"context"
	"time"

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
	Create(ctx context.Context, oportunidade *models.OportunidadeMEI) (int, error)
	GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error)
	Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filters map[string]interface{}, titulo string, limit, offset int) ([]*models.OportunidadeMEI, int, error)
	ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, limit, offset int) ([]*models.OportunidadeMEI, int, error)
	ListByOrgao(ctx context.Context, orgaoID string, limit, offset int) ([]*models.OportunidadeMEI, int, error)
	UpdateExpiredOpportunities(ctx context.Context) error
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

// CursoRepositoryInterface defines the interface for Curso repository
type CursoRepositoryInterface interface {
	Create(ctx context.Context, curso *models.Curso) (int, error)
	GetByID(ctx context.Context, id int) (*models.Curso, error)
	Update(ctx context.Context, curso *models.Curso) error
	UpdateStatus(ctx context.Context, id int, status models.StatusCurso) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error)
	CreateCustomFields(ctx context.Context, fields []models.CustomField) error
	CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error
	CreateLocationClasses(ctx context.Context, locations []models.LocationClass) error
	ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error)
	CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error)
	CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error)
	GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error)
}

// InscricaoRepositoryInterface defines the interface for Inscricao repository
type InscricaoRepositoryInterface interface {
	Create(ctx context.Context, inscricao *models.Inscricao) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error)
	Update(ctx context.Context, inscricao *models.Inscricao) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByCPFAndCurso(ctx context.Context, cpf string, cursoID int) (bool, error)
	GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error)
	UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error
	UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error)
	GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error)
	ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error)
	UpdateCertificate(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error
}

// InscricaoServiceInterface defines the interface for Inscricao service
type InscricaoServiceInterface interface {
	Create(ctx context.Context, inscricao *models.Inscricao) error
	CreateManual(ctx context.Context, inscricao *models.Inscricao) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error)
	GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error)
	UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error
	UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error)
	GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error)
	UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error
	UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error
	EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao)
	EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao)
	ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error)
}

// CursoServiceInterface defines the interface for Curso service
type CursoServiceInterface interface {
	Create(ctx context.Context, curso *models.Curso) (int, error)
	GetByID(ctx context.Context, id int) (*models.Curso, error)
	Update(ctx context.Context, curso *models.Curso) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Curso, int, error)
	SendToReview(ctx context.Context, id int) error
	Approve(ctx context.Context, id int) error
	RequestChanges(ctx context.Context, id int) error
	RequestDeletion(ctx context.Context, id int) error
}

// JobServiceInterface defines the interface for Job service
type JobServiceInterface interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// OportunidadeMEIServiceInterface defines the interface for OportunidadeMEI service
type OportunidadeMEIServiceInterface interface {
	Create(ctx context.Context, oportunidade *models.OportunidadeMEI, isDraft bool) (int, error)
	GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error)
	Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error
	Publish(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filters map[string]interface{}, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error)
	ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, page, pageSize int) ([]*models.OportunidadeMEI, int, error)
	ListDrafts(ctx context.Context, orgaoID, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error)
	ListActive(ctx context.Context, page, pageSize int) ([]*models.OportunidadeMEI, int, error)
	ListByOrgao(ctx context.Context, orgaoID string, page, pageSize int) ([]*models.OportunidadeMEI, int, error)
	UpdateExpiredOpportunities(ctx context.Context) error
}

// CategoriaServiceInterface defines the interface for Categoria service
type CategoriaServiceInterface interface {
	Create(ctx context.Context, categoria *models.Categoria) (int, error)
	GetByID(ctx context.Context, id int) (*models.Categoria, error)
	Update(ctx context.Context, categoria *models.Categoria) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Categoria, int, error)
}

// AcessibilidadeServiceInterface defines the interface for Acessibilidade service
type AcessibilidadeServiceInterface interface {
	Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Acessibilidade, error)
	Update(ctx context.Context, acessibilidade *models.Acessibilidade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Acessibilidade, int, error)
}

// EscolaridadeServiceInterface defines the interface for Escolaridade service
type EscolaridadeServiceInterface interface {
	Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Escolaridade, error)
	Update(ctx context.Context, escolaridade *models.Escolaridade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Escolaridade, int, error)
}
