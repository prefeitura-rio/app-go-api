package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/logger"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type InscricaoService struct {
	repo                     InscricaoRepositoryInterface
	cursoRepo                CursoRepositoryInterface
	citizenSnapshotRepo      CitizenSnapshotRepositoryInterface
	citizenDataFetcher       CitizenDataFetcher
	emailNotificationService *EmailNotificationService
	cfg                      *config.AppConfig
}

func NewInscricaoService(
	repo InscricaoRepositoryInterface,
	cursoRepo CursoRepositoryInterface,
	citizenSnapshotRepo CitizenSnapshotRepositoryInterface,
	citizenDataFetcher CitizenDataFetcher,
	emailNotificationService *EmailNotificationService,
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

func (s *InscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists and can accept enrollments (lightweight query)
	status, enrollmentStart, enrollmentEnd, autoApprove, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
	if err != nil {
		return err
	}

	// Check if enrollment is open
	if models.StatusCurso(status) != models.StatusCursoOpened {
		return fmt.Errorf("curso não está aberto para inscrições")
	}

	// Check enrollment dates
	now := time.Now()
	if enrollmentStart != nil && now.Before(*enrollmentStart) {
		return fmt.Errorf("período de inscrições ainda não iniciou")
	}
	if enrollmentEnd != nil && now.After(*enrollmentEnd) {
		return fmt.Errorf("período de inscrições já encerrou")
	}

	// Check if CPF is already enrolled
	exists, err := s.repo.ExistsByCPFAndCurso(ctx, inscricao.CPF, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição existente: %w", err)
	}
	if exists {
		return fmt.Errorf("CPF já inscrito neste curso")
	}

	// Validate schedule_id if provided
	// Only load full course with schedules when schedule validation is needed
	if inscricao.ScheduleID != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err != nil {
			return fmt.Errorf("erro ao verificar curso para validação de schedule: %w", err)
		}
		if curso == nil {
			return fmt.Errorf("curso não encontrado")
		}

		if err := s.validateScheduleID(ctx, *inscricao.ScheduleID, curso); err != nil {
			return err
		}

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

	// Fetch citizen data from RMI (on-demand sync) to populate Email and Phone
	if s.citizenDataFetcher != nil && inscricao.CPF != "" {
		citizenSnapshot, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, inscricao.CPF)
		if err != nil {
			// Log error but don't fail enrollment creation - use provided data as fallback
			logger.Warn("failed to fetch citizen data", "cpf", maskCPFForLog(inscricao.CPF), "error", err)
		} else if citizenSnapshot != nil {
			// Use RMI data for email and phone (overrides any value sent by frontend)
			if citizenSnapshot.Email != "" {
				inscricao.Email = citizenSnapshot.Email
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

	// Create enrollment
	if err := s.repo.Create(ctx, inscricao); err != nil {
		return err
	}

	// Send appropriate email based on status
	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			enrollmentCopy := *inscricao
			courseCopy := *curso
			approve := autoApprove
			s.emailNotificationService.Enqueue(func() {
				var emailErr error
				if approve {
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), &enrollmentCopy, &courseCopy)
				} else {
					emailErr = s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), &enrollmentCopy, &courseCopy)
				}
				if emailErr != nil {
					logger.Error("failed to send enrollment email", "error", emailErr)
				}
			})
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
		return fmt.Errorf("inscrição não encontrada")
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
			enrollmentCopy := *inscricao
			enrollmentCopy.Status = status
			enrollmentCopy.Reason = reason
			courseCopy := *curso
			newStatus := status
			s.emailNotificationService.Enqueue(func() {
				var emailErr error
				switch newStatus {
				case models.StatusInscricaoApproved:
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), &enrollmentCopy, &courseCopy)
				case models.StatusInscricaoRejected:
					emailErr = s.emailNotificationService.SendEnrollmentRejectedEmail(context.Background(), &enrollmentCopy, &courseCopy)
				}
				if emailErr != nil {
					logger.Error("failed to send status change email", "error", emailErr)
				}
			})
		}
	}

	return nil
}

func (s *InscricaoService) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	// Validate at least one ID provided
	if len(inscricaoIDs) == 0 {
		return 0, fmt.Errorf("nenhuma inscrição selecionada")
	}

	// Collect enrollment data for emails before update
	type emailData struct {
		inscricao *models.Inscricao
		oldStatus models.StatusInscricao
	}
	var enrollmentsForEmail []emailData

	if s.emailNotificationService != nil && (status == models.StatusInscricaoApproved || status == models.StatusInscricaoRejected) {
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

	// Enqueue emails for batch status updates
	if s.emailNotificationService != nil && len(enrollmentsForEmail) > 0 {
		newStatus := status
		newReason := reason

		for _, data := range enrollmentsForEmail {
			if data.oldStatus == newStatus {
				continue
			}

			enrollmentCopy := *data.inscricao
			enrollmentCopy.Status = newStatus
			enrollmentCopy.Reason = newReason
			cursoID := data.inscricao.CursoID

			s.emailNotificationService.Enqueue(func() {
				curso, err := s.cursoRepo.GetByID(context.Background(), cursoID)
				if err != nil || curso == nil {
					return
				}

				var emailErr error
				switch newStatus {
				case models.StatusInscricaoApproved:
					emailErr = s.emailNotificationService.SendEnrollmentApprovedEmail(context.Background(), &enrollmentCopy, curso)
				case models.StatusInscricaoRejected:
					emailErr = s.emailNotificationService.SendEnrollmentRejectedEmail(context.Background(), &enrollmentCopy, curso)
				}
				if emailErr != nil {
					logger.Error("failed to send batch status change email", "enrollment_id", enrollmentCopy.ID, "error", emailErr)
				}
			})
		}
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
		return fmt.Errorf("inscrição não encontrada")
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
		return fmt.Errorf("inscrição não encontrada")
	}
	if inscricao.CursoID != cursoID {
		return fmt.Errorf("inscrição não pertence ao curso especificado")
	}

	// Only allow certificate for approved or concluded enrollments
	if inscricao.Status != models.StatusInscricaoApproved && inscricao.Status != models.StatusInscricaoConcluded {
		return fmt.Errorf("certificado só pode ser atribuído a inscrições aprovadas ou concluídas")
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
		return fmt.Errorf("inscrição não encontrada")
	}

	// Validar que a inscrição pertence ao curso especificado
	if inscricao.CursoID != cursoID {
		return fmt.Errorf("inscrição não pertence ao curso especificado")
	}

	// Atualizar apenas os campos permitidos
	if updateData.Name != nil {
		inscricao.Name = *updateData.Name
	}
	if updateData.Email != nil {
		inscricao.Email = *updateData.Email
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
func (s *InscricaoService) validateScheduleID(ctx context.Context, scheduleID uuid.UUID, curso *models.Curso) error {
	// Check all locations and schedules in the course
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				if schedule.AcceptingEnrollments != nil && !*schedule.AcceptingEnrollments {
					return fmt.Errorf("esta turma não está aceitando inscrições no momento")
				}
				return nil
			}
		}
	}

	// Check remote class schedules
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				if schedule.AcceptingEnrollments != nil && !*schedule.AcceptingEnrollments {
					return fmt.Errorf("esta turma não está aceitando inscrições no momento")
				}
				return nil
			}
		}
	}

	return fmt.Errorf("schedule_id fornecido não pertence a este curso")
}

// EnrichWithPersonalInfo populates PersonalInfo for a single enrollment
// If no snapshot exists, it will fetch on-demand from RMI (for legacy enrollments)
func (s *InscricaoService) EnrichWithPersonalInfo(ctx context.Context, inscricao *models.Inscricao) {
	if s.citizenSnapshotRepo == nil {
		logger.Warn("citizenSnapshotRepo is nil, check if CitizenSync is enabled and Keycloak is configured")
		return
	}
	if inscricao == nil || inscricao.CPF == "" {
		return
	}

	snapshot, err := s.citizenSnapshotRepo.GetByCPF(ctx, inscricao.CPF)
	if err != nil {
		logger.Error("failed to get citizen snapshot", "cpf", maskCPFForLog(inscricao.CPF), "error", err)
		return
	}

	// If no snapshot exists and we have a fetcher, try to sync on-demand (for legacy enrollments)
	if snapshot == nil && s.citizenDataFetcher != nil {
		logger.Info("no snapshot found, attempting on-demand sync for legacy enrollment", "cpf", maskCPFForLog(inscricao.CPF))
		snapshot, err = s.citizenDataFetcher.SyncCitizenOnDemand(ctx, inscricao.CPF)
		if err != nil {
			logger.Error("on-demand sync failed", "cpf", maskCPFForLog(inscricao.CPF), "error", err)
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
		logger.Warn("citizenSnapshotRepo is nil, check if CitizenSync is enabled and Keycloak is configured")
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
		logger.Error("failed to get citizen snapshots", "error", err)
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
		logger.Info("syncing legacy enrollments on-demand", "count", len(missingCPFs))
		for _, cpf := range missingCPFs {
			snapshot, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, cpf)
			if err != nil {
				logger.Error("on-demand sync failed", "cpf", maskCPFForLog(cpf), "error", err)
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
		logger.Info("enriched enrollments with personal info", "enriched", enrichedCount, "total", len(cpfs))
	}
}

// maskCPFForLog masks CPF for logging (shows only first 3 and last 2 digits)
func maskCPFForLog(cpf string) string {
	if len(cpf) < 5 {
		return "***"
	}
	return cpf[:3] + "******" + cpf[len(cpf)-2:]
}

// ChangeSchedule allows a citizen to change their enrollment to a different schedule/class
func (s *InscricaoService) ChangeSchedule(ctx context.Context, inscricaoID uuid.UUID, userCPF string, request *models.ScheduleChangeRequest) (*models.Inscricao, error) {
	// Fetch current enrollment
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar inscrição: %w", err)
	}
	if inscricao == nil {
		return nil, fmt.Errorf("inscrição não encontrada")
	}

	// Validate that the enrollment belongs to the user
	if inscricao.CPF != userCPF {
		return nil, fmt.Errorf("você só pode alterar suas próprias inscrições")
	}

	// Validate status - cannot change if cancelled or concluded
	if inscricao.Status == models.StatusInscricaoCancelled || inscricao.Status == models.StatusInscricaoConcluded {
		return nil, fmt.Errorf("não é possível trocar de turma para inscrições com status '%s'", inscricao.Status)
	}

	// Get schedule deadline hours from config (default 48)
	deadlineHours := 48
	if s.cfg != nil && s.cfg.Enrollment.ScheduleChangeDeadlineHours > 0 {
		deadlineHours = s.cfg.Enrollment.ScheduleChangeDeadlineHours
	}

	// Parse the new schedule start date from EnrolledUnit
	if request.EnrolledUnit == nil || len(request.EnrolledUnit.Schedules) == 0 {
		return nil, fmt.Errorf("enrolled_unit com schedules é obrigatório")
	}

	// Get the class start date from the first schedule in EnrolledUnit
	newScheduleStartDateStr := request.EnrolledUnit.Schedules[0].ClassStartDate
	if newScheduleStartDateStr == "" {
		return nil, fmt.Errorf("data de início da turma não informada")
	}

	// Parse the date (try common formats)
	var newScheduleStartDate time.Time
	formats := []string{"2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05Z", "2006-01-02", time.RFC3339}
	for _, format := range formats {
		if parsed, err := time.Parse(format, newScheduleStartDateStr); err == nil {
			newScheduleStartDate = parsed
			break
		}
	}
	if newScheduleStartDate.IsZero() {
		return nil, fmt.Errorf("formato de data de início inválido: %s", newScheduleStartDateStr)
	}

	// Validate deadline - must be at least X hours before class start
	deadline := time.Now().Add(time.Duration(deadlineHours) * time.Hour)
	if newScheduleStartDate.Before(deadline) {
		return nil, fmt.Errorf("não é possível trocar de turma com menos de %d horas do início da aula", deadlineHours)
	}

	// Validate that the new schedule has available vacancies
	if request.ScheduleID != nil {
		scheduleIDs := []uuid.UUID{*request.ScheduleID}
		enrollmentCounts, err := s.cursoRepo.CountEnrollmentsByScheduleIDs(ctx, scheduleIDs)
		if err != nil {
			return nil, fmt.Errorf("erro ao verificar vagas: %w", err)
		}

		// Get schedule info to check vacancies
		schedule, err := s.cursoRepo.GetCourseScheduleByID(ctx, *request.ScheduleID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar turma: %w", err)
		}
		if schedule == nil {
			// Try remote schedule
			remoteSchedule, err := s.cursoRepo.GetRemoteScheduleByID(ctx, *request.ScheduleID)
			if err != nil {
				return nil, fmt.Errorf("erro ao buscar turma remota: %w", err)
			}
			if remoteSchedule == nil {
				return nil, fmt.Errorf("turma não encontrada")
			}
			if remoteSchedule.AcceptingEnrollments != nil && !*remoteSchedule.AcceptingEnrollments {
				return nil, fmt.Errorf("esta turma não está aceitando inscrições no momento")
			}
			// Check vacancies for remote schedule
			enrolledCount := enrollmentCounts[*request.ScheduleID]
			if int(enrolledCount) >= remoteSchedule.Vacancies {
				return nil, fmt.Errorf("não há vagas disponíveis nesta turma")
			}
		} else {
			if schedule.AcceptingEnrollments != nil && !*schedule.AcceptingEnrollments {
				return nil, fmt.Errorf("esta turma não está aceitando inscrições no momento")
			}
			// Check vacancies for course schedule
			enrolledCount := enrollmentCounts[*request.ScheduleID]
			if int(enrolledCount) >= schedule.Vacancies {
				return nil, fmt.Errorf("não há vagas disponíveis nesta turma")
			}
		}
	}

	// If status was approved, change to pending
	statusChanged := false
	if inscricao.Status == models.StatusInscricaoApproved {
		inscricao.Status = models.StatusInscricaoPending
		statusChanged = true
	}

	// Update schedule information
	if request.ScheduleID != nil {
		inscricao.ScheduleID = request.ScheduleID
	}
	inscricao.EnrolledUnit = request.EnrolledUnit
	inscricao.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.Update(ctx, inscricao); err != nil {
		return nil, fmt.Errorf("erro ao atualizar inscrição: %w", err)
	}

	// If status changed from approved to pending, send email notification
	if statusChanged && s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			enrollmentCopy := *inscricao
			courseCopy := *curso
			s.emailNotificationService.Enqueue(func() {
				if err := s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), &enrollmentCopy, &courseCopy); err != nil {
					logger.Error("failed to send schedule change email", "error", err)
				}
			})
		}
	}

	return inscricao, nil
}
