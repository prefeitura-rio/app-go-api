package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// CursoRepositoryInterface defines the interface for CursoRepository
type CursoRepositoryInterface interface {
	Create(ctx context.Context, curso *models.Curso) (int, error)
	GetByID(ctx context.Context, id int) (*models.Curso, error)
	Update(ctx context.Context, curso *models.Curso) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error)
	CreateCustomFields(ctx context.Context, customFields []models.CustomField) error
	CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error
	CreateLocationClasses(ctx context.Context, locationClasses []models.LocationClass) error
	CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error)
	CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart, enrollmentEnd *time.Time, autoApprove bool, err error)
	GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error)
	GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error)
}

// InscricaoRepositoryInterface defines the interface for InscricaoRepository
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

// EmpregoRepositoryInterface defines the interface for EmpregoRepository
type EmpregoRepositoryInterface interface {
	Create(ctx context.Context, emprego *models.Emprego) (int, error)
	GetByID(ctx context.Context, id int) (*models.Emprego, error)
	Update(ctx context.Context, emprego *models.Emprego) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Emprego, int, error)
}

// AcessibilidadeRepositoryInterface defines the interface for AcessibilidadeRepository
type AcessibilidadeRepositoryInterface interface {
	Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Acessibilidade, error)
	Update(ctx context.Context, acessibilidade *models.Acessibilidade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Acessibilidade, int, error)
}

// CategoriaRepositoryInterface defines the interface for CategoriaRepository
type CategoriaRepositoryInterface interface {
	Create(ctx context.Context, categoria *models.Categoria) (int, error)
	GetByID(ctx context.Context, id int) (*models.Categoria, error)
	Update(ctx context.Context, categoria *models.Categoria) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Categoria, int, error)
}

// EscolaridadeRepositoryInterface defines the interface for EscolaridadeRepository
type EscolaridadeRepositoryInterface interface {
	Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error)
	GetByID(ctx context.Context, id int) (*models.Escolaridade, error)
	Update(ctx context.Context, escolaridade *models.Escolaridade) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Escolaridade, int, error)
}

// EmpresaRepositoryInterface defines the interface for EmpresaRepository
type EmpresaRepositoryInterface interface {
	Create(ctx context.Context, empresa *models.Empresa) (int, error)
	GetByID(ctx context.Context, id int) (*models.Empresa, error)
	Update(ctx context.Context, empresa *models.Empresa) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error)
}

// InstituicaoRepositoryInterface defines the interface for InstituicaoRepository
type InstituicaoRepositoryInterface interface {
	Create(ctx context.Context, instituicao *models.InstituicaoEnsino) (int, error)
	GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error)
	Update(ctx context.Context, instituicao *models.InstituicaoEnsino) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.InstituicaoEnsino, int, error)
}

// JobRepositoryInterface defines the interface for JobRepository
type JobRepositoryInterface interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// OportunidadeMEIRepositoryInterface defines the interface for OportunidadeMEIRepository
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

// PropostaMEIRepositoryInterface defines the interface for PropostaMEIRepository
type PropostaMEIRepositoryInterface interface {
	Create(ctx context.Context, proposta *models.PropostaMEI) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error)
	Update(ctx context.Context, proposta *models.PropostaMEI) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, limit, offset int) ([]*models.PropostaMEI, int, error)
	ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, limit, offset int) ([]*models.PropostaMEI, int, error)
	ListByStatus(ctx context.Context, statusCidadao models.StatusPropostaCidadao, limit, offset int) ([]*models.PropostaMEI, int, error)
	CheckExistingProposta(ctx context.Context, oportunidadeID int, meiEmpresaID string) (bool, error)
	UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error)
}

// OrgaoSnapshotRepositoryInterface defines the interface for OrgaoSnapshotRepository
type OrgaoSnapshotRepositoryInterface interface {
	Create(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	GetByOrgaoID(ctx context.Context, orgaoID string) (*models.OrgaoSnapshot, error)
	GetByOrgaoIDs(ctx context.Context, orgaoIDs []string) (map[string]*models.OrgaoSnapshot, error)
	Update(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	Upsert(ctx context.Context, snapshot *models.OrgaoSnapshot) error
	BatchUpsert(ctx context.Context, snapshots []*models.OrgaoSnapshot) error
	GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration) ([]*models.OrgaoSnapshot, error)
	GetFailedSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error)
	GetPendingSnapshots(ctx context.Context) ([]*models.OrgaoSnapshot, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
}

// CitizenSnapshotRepositoryInterface defines the interface for CitizenSnapshotRepository
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
