package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// --- Repository Interfaces ---

type CursoRepositoryInterface interface {
	Create(ctx context.Context, curso *models.Curso) (int, error)
	GetByID(ctx context.Context, id int) (*models.Curso, error)
	Update(ctx context.Context, curso *models.Curso) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error)
	CreateCustomFields(ctx context.Context, customFields []models.CustomField) error
	CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error
	CreateLocationClasses(ctx context.Context, locationClasses []models.LocationClass) error
	ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart, enrollmentEnd *time.Time, autoApprove bool, err error)
	CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error)
	CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error)
	GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error)
}

type InscricaoRepositoryInterface interface {
	Create(ctx context.Context, inscricao *models.Inscricao) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error)
	GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error)
	UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error
	UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error)
	GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByCPFAndCurso(ctx context.Context, cpf string, cursoID int) (bool, error)
	ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error)
	UpdateCertificate(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error
	Update(ctx context.Context, inscricao *models.Inscricao) error
}

type JobRepositoryInterface interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type EmpregoRepositoryInterface interface {
	Create(ctx context.Context, emprego *models.Emprego) (int, error)
	GetByID(ctx context.Context, id int) (*models.Emprego, error)
	Update(ctx context.Context, emprego *models.Emprego) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Emprego, int, error)
}

type EmpresaRepositoryInterface interface {
	Create(ctx context.Context, empresa *models.Empresa) (int, error)
	GetByID(ctx context.Context, id int) (*models.Empresa, error)
	Update(ctx context.Context, empresa *models.Empresa) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error)
}

type InstituicaoRepositoryInterface interface {
	Create(ctx context.Context, instituicao *models.InstituicaoEnsino) (int, error)
	GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error)
	Update(ctx context.Context, instituicao *models.InstituicaoEnsino) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.InstituicaoEnsino, int, error)
}

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

type CategoriaRepositoryInterface interface {
	Create(ctx context.Context, categoria *models.Categoria) (int, error)
	GetByID(ctx context.Context, id int) (*models.Categoria, error)
	Update(ctx context.Context, categoria *models.Categoria) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Categoria, int, error)
}

type AcessibilidadeRepositoryInterface interface {
	Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Acessibilidade, error)
	Update(ctx context.Context, acessibilidade *models.Acessibilidade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Acessibilidade, int, error)
}

type EscolaridadeRepositoryInterface interface {
	Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Escolaridade, error)
	Update(ctx context.Context, escolaridade *models.Escolaridade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Escolaridade, int, error)
}

type CitizenSnapshotRepositoryInterface interface {
	GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
	GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error)
	Upsert(ctx context.Context, snapshot *models.CitizenSnapshot) error
	BatchUpsert(ctx context.Context, snapshots []*models.CitizenSnapshot) error
	GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error)
	GetCPFsWithEnrollments(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error)
	Delete(ctx context.Context, cpf string) error
	Count(ctx context.Context) (int64, error)
}

type OrgaoSnapshotRepositoryInterface interface {
	Create(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	GetByOrgaoID(ctx context.Context, orgaoID string) (*models.OrgaoSnapshot, error)
	Update(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	Upsert(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	BatchUpsert(ctx context.Context, snapshots []*models.OrgaoSnapshot) error
	GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration) ([]*models.OrgaoSnapshot, error)
	GetFailedSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error)
	GetPendingSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
}

// --- Service Interfaces ---

type RMIClientInterface interface {
	GetUserLegalEntities(ctx context.Context, authToken string, cpf string) ([]models.LegalEntity, error)
}

type LegalEntitiesCacheInterface interface {
	Get(ctx context.Context, cpf string) ([]models.LegalEntity, error)
	Set(ctx context.Context, cpf string, entities []models.LegalEntity) error
}

type CNAEValidationServiceInterface interface {
	ValidatePropostaForCNAE(ctx context.Context, authToken string, cnpj string, opportunityCNAEIDs []string) error
	CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error)
}

type CitizenDataFetcher interface {
	SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
}

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
