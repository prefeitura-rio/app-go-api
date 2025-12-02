package services

import (
	"context"
	"fmt"
	"log"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// EmailNotificationService handles enrollment email notifications
type EmailNotificationService struct {
	dataRelayClient *clients.DataRelayClient
	enabled         bool
	prefrioDomain   string
}

// NewEmailNotificationService creates a new email notification service
func NewEmailNotificationService(dataRelayClient *clients.DataRelayClient, enabled bool, prefrioDomain string) *EmailNotificationService {
	return &EmailNotificationService{
		dataRelayClient: dataRelayClient,
		enabled:         enabled,
		prefrioDomain:   prefrioDomain,
	}
}

// SendEnrollmentCreatedEmail sends "Em Análise" email when enrollment is created
func (s *EmailNotificationService) SendEnrollmentCreatedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment created email for %s", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{inscricao.Email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment created email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment created email to %s for course '%s'", inscricao.Email, curso.Titulo)
	return nil
}

// SendEnrollmentApprovedEmail sends "Inscrito" email when enrollment is approved
func (s *EmailNotificationService) SendEnrollmentApprovedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment approved email for %s", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{inscricao.Email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment approved email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment approved email to %s for course '%s'", inscricao.Email, curso.Titulo)
	return nil
}

// SendEnrollmentRejectedEmail sends "Recusado" email when enrollment is rejected
func (s *EmailNotificationService) SendEnrollmentRejectedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		log.Printf("[EmailNotificationService] Email notifications disabled - skipping enrollment rejected email for %s", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		log.Printf("[EmailNotificationService] No email address for enrollment ID %s - skipping", inscricao.ID)
		return nil
	}

	template := GetEnrollmentRejectedEmailTemplate(inscricao, curso, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{inscricao.Email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment rejected email: %w", err)
	}

	log.Printf("[EmailNotificationService] Sent enrollment rejected email to %s for course '%s'", inscricao.Email, curso.Titulo)
	return nil
}
