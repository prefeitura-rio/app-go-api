package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type CursoService struct {
	repo *repository.CursoRepository
}

func NewCursoService(repo *repository.CursoRepository) *CursoService {
	return &CursoService{
		repo: repo,
	}
}

func (s *CursoService) Create(ctx context.Context, curso *models.Curso) (int, error) {
	if err := s.validateCurso(curso); err != nil {
		return 0, fmt.Errorf("erro de validação: %w", err)
	}

	s.normalizeCurso(curso)

	// Criar o curso primeiro
	id, err := s.repo.Create(ctx, curso)
	if err != nil {
		return 0, err
	}

	// Associar o ID aos relacionamentos
	for i := range curso.CustomFields {
		curso.CustomFields[i].CursoID = id
	}
	if curso.RemoteClass != nil {
		curso.RemoteClass.CursoID = id
	}
	for i := range curso.LocationClasses {
		curso.LocationClasses[i].CursoID = id
	}

	// Salvar relacionamentos se existirem
	if len(curso.CustomFields) > 0 {
		if err := s.repo.CreateCustomFields(ctx, curso.CustomFields); err != nil {
			return 0, fmt.Errorf("erro ao criar custom fields: %w", err)
		}
	}

	if curso.RemoteClass != nil {
		if err := s.repo.CreateRemoteClass(ctx, curso.RemoteClass); err != nil {
			return 0, fmt.Errorf("erro ao criar remote class: %w", err)
		}
	}

	if len(curso.LocationClasses) > 0 {
		if err := s.repo.CreateLocationClasses(ctx, curso.LocationClasses); err != nil {
			return 0, fmt.Errorf("erro ao criar location classes: %w", err)
		}
	}

	return id, nil
}

func (s *CursoService) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CursoService) Update(ctx context.Context, curso *models.Curso) error {
	if err := s.validateCurso(curso); err != nil {
		return fmt.Errorf("erro de validação: %w", err)
	}

	s.normalizeCurso(curso)

	return s.repo.Update(ctx, curso)
}

func (s *CursoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *CursoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Curso, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

// validateCurso performs business logic validation
func (s *CursoService) validateCurso(curso *models.Curso) error {
	if strings.TrimSpace(curso.Titulo) == "" {
		return fmt.Errorf("título é obrigatório")
	}

	if len(curso.Titulo) > 255 {
		return fmt.Errorf("título deve ter no máximo 255 caracteres")
	}

	if !curso.Modalidade.IsValid() {
		return fmt.Errorf("modalidade inválida: %s", curso.Modalidade)
	}

	if !curso.Status.IsValid() {
		return fmt.Errorf("status inválido: %s", curso.Status)
	}

	if curso.FormatoAula != "" && !curso.FormatoAula.IsValid() {
		return fmt.Errorf("formato de aula inválido: %s", curso.FormatoAula)
	}

	if curso.Turno != "" && !curso.Turno.IsValid() {
		return fmt.Errorf("turno inválido: %s", curso.Turno)
	}

	// Organization can be empty for some cases, just normalize it
	if curso.Organization != "" && len(curso.Organization) > 255 {
		return fmt.Errorf("organização deve ter no máximo 255 caracteres")
	}

	if curso.NumeroVagas < 0 {
		return fmt.Errorf("número de vagas deve ser positivo")
	}

	if curso.CargaHoraria < 0 {
		return fmt.Errorf("carga horária deve ser positiva")
	}

	// Validate URL fields if not empty
	if curso.InstitutionalLogo != "" && len(curso.InstitutionalLogo) > 500 {
		return fmt.Errorf("URL do logo institucional deve ter no máximo 500 caracteres")
	}
	
	if curso.CoverImage != "" && len(curso.CoverImage) > 500 {
		return fmt.Errorf("URL da imagem de capa deve ter no máximo 500 caracteres")
	}

	// Validate text fields length
	if len(curso.Theme) > 100 {
		return fmt.Errorf("tema deve ter no máximo 100 caracteres")
	}
	
	if len(curso.Workload) > 50 {
		return fmt.Errorf("carga de trabalho deve ter no máximo 50 caracteres")
	}
	
	if len(curso.TargetAudience) > 200 {
		return fmt.Errorf("público-alvo deve ter no máximo 200 caracteres")
	}

	return nil
}

// normalizeCurso normalizes data for consistency
func (s *CursoService) normalizeCurso(curso *models.Curso) {
	curso.Status = curso.Status.Normalize()
	curso.Modalidade = curso.Modalidade.Normalize()
	curso.FormatoAula = curso.FormatoAula.Normalize()
	curso.Titulo = strings.TrimSpace(curso.Titulo)
	curso.Organization = strings.TrimSpace(curso.Organization)

	// Normalize other text fields
	curso.Theme = strings.TrimSpace(curso.Theme)
	curso.Workload = strings.TrimSpace(curso.Workload)
	curso.TargetAudience = strings.TrimSpace(curso.TargetAudience)
	curso.PreRequisitos = strings.TrimSpace(curso.PreRequisitos)
	curso.Facilitator = strings.TrimSpace(curso.Facilitator)
	curso.Objectives = strings.TrimSpace(curso.Objectives)
	curso.ExpectedResults = strings.TrimSpace(curso.ExpectedResults)
	curso.ProgramContent = strings.TrimSpace(curso.ProgramContent)
	curso.Methodology = strings.TrimSpace(curso.Methodology)
	curso.ResourcesUsed = strings.TrimSpace(curso.ResourcesUsed)
	curso.MaterialUsed = strings.TrimSpace(curso.MaterialUsed)
	curso.TeachingMaterial = strings.TrimSpace(curso.TeachingMaterial)

	if curso.Status == "" {
		curso.Status = models.StatusCursoDraft
	}
} 