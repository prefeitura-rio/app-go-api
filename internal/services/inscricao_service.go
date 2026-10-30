package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

// EnrollmentRuleError marks a rejection the citizen can act on — the course is
// not open, the enrollment window is closed, the turma is not accepting people.
// These are not server faults, so handlers answer 400 instead of 500 and the
// portal can show the reason verbatim.
type EnrollmentRuleError struct {
	msg string
}

func (e *EnrollmentRuleError) Error() string { return e.msg }

// NewEnrollmentRuleError builds a business-rule rejection carrying a message
// meant to be shown to the citizen as-is.
func NewEnrollmentRuleError(msg string) error {
	return &EnrollmentRuleError{msg: msg}
}

func enrollmentRuleErrorf(format string, args ...any) error {
	return &EnrollmentRuleError{msg: fmt.Sprintf(format, args...)}
}

// Erros de domínio das inscrições. As mensagens são as mesmas de antes da
// tipagem, que o portal já mostra ao cidadão.
var (
	ErrCPFJaInscrito             = apperrors.Conflict("CPF já inscrito neste curso")
	ErrInscricaoNaoEncontrada    = apperrors.NotFound("inscrição não encontrada")
	ErrInscricaoDeOutroCurso     = apperrors.Invalid("inscrição não pertence ao curso especificado")
	ErrCertificadoStatusInvalido = apperrors.Invalid("certificado só pode ser atribuído a inscrições aprovadas ou concluídas")
	ErrCursoNaoEncontrado        = apperrors.NotFound("curso não encontrado")
)

// CitizenDataFetcher interface for fetching citizen data
type CitizenDataFetcher interface {
	SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
	SyncCitizenForced(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
}

type InscricaoService struct {
	repo                     InscricaoRepositoryInterface
	cursoRepo                CursoRepositoryInterface
	citizenSnapshotRepo      repository.CitizenSnapshotRepositoryInterface
	citizenDataFetcher       CitizenDataFetcher
	emailNotificationService EmailNotifier
	cfg                      *config.AppConfig
}

func NewInscricaoService(
	repo *repository.InscricaoRepository,
	cursoRepo *repository.CursoRepository,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	citizenDataFetcher CitizenDataFetcher,
	emailNotificationService EmailNotifier,
	cfg *config.AppConfig,
) *InscricaoService {
	return &InscricaoService{
		repo:                     repo,
		cursoRepo:                cursoRepo,
		citizenSnapshotRepo:      citizenSnapshotRepo,
		citizenDataFetcher:       citizenDataFetcher,
		emailNotificationService: emailNotificationService,
		cfg:                      cfg,
	}
}

func NewInscricaoServiceWithInterface(
	repo InscricaoRepositoryInterface,
	cursoRepo CursoRepositoryInterface,
	citizenSnapshotRepo repository.CitizenSnapshotRepositoryInterface,
	citizenDataFetcher CitizenDataFetcher,
	emailNotificationService EmailNotifier,
	cfg *config.AppConfig,
) *InscricaoService {
	return &InscricaoService{
		repo:                     repo,
		cursoRepo:                cursoRepo,
		citizenSnapshotRepo:      citizenSnapshotRepo,
		citizenDataFetcher:       citizenDataFetcher,
		emailNotificationService: emailNotificationService,
		cfg:                      cfg,
	}
}

// SetCitizenDataFetcher allows late-binding of the citizen data fetcher, e.g. after
// the citizen sync worker has been started by the application bootstrap.
func (s *InscricaoService) SetCitizenDataFetcher(fetcher CitizenDataFetcher) {
	s.citizenDataFetcher = fetcher
}

// SetEmailNotifier replaces the email notifier after construction, e.g. to swap in
// the queued EmailWorker once it has been started by the application bootstrap.
func (s *InscricaoService) SetEmailNotifier(notifier EmailNotifier) {
	s.emailNotificationService = notifier
}

func (s *InscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists and can accept enrollments (lightweight query)
	status, enrollmentStart, enrollmentEnd, autoApprove, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
	if err != nil {
		return err
	}

	normalizedStatus := models.StatusCurso(status).Normalize()
	if normalizedStatus != models.StatusCursoOpened && normalizedStatus != models.StatusCursoPublished {
		return enrollmentRuleErrorf("curso não está aberto para inscrições")
	}

	// Check enrollment dates
	now := time.Now()
	if enrollmentStart != nil && now.Before(*enrollmentStart) {
		return enrollmentRuleErrorf("período de inscrições ainda não iniciou")
	}
	if enrollmentEnd != nil && now.After(*enrollmentEnd) {
		return enrollmentRuleErrorf("período de inscrições já encerrou")
	}

	// Check if CPF is already enrolled
	exists, err := s.repo.ExistsByCPFAndCurso(ctx, inscricao.CPF, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição existente: %w", err)
	}
	if exists {
		return ErrCPFJaInscrito
	}

	// The enrolled_unit is never taken from the request body: it is rebuilt below
	// from the course data for the validated schedule_id, so a citizen cannot
	// persist a unit or schedule that skipped validateScheduleID.
	inscricao.EnrolledUnit = nil

	// Validate schedule_id if provided
	// Only load full course with schedules when schedule validation is needed
	if inscricao.ScheduleID != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err != nil {
			return fmt.Errorf("erro ao verificar curso para validação de schedule: %w", err)
		}
		if curso == nil {
			return ErrCursoNaoEncontrado
		}

		if err := s.validateScheduleID(ctx, *inscricao.ScheduleID, curso, true); err != nil {
			return err
		}
		inscricao.EnrolledUnit = enrolledUnitForSchedule(*inscricao.ScheduleID, curso)

		// If auto-approve is enabled, check vacancy availability to prevent overbooking
		// Note: This is an advisory check (COUNT then INSERT, no lock). In high-concurrency
		// scenarios, multiple requests could pass the check and exceed vacancies. This matches
		// the existing pattern used in ChangeSchedule and manual approval flows. For strict
		// enforcement, consider adding DB constraints or SELECT FOR UPDATE locks.
		if autoApprove {
			vacancies := s.findScheduleVacancies(*inscricao.ScheduleID, curso)
			if vacancies > 0 {
				enrolledCount, err := s.cursoRepo.CountEnrollmentsByScheduleID(ctx, *inscricao.ScheduleID)
				// Fail-safe: if we cannot verify capacity or schedule is full, disable auto-approve
				if err != nil || int(enrolledCount) >= vacancies {
					autoApprove = false
				}
			}
		}
	}

	// Fetch citizen data from RMI (forced sync) to populate Email and Phone
	if s.citizenDataFetcher != nil && inscricao.CPF != "" {
		citizenSnapshot, err := s.citizenDataFetcher.SyncCitizenForced(ctx, inscricao.CPF)
		if err != nil {
			// Log error but don't fail enrollment creation - use provided data as fallback
			fmt.Printf("[InscricaoService] Failed to fetch citizen data for CPF %s: %v\n", utils.MaskCPFForLog(inscricao.CPF), err)
		} else if citizenSnapshot != nil {
			// Use RMI data for email and phone (overrides any value sent by frontend)
			if citizenSnapshot.Email != "" {
				inscricao.Email = models.SanitizeEmail(citizenSnapshot.Email)
			}
			if citizenSnapshot.Celular != "" {
				inscricao.Phone = citizenSnapshot.Celular
			}
			if citizenSnapshot.Nome != "" && inscricao.Name == "" {
				inscricao.Name = citizenSnapshot.Nome
			}
		}
	}

	// Set initial status based on auto-approve setting
	if autoApprove {
		inscricao.Status = models.StatusInscricaoApproved
	} else {
		inscricao.Status = models.StatusInscricaoPending
	}
	inscricao.EnrolledAt = time.Now()
	inscricao.UpdatedAt = time.Now()
	origem := models.OrigemInscricaoPortalCidadao
	inscricao.OrigemInscricao = &origem

	// Create enrollment
	if err := s.repo.Create(ctx, inscricao); err != nil {
		return err
	}

	// Send appropriate email based on status
	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			go func() {
				var emailErr error
				if autoApprove {
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), inscricao, curso)
				} else {
					emailErr = s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), inscricao, curso)
				}
				if emailErr != nil {
					log.Printf("[InscricaoService] falha ao enviar email de inscrição: %v", emailErr)
				}
			}()
		}
	}

	return nil
}

// CreateByAdmin creates an enrollment bypassing course status and date validation.
// Used by internal flows (CSV import) where an admin is enrolling participants
// in courses that may be closed or past their enrollment period.
func (s *InscricaoService) CreateByAdmin(ctx context.Context, inscricao *models.Inscricao) error {
	_, _, _, autoApprove, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
	if err != nil {
		return err
	}

	exists, err := s.repo.ExistsByCPFAndCurso(ctx, inscricao.CPF, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição existente: %w", err)
	}
	if exists {
		return ErrCPFJaInscrito
	}

	if inscricao.ScheduleID != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err != nil {
			return fmt.Errorf("erro ao verificar curso para validação de schedule: %w", err)
		}
		if curso == nil {
			return ErrCursoNaoEncontrado
		}
		if err := s.validateScheduleID(ctx, *inscricao.ScheduleID, curso, false); err != nil {
			return err
		}
		if autoApprove {
			vacancies := s.findScheduleVacancies(*inscricao.ScheduleID, curso)
			if vacancies > 0 {
				enrolledCount, err := s.cursoRepo.CountEnrollmentsByScheduleID(ctx, *inscricao.ScheduleID)
				if err != nil || int(enrolledCount) >= vacancies {
					autoApprove = false
				}
			}
		}
	}

	// The email/phone on the enrollment are the contact provided by the órgão/secretaria
	// during manual import — that is the whole point of this flow, so they must be preserved.
	// Sanitize the provided email (parity with CreateManual) before persisting.
	inscricao.Email = models.SanitizeEmail(inscricao.Email)

	// Sync citizen data from RMI on-demand so the official contact is available downstream
	// (citizen_snapshots → personal_info for display/comparison). Do NOT overwrite the
	// órgão-provided email/phone with the RMI snapshot; only fill in a missing name.
	if s.citizenDataFetcher != nil && inscricao.CPF != "" {
		citizenSnapshot, err := s.citizenDataFetcher.SyncCitizenForced(ctx, inscricao.CPF)
		if err != nil {
			fmt.Printf("[InscricaoService] Failed to fetch citizen data for CPF %s: %v\n", utils.MaskCPFForLog(inscricao.CPF), err)
		} else if citizenSnapshot != nil {
			if citizenSnapshot.Nome != "" && inscricao.Name == "" {
				inscricao.Name = citizenSnapshot.Nome
			}
		}
	}

	if autoApprove {
		inscricao.Status = models.StatusInscricaoApproved
	} else {
		inscricao.Status = models.StatusInscricaoPending
	}
	inscricao.EnrolledAt = time.Now()
	inscricao.UpdatedAt = time.Now()
	origemAdmin := models.OrigemInscricaoSecretaria
	inscricao.OrigemInscricao = &origemAdmin

	if err := s.repo.Create(ctx, inscricao); err != nil {
		return err
	}

	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			go func() {
				var emailErr error
				if autoApprove {
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), inscricao, curso)
				} else {
					emailErr = s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), inscricao, curso)
				}
				if emailErr != nil {
					log.Printf("[InscricaoService] falha ao enviar email de inscrição: %v", emailErr)
				}
			}()
		}
	}

	return nil
}

// CreateManual creates an enrollment bypassing date restrictions and citizen data override.
// Used by admins to manually enroll someone regardless of the enrollment period or course status.
func (s *InscricaoService) CreateManual(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists
	_, _, _, autoApprove, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
	if err != nil {
		return err
	}

	// NOTE: status and date validation intentionally skipped for manual admin enrollments

	// Check if CPF is already enrolled
	exists, err := s.repo.ExistsByCPFAndCurso(ctx, inscricao.CPF, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição existente: %w", err)
	}
	if exists {
		return ErrCPFJaInscrito
	}

	// Validate schedule_id if provided
	if inscricao.ScheduleID != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err != nil {
			return fmt.Errorf("erro ao verificar curso para validação de schedule: %w", err)
		}
		if curso == nil {
			return ErrCursoNaoEncontrado
		}
		if err := s.validateScheduleID(ctx, *inscricao.ScheduleID, curso, false); err != nil {
			return err
		}
		// Check vacancy availability to prevent overbooking when auto-approving
		if autoApprove {
			vacancies := s.findScheduleVacancies(*inscricao.ScheduleID, curso)
			if vacancies > 0 {
				enrolledCount, err := s.cursoRepo.CountEnrollmentsByScheduleID(ctx, *inscricao.ScheduleID)
				if err != nil || int(enrolledCount) >= vacancies {
					autoApprove = false
				}
			}
		}
	}

	// NOTE: citizen data fetching is skipped — admin provides data explicitly

	// Sanitize admin-provided email before persisting
	inscricao.Email = models.SanitizeEmail(inscricao.Email)

	if autoApprove {
		inscricao.Status = models.StatusInscricaoApproved
	} else {
		inscricao.Status = models.StatusInscricaoPending
	}
	inscricao.EnrolledAt = time.Now()
	inscricao.UpdatedAt = time.Now()
	origemManual := models.OrigemInscricaoSecretaria
	inscricao.OrigemInscricao = &origemManual

	if err := s.repo.Create(ctx, inscricao); err != nil {
		return err
	}

	// Send confirmation email (same logic as Create)
	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			go func() {
				var emailErr error
				if autoApprove {
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), inscricao, curso)
				} else {
					emailErr = s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), inscricao, curso)
				}
				if emailErr != nil {
					log.Printf("[InscricaoService] falha ao enviar email de inscrição manual: %v", emailErr)
				}
			}()
		}
	}

	return nil
}

func (s *InscricaoService) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *InscricaoService) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetByCursoID(ctx, cursoID, filter, pageSize, offset)
}

func (s *InscricaoService) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	// Validate enrollment exists
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return ErrInscricaoNaoEncontrada
	}

	// Store old status for email decision
	oldStatus := inscricao.Status

	// Update status
	if err := s.repo.UpdateStatus(ctx, inscricaoID, status, reason, adminNotes); err != nil {
		return err
	}

	// Send status change emails
	if s.emailNotificationService != nil && oldStatus != status {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			// Update inscricao with new status and reason for email template
			inscricao.Status = status
			inscricao.Reason = reason

			// Send email asynchronously
			go func() {
				var emailErr error
				switch status {
				case models.StatusInscricaoApproved:
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), inscricao, curso)
				case models.StatusInscricaoRejected:
					emailErr = s.emailNotificationService.SendEnrollmentRejectedEmail(context.Background(), inscricao, curso)
				case models.StatusInscricaoConcluded:
					emailErr = s.emailNotificationService.SendEnrollmentConcludedEmail(context.Background(), inscricao, curso)
				}
				if emailErr != nil {
					log.Printf("[InscricaoService] falha ao enviar email de mudança de status: %v", emailErr)
				}
			}()
		}
	}

	return nil
}

func (s *InscricaoService) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	// Validate at least one ID provided
	if len(inscricaoIDs) == 0 {
		return 0, apperrors.Invalid("nenhuma inscrição selecionada")
	}

	// Collect enrollment data for emails before update
	type emailData struct {
		inscricao *models.Inscricao
		oldStatus models.StatusInscricao
	}
	var enrollmentsForEmail []emailData

	if s.emailNotificationService != nil && (status == models.StatusInscricaoApproved || status == models.StatusInscricaoRejected || status == models.StatusInscricaoConcluded) {
		for _, id := range inscricaoIDs {
			inscricao, err := s.repo.GetByID(ctx, id)
			if err == nil && inscricao != nil {
				enrollmentsForEmail = append(enrollmentsForEmail, emailData{
					inscricao: inscricao,
					oldStatus: inscricao.Status,
				})
			}
		}
	}

	// Update statuses
	updatedCount, err := s.repo.UpdateMultipleStatus(ctx, inscricaoIDs, status, reason, adminNotes)
	if err != nil {
		return updatedCount, err
	}

	// Send emails asynchronously for batch status updates
	if s.emailNotificationService != nil && len(enrollmentsForEmail) > 0 {
		// Capture values for goroutine to avoid race conditions
		newStatus := status
		newReason := reason

		go func() {
			for _, data := range enrollmentsForEmail {
				// Skip if status didn't change
				if data.oldStatus == newStatus {
					continue
				}

				curso, err := s.cursoRepo.GetByID(context.Background(), data.inscricao.CursoID)
				if err != nil || curso == nil {
					continue
				}

				// Create a copy of inscricao with updated status for email
				inscricaoCopy := *data.inscricao
				inscricaoCopy.Status = newStatus
				inscricaoCopy.Reason = newReason

				var emailErr error
				switch newStatus {
				case models.StatusInscricaoApproved:
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), &inscricaoCopy, curso)
				case models.StatusInscricaoRejected:
					emailErr = s.emailNotificationService.SendEnrollmentRejectedEmail(context.Background(), &inscricaoCopy, curso)
				case models.StatusInscricaoConcluded:
					emailErr = s.emailNotificationService.SendEnrollmentConcludedEmail(context.Background(), &inscricaoCopy, curso)
				}
				if emailErr != nil {
					log.Printf("[InscricaoService] falha ao enviar email de mudança de status em lote (inscrição %s): %v", data.inscricao.ID, emailErr)
				}
			}
		}()
	}

	return updatedCount, nil
}

func (s *InscricaoService) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	// Validate course exists (lightweight query - only checks existence)
	_, _, _, _, err := s.cursoRepo.ValidateForEnrollment(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetSummaryByCursoID(ctx, cursoID)
}

func (s *InscricaoService) Delete(ctx context.Context, id uuid.UUID) error {
	// Validate enrollment exists
	inscricao, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return ErrInscricaoNaoEncontrada
	}

	return s.repo.Delete(ctx, id)
}

func (s *InscricaoService) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	return s.repo.ListByCPF(ctx, cpf, filter, offset, limit)
}

func (s *InscricaoService) UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error {
	// Verify enrollment exists and belongs to the course
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return ErrInscricaoNaoEncontrada
	}
	if inscricao.CursoID != cursoID {
		return ErrInscricaoDeOutroCurso
	}

	// Only allow certificate for approved or concluded enrollments
	if inscricao.Status != models.StatusInscricaoApproved && inscricao.Status != models.StatusInscricaoConcluded {
		return ErrCertificadoStatusInvalido
	}

	// Update certificate URL
	return s.repo.UpdateCertificate(ctx, inscricaoID, certificateURL)
}

func (s *InscricaoService) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	// Buscar inscrição existente
	inscricao, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return ErrInscricaoNaoEncontrada
	}

	// Validar que a inscrição pertence ao curso especificado
	if inscricao.CursoID != cursoID {
		return ErrInscricaoDeOutroCurso
	}

	// Atualizar apenas os campos permitidos
	if updateData.Name != nil {
		inscricao.Name = *updateData.Name
	}
	if updateData.Email != nil {
		inscricao.Email = models.SanitizeEmail(*updateData.Email)
	}
	if updateData.Phone != nil {
		inscricao.Phone = *updateData.Phone
	}
	if updateData.CustomFieldsData != nil {
		inscricao.CustomFieldsData = updateData.CustomFieldsData
	}
	if updateData.AdminNotes != nil {
		inscricao.AdminNotes = *updateData.AdminNotes
	}
	if updateData.EnrolledUnit != nil {
		inscricao.EnrolledUnit = updateData.EnrolledUnit
	}

	inscricao.UpdatedAt = time.Now()

	return s.repo.Update(ctx, inscricao)
}

// countCourseSchedules returns the total number of schedules across all location classes and remote class
func countCourseSchedules(curso *models.Curso) int {
	count := 0
	for _, loc := range curso.LocationClasses {
		count += len(loc.Schedules)
	}
	if curso.RemoteClass != nil {
		count += len(curso.RemoteClass.Schedules)
	}
	return count
}

// findScheduleVacancies returns the vacancy limit for a schedule within a course
func (s *InscricaoService) findScheduleVacancies(scheduleID uuid.UUID, curso *models.Curso) int {
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				return schedule.Vacancies
			}
		}
	}
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				return schedule.Vacancies
			}
		}
	}
	return 0
}

// validateScheduleID validates that the schedule exists, belongs to the course, and is accepting enrollments
func (s *InscricaoService) validateScheduleID(_ context.Context, scheduleID uuid.UUID, curso *models.Curso, enforceWindow bool) error {
	// checkTurma validates a matched turma: it must be accepting enrollments and,
	// when enforceWindow is set (citizen flow), be within its own enrollment window.
	checkTurma := func(accepting *bool, enrollmentStart, enrollmentEnd *time.Time) error {
		if accepting != nil && !*accepting {
			return enrollmentRuleErrorf("esta turma não está aceitando inscrições no momento")
		}
		if enforceWindow {
			now := time.Now()
			if enrollmentStart != nil && now.Before(*enrollmentStart) {
				return enrollmentRuleErrorf("as inscrições desta turma ainda não iniciaram")
			}
			if enrollmentEnd != nil && now.After(*enrollmentEnd) {
				return enrollmentRuleErrorf("as inscrições desta turma já encerraram")
			}
		}
		return nil
	}

	// Check all locations and schedules in the course
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				return checkTurma(schedule.AcceptingEnrollments, schedule.EnrollmentStartDate, schedule.EnrollmentEndDate)
			}
		}
	}

	// Check remote class schedules
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				return checkTurma(schedule.AcceptingEnrollments, schedule.EnrollmentStartDate, schedule.EnrollmentEndDate)
			}
		}
	}

	return enrollmentRuleErrorf("a turma selecionada não pertence a este curso")
}

// enrolledUnitForSchedule builds, from the course data, the enrolled_unit
// snapshot of the location (or remote class) that holds scheduleID, with all of
// its schedules. It returns nil when the course has no such schedule.
func enrolledUnitForSchedule(scheduleID uuid.UUID, curso *models.Curso) *models.EnrolledUnit {
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID != scheduleID {
				continue
			}
			schedules := make([]models.EnrolledUnitSchedule, 0, len(location.Schedules))
			for _, sch := range location.Schedules {
				schedules = append(schedules, models.EnrolledUnitSchedule{
					ID:             sch.ID.String(),
					LocationID:     location.ID.String(),
					Vacancies:      sch.Vacancies,
					ClassStartDate: sch.ClassStartDate.Format(time.RFC3339),
					ClassEndDate:   sch.ClassEndDate.Format(time.RFC3339),
					ClassTime:      sch.ClassTime,
					ClassDays:      sch.ClassDays,
					CreatedAt:      sch.CreatedAt.Format(time.RFC3339),
					UpdatedAt:      sch.UpdatedAt.Format(time.RFC3339),
				})
			}
			return &models.EnrolledUnit{
				ID:               location.ID.String(),
				CursoID:          curso.ID,
				Address:          location.Address,
				Neighborhood:     location.Neighborhood,
				NeighborhoodZone: location.NeighborhoodZone,
				Schedules:        schedules,
				CreatedAt:        location.CreatedAt.Format(time.RFC3339),
				UpdatedAt:        location.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	if curso.RemoteClass != nil {
		remote := curso.RemoteClass
		for _, schedule := range remote.Schedules {
			if schedule.ID != scheduleID {
				continue
			}
			schedules := make([]models.EnrolledUnitSchedule, 0, len(remote.Schedules))
			for _, sch := range remote.Schedules {
				schedules = append(schedules, models.EnrolledUnitSchedule{
					ID:             sch.ID.String(),
					LocationID:     remote.ID.String(),
					Vacancies:      sch.Vacancies,
					ClassStartDate: formatOptionalTime(sch.ClassStartDate),
					ClassEndDate:   formatOptionalTime(sch.ClassEndDate),
					ClassTime:      derefString(sch.ClassTime),
					ClassDays:      derefString(sch.ClassDays),
					CreatedAt:      sch.CreatedAt.Format(time.RFC3339),
					UpdatedAt:      sch.UpdatedAt.Format(time.RFC3339),
				})
			}
			return &models.EnrolledUnit{
				ID:        remote.ID.String(),
				CursoID:   curso.ID,
				Schedules: schedules,
				CreatedAt: remote.CreatedAt.Format(time.RFC3339),
				UpdatedAt: remote.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	return nil
}

// scheduleClassStartDate returns the class start date of scheduleID in the
// course (location or remote class), or nil when the course has no such
// schedule or the schedule has no start date.
func scheduleClassStartDate(scheduleID uuid.UUID, curso *models.Curso) *time.Time {
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				if schedule.ClassStartDate.IsZero() {
					return nil
				}
				start := schedule.ClassStartDate
				return &start
			}
		}
	}
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				return schedule.ClassStartDate
			}
		}
	}
	return nil
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// EnrichWithPersonalInfo populates PersonalInfo for a single enrollment
// If no snapshot exists, it will fetch on-demand from RMI (for legacy enrollments)
func (s *InscricaoService) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
	if s.citizenSnapshotRepo == nil {
		fmt.Println("[InscricaoService] EnrichWithPersonalInfo: citizenSnapshotRepo is nil - check if CitizenSync is enabled and Keycloak is configured")
		return
	}
	if inscricao == nil || inscricao.CPF == "" {
		return
	}

	snapshot, err := s.citizenSnapshotRepo.GetByCPF(ctx, inscricao.CPF)
	if err != nil {
		fmt.Printf("[InscricaoService] Failed to get citizen snapshot for CPF %s: %v\n", utils.MaskCPFForLog(inscricao.CPF), err)
		return
	}

	// If no snapshot exists and we have a fetcher, try to sync on-demand (for legacy enrollments)
	if snapshot == nil && s.citizenDataFetcher != nil {
		fmt.Printf("[InscricaoService] No snapshot for CPF %s - attempting on-demand sync for legacy enrollment\n", utils.MaskCPFForLog(inscricao.CPF))
		snapshot, err = s.citizenDataFetcher.SyncCitizenOnDemand(ctx, inscricao.CPF)
		if err != nil {
			fmt.Printf("[InscricaoService] On-demand sync failed for CPF %s: %v\n", utils.MaskCPFForLog(inscricao.CPF), err)
			return
		}
	}

	if snapshot == nil {
		return
	}

	inscricao.PersonalInfo = snapshot.ToPersonalInfo()
}

// EnrichMultipleWithPersonalInfo populates PersonalInfo for multiple enrollments in a single batch query
// For legacy enrollments without snapshots, it will fetch on-demand from RMI
func (s *InscricaoService) EnrichMultipleWithPersonalInfo(ctx context.Context, inscricoes []*models.Inscricao) {
	if s.citizenSnapshotRepo == nil {
		fmt.Println("[InscricaoService] EnrichMultipleWithPersonalInfo: citizenSnapshotRepo is nil - check if CitizenSync is enabled and Keycloak is configured")
		return
	}
	if len(inscricoes) == 0 {
		return
	}

	// Collect unique CPFs
	cpfSet := make(map[string]struct{})
	for _, inscricao := range inscricoes {
		if inscricao.CPF != "" {
			cpfSet[inscricao.CPF] = struct{}{}
		}
	}

	cpfs := make([]string, 0, len(cpfSet))
	for cpf := range cpfSet {
		cpfs = append(cpfs, cpf)
	}

	if len(cpfs) == 0 {
		return
	}

	// Batch query for all snapshots
	snapshotMap, err := s.citizenSnapshotRepo.GetByCPFs(ctx, cpfs)
	if err != nil {
		fmt.Printf("[InscricaoService] Failed to get citizen snapshots: %v\n", err)
		return
	}

	// Find CPFs without snapshots (legacy enrollments)
	var missingCPFs []string
	for _, cpf := range cpfs {
		if _, ok := snapshotMap[cpf]; !ok {
			missingCPFs = append(missingCPFs, cpf)
		}
	}

	// Sync missing CPFs on-demand if we have a fetcher
	if len(missingCPFs) > 0 && s.citizenDataFetcher != nil {
		fmt.Printf("[InscricaoService] Syncing %d legacy enrollments on-demand\n", len(missingCPFs))
		for _, cpf := range missingCPFs {
			snapshot, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, cpf)
			if err != nil {
				fmt.Printf("[InscricaoService] On-demand sync failed for CPF %s: %v\n", utils.MaskCPFForLog(cpf), err)
				continue
			}
			if snapshot != nil {
				snapshotMap[cpf] = snapshot
			}
		}
	}

	// Enrich enrollments with snapshots
	enrichedCount := 0
	for _, inscricao := range inscricoes {
		if snapshot, ok := snapshotMap[inscricao.CPF]; ok && snapshot != nil {
			inscricao.PersonalInfo = snapshot.ToPersonalInfo()
			enrichedCount++
		}
	}

	if enrichedCount < len(cpfs) {
		fmt.Printf("[InscricaoService] Enriched %d/%d enrollments with personal info\n", enrichedCount, len(cpfs))
	}
}

// ChangeSchedule allows a citizen to change their enrollment to a different schedule/class
func (s *InscricaoService) ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	// Fetch current enrollment
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar inscrição: %w", err)
	}
	if inscricao == nil {
		return nil, ErrInscricaoNaoEncontrada
	}

	// Validate that the enrollment belongs to the user
	if inscricao.CPF != userCPF {
		return nil, fmt.Errorf("você só pode alterar suas próprias inscrições")
	}

	// Validate status - cannot change if cancelled or concluded
	if inscricao.Status == models.StatusInscricaoCancelled || inscricao.Status == models.StatusInscricaoConcluded {
		return nil, fmt.Errorf("não é possível trocar de turma para inscrições com status '%s'", inscricao.Status)
	}

	// Validate that the course has more than one schedule
	curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar curso: %w", err)
	}
	if curso == nil {
		return nil, fmt.Errorf("curso não encontrado")
	}
	if countCourseSchedules(curso) <= 1 {
		return nil, fmt.Errorf("não é possível trocar de turma: este curso possui apenas uma turma disponível")
	}

	// The new turma comes only from a schedule_id of this enrollment's course.
	// The enrolled_unit sent in the body is never read: it is rebuilt below from
	// the course data, so a citizen cannot persist a unit, schedule or date that
	// skipped validation (DEV-136, same rule as Create since DEV-134).
	if request.ScheduleID == nil {
		return nil, enrollmentRuleErrorf("não é possível trocar de turma: selecione a nova turma")
	}
	scheduleID := *request.ScheduleID

	// Belongs to the course and is accepting enrollments. The turma's own
	// enrollment window is not enforced here, as before: the citizen is already
	// enrolled in the course.
	if err := s.validateScheduleID(ctx, scheduleID, curso, false); err != nil {
		return nil, err
	}

	// Get schedule deadline hours from config (default 48)
	deadlineHours := 48
	if s.cfg != nil && s.cfg.Enrollment.ScheduleChangeDeadlineHours > 0 {
		deadlineHours = s.cfg.Enrollment.ScheduleChangeDeadlineHours
	}

	// Validate deadline against the real class start of the new schedule -
	// must be at least X hours before class start
	newScheduleStartDate := scheduleClassStartDate(scheduleID, curso)
	if newScheduleStartDate == nil {
		return nil, enrollmentRuleErrorf("não é possível trocar de turma: a turma não tem data de início definida")
	}
	deadline := time.Now().Add(time.Duration(deadlineHours) * time.Hour)
	if newScheduleStartDate.Before(deadline) {
		return nil, fmt.Errorf("não é possível trocar de turma com menos de %d horas do início da aula", deadlineHours)
	}

	// Validate that the new schedule has available vacancies (limit from the course data)
	enrollmentCounts, err := s.cursoRepo.CountEnrollmentsByScheduleIDs(ctx, []uuid.UUID{scheduleID})
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar vagas: %w", err)
	}
	if int(enrollmentCounts[scheduleID]) >= s.findScheduleVacancies(scheduleID, curso) {
		return nil, fmt.Errorf("não há vagas disponíveis nesta turma")
	}

	// Update schedule information
	inscricao.ScheduleID = &scheduleID
	inscricao.EnrolledUnit = enrolledUnitForSchedule(scheduleID, curso)
	inscricao.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.Update(ctx, inscricao); err != nil {
		return nil, fmt.Errorf("erro ao atualizar inscrição: %w", err)
	}

	// Send email notification on schedule change
	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			go func() {
				if err := s.emailNotificationService.SendScheduleChangedEmail(context.Background(), inscricao, curso); err != nil {
					log.Printf("[InscricaoService] falha ao enviar email de alteração de horário: %v", err)
				}
			}()
		}
	}

	return inscricao, nil
}
