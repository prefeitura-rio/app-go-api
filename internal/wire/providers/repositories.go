package providers

import (
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// ProvideCategoriaRepository creates CategoriaRepository
func ProvideCategoriaRepository(db *gorm.DB) repository.CategoriaRepositoryInterface {
	return repository.NewCategoriaRepository(db)
}

// ProvideCursoRepository creates CursoRepository
func ProvideCursoRepository(db *gorm.DB) *repository.CursoRepository {
	return repository.NewCursoRepository(db)
}

// ProvideInscricaoRepository creates InscricaoRepository
func ProvideInscricaoRepository(db *gorm.DB) *repository.InscricaoRepository {
	return repository.NewInscricaoRepository(db)
}

// ProvideEmpregoRepository creates EmpregoRepository
func ProvideEmpregoRepository(db *gorm.DB) *repository.EmpregoRepository {
	return repository.NewEmpregoRepository(db)
}

// ProvideAcessibilidadeRepository creates AcessibilidadeRepository
func ProvideAcessibilidadeRepository(db *gorm.DB) *repository.AcessibilidadeRepository {
	return repository.NewAcessibilidadeRepository(db)
}

// ProvideEscolaridadeRepository creates EscolaridadeRepository
func ProvideEscolaridadeRepository(db *gorm.DB) *repository.EscolaridadeRepository {
	return repository.NewEscolaridadeRepository(db)
}

// ProvideEmpresaRepository creates the core EmpresaRepository
func ProvideEmpresaRepository(db *gorm.DB) *repository.EmpresaRepository {
	return repository.NewEmpresaRepository(db)
}

// ProvideInstituicaoRepository creates InstituicaoRepository
func ProvideInstituicaoRepository(db *gorm.DB) *repository.InstituicaoRepository {
	return repository.NewInstituicaoRepository(db)
}

// ProvideJobRepository creates JobRepository
func ProvideJobRepository(db *gorm.DB) *repository.JobRepository {
	return repository.NewJobRepository(db)
}

// ProvideOportunidadeMEIRepository creates OportunidadeMEIRepository
func ProvideOportunidadeMEIRepository(db *gorm.DB) *repository.OportunidadeMEIRepository {
	return repository.NewOportunidadeMEIRepository(db)
}

// ProvidePropostaMEIRepository creates PropostaMEIRepository
func ProvidePropostaMEIRepository(db *gorm.DB) *repository.PropostaMEIRepository {
	return repository.NewPropostaMEIRepository(db)
}

// ProvideOrgaoSnapshotRepository creates OrgaoSnapshotRepository
func ProvideOrgaoSnapshotRepository(db *gorm.DB) *repository.OrgaoSnapshotRepository {
	return repository.NewOrgaoSnapshotRepository(db)
}

// ProvideCitizenSnapshotRepository creates CitizenSnapshotRepository
func ProvideCitizenSnapshotRepository(db *gorm.DB) *repository.CitizenSnapshotRepository {
	return repository.NewCitizenSnapshotRepository(db)
}
