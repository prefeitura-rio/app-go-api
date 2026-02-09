package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/logger"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type emailTask struct {
	fn func()
}

type EmailNotificationService struct {
	dataRelayClient   *clients.DataRelayClient
	cursoRepo         CursoRepositoryInterface
	orgaoSnapshotRepo OrgaoSnapshotRepositoryInterface
	enabled           bool
	prefrioDomain     string
	taskCh            chan emailTask
	wg                sync.WaitGroup
}

const emailWorkerCount = 5
const emailQueueSize = 200

func NewEmailNotificationService(
	dataRelayClient *clients.DataRelayClient,
	cursoRepo CursoRepositoryInterface,
	orgaoSnapshotRepo OrgaoSnapshotRepositoryInterface,
	enabled bool,
	prefrioDomain string,
) *EmailNotificationService {
	s := &EmailNotificationService{
		dataRelayClient:   dataRelayClient,
		cursoRepo:         cursoRepo,
		orgaoSnapshotRepo: orgaoSnapshotRepo,
		enabled:           enabled,
		prefrioDomain:     prefrioDomain,
		taskCh:            make(chan emailTask, emailQueueSize),
	}

	for i := 0; i < emailWorkerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	return s
}

func (s *EmailNotificationService) worker() {
	defer s.wg.Done()
	for task := range s.taskCh {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in email worker recovered", "panic", r)
				}
			}()
			task.fn()
		}()
	}
}

func (s *EmailNotificationService) Enqueue(fn func()) {
	select {
	case s.taskCh <- emailTask{fn: fn}:
	default:
		logger.Warn("email queue full, dropping task")
	}
}

func (s *EmailNotificationService) Shutdown() {
	close(s.taskCh)
	s.wg.Wait()
}

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

type ScheduleInfo struct {
	ClassTime      string
	ClassStartDate string
	ClassDays      string
	Address        string
}

func (s *EmailNotificationService) getScheduleInfo(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) *ScheduleInfo {
	if inscricao.ScheduleID == nil || s.cursoRepo == nil {
		return nil
	}

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

func (s *EmailNotificationService) SendEnrollmentCreatedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		logger.Debug("email notifications disabled, skipping enrollment created email", "email", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		logger.Debug("no email address for enrollment, skipping", "enrollment_id", inscricao.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, curso)
	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, orgaoName, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{inscricao.Email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment created email: %w", err)
	}

	logger.Info("sent enrollment created email", "email", inscricao.Email, "course", curso.Titulo)
	return nil
}

func (s *EmailNotificationService) SendEnrollmentApprovedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		logger.Debug("email notifications disabled, skipping enrollment approved email", "email", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		logger.Debug("no email address for enrollment, skipping", "enrollment_id", inscricao.ID)
		return nil
	}

	orgaoName := s.getOrgaoName(ctx, curso)
	scheduleInfo := s.getScheduleInfo(ctx, inscricao, curso)
	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, orgaoName, scheduleInfo, s.prefrioDomain)

	emailReq := &clients.EmailRequest{
		ToAddresses: []string{inscricao.Email},
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTMLBody:  template.IsHTML,
	}

	if err := s.dataRelayClient.SendEmail(ctx, emailReq); err != nil {
		return fmt.Errorf("failed to send enrollment approved email: %w", err)
	}

	logger.Info("sent enrollment approved email", "email", inscricao.Email, "course", curso.Titulo)
	return nil
}

func (s *EmailNotificationService) SendEnrollmentRejectedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	if !s.enabled {
		logger.Debug("email notifications disabled, skipping enrollment rejected email", "email", inscricao.Email)
		return nil
	}

	if inscricao.Email == "" {
		logger.Debug("no email address for enrollment, skipping", "enrollment_id", inscricao.ID)
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

	logger.Info("sent enrollment rejected email", "email", inscricao.Email, "course", curso.Titulo)
	return nil
}
