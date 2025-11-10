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

	if len(curso.Titulo) > 20000 {
		return fmt.Errorf("título deve ter no máximo 20000 caracteres")
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
	if curso.Organization != "" && len(curso.Organization) > 20000 {
		return fmt.Errorf("organização deve ter no máximo 20000 caracteres")
	}

	if curso.NumeroVagas < 0 {
		return fmt.Errorf("número de vagas deve ser positivo")
	}

	if curso.CargaHoraria < 0 {
		return fmt.Errorf("carga horária deve ser positiva")
	}

	// Validate URL fields if not empty
	if curso.InstitutionalLogo != "" && len(curso.InstitutionalLogo) > 20000 {
		return fmt.Errorf("URL do logo institucional deve ter no máximo 20000 caracteres")
	}

	if curso.CoverImage != "" && len(curso.CoverImage) > 20000 {
		return fmt.Errorf("URL da imagem de capa deve ter no máximo 20000 caracteres")
	}

	// Validate text fields length
	if len(curso.Theme) > 20000 {
		return fmt.Errorf("tema deve ter no máximo 20000 caracteres")
	}

	if len(curso.Workload) > 20000 {
		return fmt.Errorf("carga de trabalho deve ter no máximo 20000 caracteres")
	}

	if len(curso.TargetAudience) > 20000 {
		return fmt.Errorf("público-alvo deve ter no máximo 20000 caracteres")
	}

	// Validate locations and schedules
	if err := s.validateLocationClasses(curso.LocationClasses); err != nil {
		return fmt.Errorf("erro de validação em locations: %w", err)
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
	curso.Accessibility = strings.TrimSpace(curso.Accessibility)

	if curso.Status == "" {
		curso.Status = models.StatusCursoDraft
	}
}

// validateLocationClasses validates location classes and their schedules
func (s *CursoService) validateLocationClasses(locations []models.LocationClass) error {
	for i, location := range locations {
		// Validate address
		if len(strings.TrimSpace(location.Address)) < 10 {
			return fmt.Errorf("location[%d]: endereço deve ter pelo menos 10 caracteres", i)
		}
		if len(location.Address) > 20000 {
			return fmt.Errorf("location[%d]: endereço deve ter no máximo 20000 caracteres", i)
		}

		// Validate neighborhood
		if len(strings.TrimSpace(location.Neighborhood)) < 3 {
			return fmt.Errorf("location[%d]: bairro deve ter pelo menos 3 caracteres", i)
		}
		if len(location.Neighborhood) > 20000 {
			return fmt.Errorf("location[%d]: bairro deve ter no máximo 20000 caracteres", i)
		}

		// Validate schedules - at least 1 required
		if len(location.Schedules) < 1 {
			return fmt.Errorf("location[%d]: deve ter pelo menos 1 turma (schedule)", i)
		}

		// Validate each schedule
		if err := s.validateSchedules(location.Schedules, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSchedules validates course schedules
func (s *CursoService) validateSchedules(schedules []models.CourseSchedule, locationIndex int) error {
	for j, schedule := range schedules {
		// Validate vacancies
		if schedule.Vacancies < 1 || schedule.Vacancies > 1000 {
			return fmt.Errorf("location[%d].schedule[%d]: número de vagas deve estar entre 1 e 1000", locationIndex, j)
		}

		// Validate dates
		if schedule.ClassStartDate.IsZero() {
			return fmt.Errorf("location[%d].schedule[%d]: data de início é obrigatória", locationIndex, j)
		}

		if schedule.ClassEndDate.IsZero() {
			return fmt.Errorf("location[%d].schedule[%d]: data de término é obrigatória", locationIndex, j)
		}

		if schedule.ClassEndDate.Before(schedule.ClassStartDate) {
			return fmt.Errorf("location[%d].schedule[%d]: data de término deve ser maior ou igual à data de início", locationIndex, j)
		}

		// Validate class time
		if strings.TrimSpace(schedule.ClassTime) == "" {
			return fmt.Errorf("location[%d].schedule[%d]: horário da aula é obrigatório", locationIndex, j)
		}
		if len(schedule.ClassTime) > 20000 {
			return fmt.Errorf("location[%d].schedule[%d]: horário da aula deve ter no máximo 20000 caracteres", locationIndex, j)
		}

		// Validate class days
		if strings.TrimSpace(schedule.ClassDays) == "" {
			return fmt.Errorf("location[%d].schedule[%d]: dias da semana são obrigatórios", locationIndex, j)
		}
		if len(schedule.ClassDays) > 20000 {
			return fmt.Errorf("location[%d].schedule[%d]: dias da semana deve ter no máximo 20000 caracteres", locationIndex, j)
		}
	}

	return nil
}
