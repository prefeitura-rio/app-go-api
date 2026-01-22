package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// CitizenDataFetcher interface for fetching citizen data
type CitizenDataFetcher interface {
	SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
}

type InscricaoService struct {
	repo                     *repository.InscricaoRepository
	cursoRepo                *repository.CursoRepository
	citizenSnapshotRepo      *repository.CitizenSnapshotRepository
	citizenDataFetcher       CitizenDataFetcher
	emailNotificationService *EmailNotificationService
}

func NewInscricaoService(
	repo *repository.InscricaoRepository,
	cursoRepo *repository.CursoRepository,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	citizenDataFetcher CitizenDataFetcher,
	emailNotificationService *EmailNotificationService,
) *InscricaoService {
	return &InscricaoService{
		repo:                     repo,
		cursoRepo:                cursoRepo,
		citizenSnapshotRepo:      citizenSnapshotRepo,
		citizenDataFetcher:       citizenDataFetcher,
		emailNotificationService: emailNotificationService,
	}
}

func (s *InscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists and can accept enrollments (lightweight query)
	status, enrollmentStart, enrollmentEnd, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
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
	}

	// Fetch citizen data from RMI (on-demand sync) to populate Email and Phone
	if s.citizenDataFetcher != nil && inscricao.CPF != "" {
		citizenSnapshot, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, inscricao.CPF)
		if err != nil {
			// Log error but don't fail enrollment creation - use provided data as fallback
			fmt.Printf("[InscricaoService] Failed to fetch citizen data for CPF %s: %v\n", maskCPFForLog(inscricao.CPF), err)
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

	// Set default values
	inscricao.Status = models.StatusInscricaoPending
	inscricao.EnrolledAt = time.Now()
	inscricao.UpdatedAt = time.Now()

	// Create enrollment
	if err := s.repo.Create(ctx, inscricao); err != nil {
		return err
	}

	// Send "Em Análise" email (pending status)
	if s.emailNotificationService != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err == nil && curso != nil {
			// Send email asynchronously to avoid blocking the creation
			go func() {
				if err := s.emailNotificationService.SendEnrollmentCreatedEmail(context.Background(), inscricao, curso); err != nil {
					fmt.Printf("Failed to send enrollment created email: %v\n", err)
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
				}
				if emailErr != nil {
					fmt.Printf("Failed to send status change email: %v\n", emailErr)
				}
			}()
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
				}
				if emailErr != nil {
					fmt.Printf("Failed to send batch status change email for enrollment %s: %v\n", data.inscricao.ID, emailErr)
				}
			}
		}()
	}

	return updatedCount, nil
}

func (s *InscricaoService) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	// Validate course exists (lightweight query - only checks existence)
	_, _, _, err := s.cursoRepo.ValidateForEnrollment(ctx, cursoID)
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

// validateScheduleID validates that the schedule exists and belongs to the course
func (s *InscricaoService) validateScheduleID(ctx context.Context, scheduleID uuid.UUID, curso *models.Curso) error {
	// Check all locations and schedules in the course
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				// Schedule found and belongs to this course
				return nil
			}
		}
	}

	// Check remote class schedules
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				// Schedule found and belongs to this course
				return nil
			}
		}
	}

	return fmt.Errorf("schedule_id fornecido não pertence a este curso")
}

// EnrichWithPersonalInfo populates PersonalInfo for a single enrollment
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
		fmt.Printf("[InscricaoService] Failed to get citizen snapshot for CPF %s: %v\n", maskCPFForLog(inscricao.CPF), err)
		return
	}

	if snapshot == nil {
		fmt.Printf("[InscricaoService] No citizen snapshot found for CPF %s - citizen may not be synced yet\n", maskCPFForLog(inscricao.CPF))
		return
	}

	inscricao.PersonalInfo = snapshot.ToPersonalInfo()
}

// EnrichMultipleWithPersonalInfo populates PersonalInfo for multiple enrollments in a single batch query
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

	// Log statistics
	enrichedCount := 0
	for _, inscricao := range inscricoes {
		if snapshot, ok := snapshotMap[inscricao.CPF]; ok && snapshot != nil {
			inscricao.PersonalInfo = snapshot.ToPersonalInfo()
			enrichedCount++
		}
	}

	if enrichedCount == 0 && len(cpfs) > 0 {
		fmt.Printf("[InscricaoService] No citizen snapshots found for %d CPFs - citizens may not be synced yet\n", len(cpfs))
	}
}

// maskCPFForLog masks CPF for logging (shows only first 3 and last 2 digits)
func maskCPFForLog(cpf string) string {
	if len(cpf) < 5 {
		return "***"
	}
	return cpf[:3] + "******" + cpf[len(cpf)-2:]
}
