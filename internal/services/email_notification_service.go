package services

import (
	"context"
	"fmt"
	"log"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// EmailSender is an interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, req *clients.EmailRequest) error
}

// EmailNotificationService handles enrollment email notifications
type EmailNotificationService struct {
	dataRelayClient     EmailSender
	cursoRepo           *repository.CursoRepository
	orgaoSnapshotRepo   *repository.OrgaoSnapshotRepository
	citizenSnapshotRepo *repository.CitizenSnapshotRepository
	enabled             bool
	prefrioDomain       string
}

// NewEmailNotificationService creates a new email notification service
func NewEmailNotificationService(
	dataRelayClient EmailSender,
	cursoRepo *repository.CursoRepository,
	orgaoSnapshotRepo *repository.OrgaoSnapshotRepository,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	enabled bool,
	prefrioDomain string,
) *EmailNotificationService {
	return &EmailNotificationService{
		dataRelayClient:     dataRelayClient,
		cursoRepo:           cursoRepo,
		orgaoSnapshotRepo:   orgaoSnapshotRepo,
		citizenSnapshotRepo: citizenSnapshotRepo,
		enabled:             enabled,
		prefrioDomain:       prefrioDomain,
	}
}

// resolveEmail returns the most up-to-date email for an enrollment.
// It prefers the snapshot email (live RMI data) and falls back to the stored inscricao.Email.
// Both candidates are sanitized before use to avoid sending to malformed addresses.
func (s *EmailNotificationService) resolveEmail(ctx context.Context, inscricao *models.Inscricao) string {
	var email string
	if s.citizenSnapshotRepo != nil && inscricao.CPF != "" {
		snapshot, err := s.citizenSnapshotRepo.GetByCPF(ctx, inscricao.CPF)
		if err == nil && snapshot != nil && snapshot.Email != "" {
			email = models.SanitizeEmail(snapshot.Email)
		}
	}
	if email == "" {
		email = models.SanitizeEmail(inscricao.Email)
	}
	return email
}

// getOrgaoName fetches the organization name from orgao_snapshots or falls back to curso.Organization
func (s *EmailNotificationService) getOrgaoName(ctx context.Context, curso *models.Curso) string {
	if curso.OrgaoID != "" && s.orgaoSnapshotRepo != nil {
		snapshot, err := s.orgaoSnapshotRepo.GetByOrgaoID(ctx, curso.OrgaoID)
		if err == nil && snapshot != nil && snapshot.Name != "" {
			return snapshot.Name
		}
	}
	if curso.Organization != "" {
		return curso.Organization
	}
	return "órgão responsável"
}

// ScheduleInfo contains the schedule information for email templates
type ScheduleInfo struct {
	ClassTime      string
	ClassStartDate string
	ClassDays      string
	Address        string
}

// getScheduleInfo fetches schedule information from the enrollment's schedule_id
func (s *EmailNotificationService) getScheduleInfo(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) *ScheduleInfo {
	if inscricao.ScheduleID == nil || s.cursoRepo == nil {
		return nil
	}

	// Try to find in course schedules first
	courseSchedule, err := s.cursoRepo.GetCourseScheduleByID(ctx, *inscricao.ScheduleID)
	if err == nil && courseSchedule != nil {
		info := &ScheduleInfo{
			ClassTime:      courseSchedule.ClassTime,
			ClassStartDate: courseSchedule.ClassStartDate.Format("02/01/2006"),
			ClassDays:      courseSchedule.ClassDays,
		}
		if courseSchedule.Location != nil {
			info.Address = courseSchedule.Location.Address
		}
		return info
	}

	// Try remote schedule
	remoteSchedule, err := s.cursoRepo.GetRemoteScheduleByID(ctx, *inscricao.ScheduleID)
	if err == nil && remoteSchedule != nil {
		info := &ScheduleInfo{
			Address: "online",
		}
		if remoteSchedule.ClassTime != nil {
			info.ClassTime = *remoteSchedule.ClassTime
		}
		if remoteSchedule.ClassStartDate != nil {
			info.ClassStartDate = remoteSchedule.ClassStartDate.Format("02/01/2006")
		}
		if remoteSchedule.ClassDays != nil {
			info.ClassDays = *remoteSchedule.ClassDays
		}
		return info
	}

	return nil
}

// SendEnrollmentCreatedEmail sends "Em Análise" email when enrollment is created
func (s *EmailNotificationService) SendEnrollmentCreatedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment created email for %s", inscricao.Email)
		return nil
	}

	email := s.resolveEmail(ctx, inscricao)
	if email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, curso)
	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, orgaoName, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment created email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment created email to %s for course '%s'", email, curso.Titulo)
	return nil
}

// SendEnrollmentApprovedEmail sends "Inscrito" email when enrollment is approved
func (s *EmailNotificationService) SendEnrollmentApprovedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment approved email for %s", inscricao.Email)
		return nil
	}

	email := s.resolveEmail(ctx, inscricao)
	if email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, curso)
	scheduleInfo := s.getScheduleInfo(ctx, inscricao, curso)
	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, orgaoName, scheduleInfo, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment approved email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment approved email to %s for course '%s'", email, curso.Titulo)
	return nil
}

// SendEnrollmentRejectedEmail sends "Recusado" email when enrollment is rejected
func (s *EmailNotificationService) SendEnrollmentRejectedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment rejected email for %s", inscricao.Email)
		return nil
	}

	email := s.resolveEmail(ctx, inscricao)
	if email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	template := GetEnrollmentRejectedEmailTemplate(inscricao, curso, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment rejected email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment rejected email to %s for course '%s'", email, curso.Titulo)
	return nil
}

// SendScheduleChangedEmail sends a email confirming the new schedule
func (s *EmailNotificationService) SendScheduleChangedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping schedule changed email for %s", inscricao.Email)
		return nil
	}

	email := s.resolveEmail(ctx, inscricao)

	if email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, curso)
	scheduleInfo := s.getScheduleInfo(ctx, inscricao, curso)
	template := GetScheduleChangedEmailTemplate(inscricao, curso, scheduleInfo, orgaoName, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send schedule changed email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent schedule changed email to %s for course '%s'", email, curso.Titulo)

	return nil
}

// SendCandidaturaEnviadaEmail sends a email confirming the application was received
func (s *EmailNotificationService) SendCandidaturaEnviadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	if !s.enabled {
		logMessage := "[EmailNotificationService] Email notifications disabled - skipping received application email"

		if candidatura.Email != nil {
			logMessage += fmt.Sprintf(" for %s", *candidatura.Email)
		}

		log.Printf("%s", logMessage)
		return nil
	}

	email := s.resolveEmail(ctx, &models.Inscricao{
		Email: *candidatura.Email,
		CPF:   candidatura.CPF,
	})

	if email == "" {
		log.Printf("[EmailNotificationService] No email address for application ID %s - skipping", candidatura.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, &models.Curso{Organization: candidatura.Vaga.OrgaoParceiro.Name, OrgaoID: candidatura.Vaga.OrgaoParceiro.OrgaoID})
	template := GetCandidaturaEnviadaEmailTemplate(candidatura, candidatura.Vaga, candidatura.Vaga.Contratante, orgaoName, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send received application email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent received application email to %s for position '%s'", email, candidatura.Vaga.Titulo)

	return nil
}

// SendCandidaturaAprovadaEmail sends a email confirming the application approval
func (s *EmailNotificationService) SendCandidaturaAprovadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	if !s.enabled {
		logMessage := "[EmailNotificationService] Email notifications disabled - skipping approved application email"

		if candidatura.Email != nil {
			logMessage += fmt.Sprintf(" for %s", *candidatura.Email)
		}

		log.Printf("%s", logMessage)
		return nil
	}

	email := s.resolveEmail(ctx, &models.Inscricao{
		Email: *candidatura.Email,
		CPF:   candidatura.CPF,
	})

	if email == "" {
		log.Printf("[EmailNotificationService] No email address for application ID %s - skipping", candidatura.ID)
		return nil
	}

	template := GetCandidaturaAprovadaEmailTemplate(candidatura, candidatura.Vaga, candidatura.Vaga.Contratante)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send approved application email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent approved application email to %s for position '%s'", email, candidatura.Vaga.Titulo)

	return nil
}

// SendCandidaturaReprovadaEmail sends a email confirming the application failure
func (s *EmailNotificationService) SendCandidaturaReprovadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	if !s.enabled {
		logMessage := "[EmailNotificationService] Email notifications disabled - skipping failed application email"

		if candidatura.Email != nil {
			logMessage += fmt.Sprintf(" for %s", *candidatura.Email)
		}

		log.Printf("%s", logMessage)
		return nil
	}

	email := s.resolveEmail(ctx, &models.Inscricao{
		Email: *candidatura.Email,
		CPF:   candidatura.CPF,
	})

	if email == "" {
		log.Printf("[EmailNotificationService] No email address for application ID %s - skipping", candidatura.ID)
		return nil
	}

	template := GetCandidaturaReprovadaEmailTemplate(candidatura, candidatura.Vaga, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send failed application email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent failed application email to %s for position '%s'", email, candidatura.Vaga.Titulo)

	return nil
}
