package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDataRelayClient mocks the DataRelayClient for testing
type MockDataRelayClient struct {
	sendEmailFunc func(ctx context.Context, req *clients.EmailRequest) error
	sentEmails    []*clients.EmailRequest
}

func (m *MockDataRelayClient) SendEmail(ctx context.Context, req *clients.EmailRequest) error {
	if m.sendEmailFunc != nil {
		return m.sendEmailFunc(ctx, req)
	}
	m.sentEmails = append(m.sentEmails, req)
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Migrate schemas (skip problematic tables for SQLite)
	err = db.AutoMigrate(
		&models.CitizenSnapshot{},
		&models.OrgaoSnapshot{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestNewEmailNotificationService(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.enabled != true {
		t.Error("Expected enabled to be true")
	}

	if service.prefrioDomain != "oportunidades.rio" {
		t.Errorf("Expected prefrioDomain to be 'oportunidades.rio', got '%s'", service.prefrioDomain)
	}
}

func TestSendEnrollmentCreatedEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "João Silva",
		Email: "joao@test.com",
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:           1,
		Titulo:       "Curso de Go",
		Organization: "SMTR",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	if len(email.ToAddresses) != 1 || email.ToAddresses[0] != "joao@test.com" {
		t.Errorf("Expected email to joao@test.com, got %v", email.ToAddresses)
	}

	if !strings.Contains(email.Subject, "Inscrição recebida!") {
		t.Errorf("Expected subject to contain 'Inscrição recebida!', got: %s", email.Subject)
	}

	if !strings.Contains(email.Body, "João Silva") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(email.Body, "Curso de Go") {
		t.Error("Expected body to contain course title")
	}

	if !email.IsHTMLBody {
		t.Error("Expected HTML email")
	}
}

func TestSendEnrollmentCreatedEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "João Silva",
		Email: "joao@test.com",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Curso de Go",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentCreatedEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "João Silva",
		Email: "", // No email
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Curso de Go",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentCreatedEmail_WithCitizenSnapshot(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	// Create citizen snapshot with updated email
	snapshot := &models.CitizenSnapshot{
		CPF:   "12345678901",
		Nome:  "João Silva",
		Email: "updated@rmi.com",
	}
	db.Create(snapshot)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "João Silva",
		Email: "old@test.com", // Old email
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Curso de Go",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	// Should use the snapshot email, not the inscricao email
	if len(email.ToAddresses) != 1 || email.ToAddresses[0] != "updated@rmi.com" {
		t.Errorf("Expected email to updated@rmi.com from snapshot, got %v", email.ToAddresses)
	}
}

func TestSendEnrollmentCreatedEmail_WithOrgaoSnapshot(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	// Create orgao snapshot
	orgao := &models.OrgaoSnapshot{
		OrgaoID:      "smtr-001",
		Name:         "Secretaria Municipal de Transportes",
		LastSyncedAt: time.Now(),
		SyncStatus:   "synced",
	}
	db.Create(orgao)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "João Silva",
		Email: "joao@test.com",
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:           1,
		Titulo:       "Curso de Go",
		OrgaoID:      "smtr-001",
		Organization: "Old Name",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	// Should use the orgao snapshot name
	if !strings.Contains(email.Body, "Secretaria Municipal de Transportes") {
		t.Error("Expected body to contain orgao snapshot name")
	}
}

func TestSendEnrollmentApprovedEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Maria Santos",
		Email: "maria@test.com",
		CPF:   "98765432100",
	}

	dataInicio := time.Now().Add(7 * 24 * time.Hour)
	curso := &models.Curso{
		ID:              2,
		Titulo:          "Curso Avançado de Go",
		Organization:    "SMTR",
		Modalidade:      models.ModalidadePresencial,
		LocalRealizacao: "Rua das Flores, 123",
		DataInicio:      &dataInicio,
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	if len(email.ToAddresses) != 1 || email.ToAddresses[0] != "maria@test.com" {
		t.Errorf("Expected email to maria@test.com, got %v", email.ToAddresses)
	}

	if !strings.Contains(email.Subject, "Parabéns!") {
		t.Errorf("Expected subject to contain 'Parabéns!', got: %s", email.Subject)
	}

	if !strings.Contains(email.Body, "Maria Santos") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(email.Body, "Curso Avançado de Go") {
		t.Error("Expected body to contain course title")
	}

	if !strings.Contains(email.Body, "Rua das Flores, 123") {
		t.Error("Expected body to contain location")
	}

	if !email.IsHTMLBody {
		t.Error("Expected HTML email")
	}
}

func TestSendEnrollmentApprovedEmail_WithScheduleInfo(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil, // No cursoRepo needed - schedule won't be found
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	scheduleID := uuid.New()
	inscricao := &models.Inscricao{
		ID:         uuid.New(),
		Name:       "Pedro Costa",
		Email:      "pedro@test.com",
		CPF:        "11122233344",
		ScheduleID: &scheduleID,
	}

	dataInicio := time.Now().Add(10 * 24 * time.Hour)
	curso := &models.Curso{
		ID:              3,
		Titulo:          "Curso de Kubernetes",
		Modalidade:      models.ModalidadePresencial,
		LocalRealizacao: "Av. Presidente Vargas, 500",
		DataInicio:      &dataInicio,
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	// Since schedule info won't be found, it should fall back to curso data
	if !strings.Contains(email.Body, "Av. Presidente Vargas, 500") {
		t.Error("Expected body to contain course location")
	}
}

func TestSendEnrollmentApprovedEmail_OnlineCourse(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Ana Paula",
		Email: "ana@test.com",
	}

	curso := &models.Curso{
		ID:         4,
		Titulo:     "Curso Online de Docker",
		Modalidade: models.ModalidadeRemoto,
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	if !strings.Contains(email.Body, "online") {
		t.Error("Expected body to contain 'online' for remote course")
	}
}

func TestSendEnrollmentApprovedEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@test.com",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentApprovedEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "", // No email
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentApprovedEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@test.com",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send enrollment approved email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendEnrollmentRejectedEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:     uuid.New(),
		Name:   "Carlos Oliveira",
		Email:  "carlos@test.com",
		CPF:    "55566677788",
		Reason: "Documentação incompleta",
	}

	curso := &models.Curso{
		ID:     5,
		Titulo: "Curso de Linux",
	}

	ctx := context.Background()
	err := service.SendEnrollmentRejectedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	email := mockClient.sentEmails[0]
	if len(email.ToAddresses) != 1 || email.ToAddresses[0] != "carlos@test.com" {
		t.Errorf("Expected email to carlos@test.com, got %v", email.ToAddresses)
	}

	if !strings.Contains(email.Subject, "Informações sobre sua inscrição") {
		t.Errorf("Expected subject about enrollment info, got: %s", email.Subject)
	}

	if !strings.Contains(email.Body, "Carlos Oliveira") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(email.Body, "Curso de Linux") {
		t.Error("Expected body to contain course title")
	}

	if !strings.Contains(email.Body, "não foi aprovada") {
		t.Error("Expected body to contain rejection message")
	}
}

func TestSendEnrollmentRejectedEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Carlos Oliveira",
		Email: "carlos@test.com",
	}

	curso := &models.Curso{
		ID:     5,
		Titulo: "Curso de Linux",
	}

	ctx := context.Background()
	err := service.SendEnrollmentRejectedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentRejectedEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "", // No email
		CPF:   "12345678901",
	}

	curso := &models.Curso{
		ID:     5,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentRejectedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendEnrollmentRejectedEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@test.com",
	}

	curso := &models.Curso{
		ID:     5,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentRejectedEmail(ctx, inscricao, curso)
	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send enrollment rejected email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@test.com",
	}

	curso := &models.Curso{
		ID:     1,
		Titulo: "Test Course",
	}

	ctx := context.Background()
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send enrollment created email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestResolveEmail_PreferSnapshot(t *testing.T) {
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	// Create citizen snapshot
	snapshot := &models.CitizenSnapshot{
		CPF:   "12345678901",
		Email: "snapshot@rmi.com",
	}
	db.Create(snapshot)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		Email: "old@inscricao.com",
		CPF:   "12345678901",
	}

	ctx := context.Background()
	email := service.resolveEmail(ctx, inscricao)

	if email != "snapshot@rmi.com" {
		t.Errorf("Expected snapshot email, got: %s", email)
	}
}

func TestResolveEmail_FallbackToInscricao(t *testing.T) {
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		Email: "inscricao@test.com",
		CPF:   "99999999999", // No snapshot exists
	}

	ctx := context.Background()
	email := service.resolveEmail(ctx, inscricao)

	if email != "inscricao@test.com" {
		t.Errorf("Expected inscricao email, got: %s", email)
	}
}

func TestGetOrgaoName_PreferSnapshot(t *testing.T) {
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	// Create orgao snapshot
	orgao := &models.OrgaoSnapshot{
		OrgaoID:      "smtr-001",
		Name:         "Secretaria de Transporte (Snapshot)",
		LastSyncedAt: time.Now(),
		SyncStatus:   "synced",
	}
	db.Create(orgao)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	curso := &models.Curso{
		OrgaoID:      "smtr-001",
		Organization: "Old Organization Name",
	}

	ctx := context.Background()
	name := service.getOrgaoName(ctx, curso)

	if name != "Secretaria de Transporte (Snapshot)" {
		t.Errorf("Expected snapshot name, got: %s", name)
	}
}

func TestGetOrgaoName_FallbackToOrganization(t *testing.T) {
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	curso := &models.Curso{
		OrgaoID:      "nonexistent",
		Organization: "Fallback Organization",
	}

	ctx := context.Background()
	name := service.getOrgaoName(ctx, curso)

	if name != "Fallback Organization" {
		t.Errorf("Expected fallback organization, got: %s", name)
	}
}

func TestGetOrgaoName_DefaultFallback(t *testing.T) {
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	curso := &models.Curso{
		OrgaoID:      "",
		Organization: "",
	}

	ctx := context.Background()
	name := service.getOrgaoName(ctx, curso)

	if name != "órgão responsável" {
		t.Errorf("Expected default fallback, got: %s", name)
	}
}

func TestGetScheduleInfo_CourseSchedule(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	t.Run("valid course schedule with location", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		// Migrate the necessary schemas
		err := db.AutoMigrate(
			&models.Curso{},
			&models.LocationClass{},
			&models.CourseSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     1,
			Titulo: "Test Course",
		}
		db.Create(curso)

		// Create a location
		location := &models.LocationClass{
			ID:           uuid.New(),
			CursoID:      curso.ID,
			Address:      "Rua das Flores, 123",
			Neighborhood: "Centro",
		}
		db.Create(location)

		// Create a course schedule
		classStartDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
		classEndDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
		schedule := &models.CourseSchedule{
			ID:             uuid.New(),
			LocationID:     location.ID,
			Vacancies:      20,
			ClassStartDate: classStartDate,
			ClassEndDate:   classEndDate,
			ClassTime:      "14:00 - 18:00",
			ClassDays:      "Segunda, Quarta, Sexta",
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		if info.ClassTime != "14:00 - 18:00" {
			t.Errorf("Expected ClassTime '14:00 - 18:00', got '%s'", info.ClassTime)
		}

		if info.ClassStartDate != "15/04/2026" {
			t.Errorf("Expected ClassStartDate '15/04/2026', got '%s'", info.ClassStartDate)
		}

		if info.ClassDays != "Segunda, Quarta, Sexta" {
			t.Errorf("Expected ClassDays 'Segunda, Quarta, Sexta', got '%s'", info.ClassDays)
		}

		if info.Address != "Rua das Flores, 123" {
			t.Errorf("Expected Address 'Rua das Flores, 123', got '%s'", info.Address)
		}
	})

	t.Run("course schedule without location", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.LocationClass{},
			&models.CourseSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     2,
			Titulo: "Test Course 2",
		}
		db.Create(curso)

		// Create a location (but we won't associate it in the retrieval)
		location := &models.LocationClass{
			ID:           uuid.New(),
			CursoID:      curso.ID,
			Address:      "Av. Brasil, 456",
			Neighborhood: "Zona Sul",
		}
		db.Create(location)

		// Create a course schedule
		classStartDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		classEndDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		schedule := &models.CourseSchedule{
			ID:             uuid.New(),
			LocationID:     location.ID,
			Vacancies:      15,
			ClassStartDate: classStartDate,
			ClassEndDate:   classEndDate,
			ClassTime:      "09:00 - 12:00",
			ClassDays:      "Terça, Quinta",
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		if info.ClassTime != "09:00 - 12:00" {
			t.Errorf("Expected ClassTime '09:00 - 12:00', got '%s'", info.ClassTime)
		}

		if info.ClassStartDate != "01/05/2026" {
			t.Errorf("Expected ClassStartDate '01/05/2026', got '%s'", info.ClassStartDate)
		}

		if info.ClassDays != "Terça, Quinta" {
			t.Errorf("Expected ClassDays 'Terça, Quinta', got '%s'", info.ClassDays)
		}

		// Address should still be populated if location exists
		if info.Address != "Av. Brasil, 456" {
			t.Errorf("Expected Address 'Av. Brasil, 456', got '%s'", info.Address)
		}
	})

	t.Run("nil repository", func(t *testing.T) {
		service := NewEmailNotificationService(
			nil,
			nil, // No repo
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		scheduleID := uuid.New()
		inscricao := &models.Inscricao{
			ScheduleID: &scheduleID,
		}

		curso := &models.Curso{}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info != nil {
			t.Errorf("Expected nil schedule info when repo is nil, got: %v", info)
		}
	})
}

func TestGetScheduleInfo_RemoteSchedule(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	t.Run("valid remote schedule with all fields", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.RemoteClass{},
			&models.RemoteSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     3,
			Titulo: "Remote Course",
		}
		db.Create(curso)

		// Create a remote class
		remoteClass := &models.RemoteClass{
			ID:      uuid.New(),
			CursoID: curso.ID,
		}
		db.Create(remoteClass)

		// Create a remote schedule with all optional fields
		classStartDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		classEndDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		classTime := "19:00 - 21:00"
		classDays := "Segunda a Sexta"

		schedule := &models.RemoteSchedule{
			ID:             uuid.New(),
			RemoteClassID:  remoteClass.ID,
			Vacancies:      50,
			ClassStartDate: &classStartDate,
			ClassEndDate:   &classEndDate,
			ClassTime:      &classTime,
			ClassDays:      &classDays,
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		if info.Address != "online" {
			t.Errorf("Expected Address 'online' for remote schedule, got '%s'", info.Address)
		}

		if info.ClassTime != "19:00 - 21:00" {
			t.Errorf("Expected ClassTime '19:00 - 21:00', got '%s'", info.ClassTime)
		}

		if info.ClassStartDate != "01/06/2026" {
			t.Errorf("Expected ClassStartDate '01/06/2026', got '%s'", info.ClassStartDate)
		}

		if info.ClassDays != "Segunda a Sexta" {
			t.Errorf("Expected ClassDays 'Segunda a Sexta', got '%s'", info.ClassDays)
		}
	})

	t.Run("remote schedule with nil optional fields", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.RemoteClass{},
			&models.RemoteSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     4,
			Titulo: "Remote Course Minimal",
		}
		db.Create(curso)

		// Create a remote class
		remoteClass := &models.RemoteClass{
			ID:      uuid.New(),
			CursoID: curso.ID,
		}
		db.Create(remoteClass)

		// Create a remote schedule with nil optional fields
		schedule := &models.RemoteSchedule{
			ID:             uuid.New(),
			RemoteClassID:  remoteClass.ID,
			Vacancies:      30,
			ClassStartDate: nil,
			ClassEndDate:   nil,
			ClassTime:      nil,
			ClassDays:      nil,
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		if info.Address != "online" {
			t.Errorf("Expected Address 'online' for remote schedule, got '%s'", info.Address)
		}

		// Optional fields should be empty strings
		if info.ClassTime != "" {
			t.Errorf("Expected empty ClassTime, got '%s'", info.ClassTime)
		}

		if info.ClassStartDate != "" {
			t.Errorf("Expected empty ClassStartDate, got '%s'", info.ClassStartDate)
		}

		if info.ClassDays != "" {
			t.Errorf("Expected empty ClassDays, got '%s'", info.ClassDays)
		}
	})

	t.Run("remote schedule with partial fields", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.RemoteClass{},
			&models.RemoteSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     5,
			Titulo: "Remote Course Partial",
		}
		db.Create(curso)

		// Create a remote class
		remoteClass := &models.RemoteClass{
			ID:      uuid.New(),
			CursoID: curso.ID,
		}
		db.Create(remoteClass)

		// Create a remote schedule with some fields populated
		classStartDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		classTime := "10:00 - 13:00"

		schedule := &models.RemoteSchedule{
			ID:             uuid.New(),
			RemoteClassID:  remoteClass.ID,
			Vacancies:      25,
			ClassStartDate: &classStartDate,
			ClassEndDate:   nil,
			ClassTime:      &classTime,
			ClassDays:      nil, // No class days
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		if info.Address != "online" {
			t.Errorf("Expected Address 'online', got '%s'", info.Address)
		}

		if info.ClassTime != "10:00 - 13:00" {
			t.Errorf("Expected ClassTime '10:00 - 13:00', got '%s'", info.ClassTime)
		}

		if info.ClassStartDate != "15/07/2026" {
			t.Errorf("Expected ClassStartDate '15/07/2026', got '%s'", info.ClassStartDate)
		}

		if info.ClassDays != "" {
			t.Errorf("Expected empty ClassDays, got '%s'", info.ClassDays)
		}
	})
}

func TestGetScheduleInfo_NoScheduleID(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)

	service := NewEmailNotificationService(
		nil,
		cursoRepo,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ScheduleID: nil,
	}

	curso := &models.Curso{}

	ctx := context.Background()
	info := service.getScheduleInfo(ctx, inscricao, curso)

	if info != nil {
		t.Errorf("Expected nil schedule info when no schedule ID, got: %v", info)
	}
}

func TestGetScheduleInfo_ScheduleNotFound(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	t.Run("course schedule not found", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.LocationClass{},
			&models.CourseSchedule{},
			&models.RemoteClass{},
			&models.RemoteSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Use a non-existent schedule ID
		nonExistentID := uuid.New()
		inscricao := &models.Inscricao{
			ScheduleID: &nonExistentID,
		}

		curso := &models.Curso{}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		// Should return nil when schedule is not found
		if info != nil {
			t.Errorf("Expected nil schedule info when schedule not found, got: %v", info)
		}
	})
}

func TestGetScheduleInfo_DateFormatting(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	t.Run("various date formats", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.LocationClass{},
			&models.CourseSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     6,
			Titulo: "Date Format Test",
		}
		db.Create(curso)

		// Create a location
		location := &models.LocationClass{
			ID:           uuid.New(),
			CursoID:      curso.ID,
			Address:      "Test Address",
			Neighborhood: "Test Neighborhood",
		}
		db.Create(location)

		// Test various dates
		testCases := []struct {
			name           string
			date           time.Time
			expectedFormat string
		}{
			{
				name:           "single digit day and month",
				date:           time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
				expectedFormat: "05/01/2026",
			},
			{
				name:           "double digit day and month",
				date:           time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
				expectedFormat: "25/12/2026",
			},
			{
				name:           "end of year",
				date:           time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
				expectedFormat: "31/12/2026",
			},
			{
				name:           "beginning of year",
				date:           time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
				expectedFormat: "01/01/2027",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schedule := &models.CourseSchedule{
					ID:             uuid.New(),
					LocationID:     location.ID,
					Vacancies:      10,
					ClassStartDate: tc.date,
					ClassEndDate:   tc.date.Add(30 * 24 * time.Hour),
					ClassTime:      "10:00 - 12:00",
					ClassDays:      "Todos os dias",
				}
				db.Create(schedule)

				inscricao := &models.Inscricao{
					ScheduleID: &schedule.ID,
				}

				ctx := context.Background()
				info := service.getScheduleInfo(ctx, inscricao, curso)

				if info == nil {
					t.Fatal("Expected schedule info, got nil")
				}

				if info.ClassStartDate != tc.expectedFormat {
					t.Errorf("Expected date format '%s', got '%s'", tc.expectedFormat, info.ClassStartDate)
				}
			})
		}
	})
}

func TestGetScheduleInfo_FallbackBehavior(t *testing.T) {
	t.Skip("Skipping due to SQLite UUID compatibility issues")
	t.Run("fallback from course schedule to remote schedule", func(t *testing.T) {
		db := setupTestDB(t)
		cursoRepo := repository.NewCursoRepository(db)

		err := db.AutoMigrate(
			&models.Curso{},
			&models.RemoteClass{},
			&models.RemoteSchedule{},
		)
		if err != nil {
			t.Fatalf("failed to migrate schemas: %v", err)
		}

		service := NewEmailNotificationService(
			nil,
			cursoRepo,
			nil,
			nil,
			true,
			"oportunidades.rio",
		)

		// Create a course
		curso := &models.Curso{
			ID:     7,
			Titulo: "Fallback Test",
		}
		db.Create(curso)

		// Create only a remote class (no course schedule)
		remoteClass := &models.RemoteClass{
			ID:      uuid.New(),
			CursoID: curso.ID,
		}
		db.Create(remoteClass)

		classTime := "16:00 - 18:00"
		schedule := &models.RemoteSchedule{
			ID:            uuid.New(),
			RemoteClassID: remoteClass.ID,
			Vacancies:     40,
			ClassTime:     &classTime,
		}
		db.Create(schedule)

		inscricao := &models.Inscricao{
			ScheduleID: &schedule.ID,
		}

		ctx := context.Background()
		info := service.getScheduleInfo(ctx, inscricao, curso)

		if info == nil {
			t.Fatal("Expected schedule info, got nil")
		}

		// Should get remote schedule since course schedule doesn't exist
		if info.Address != "online" {
			t.Errorf("Expected 'online' address for remote schedule, got '%s'", info.Address)
		}

		if info.ClassTime != "16:00 - 18:00" {
			t.Errorf("Expected ClassTime '16:00 - 18:00', got '%s'", info.ClassTime)
		}
	})
}

func TestSendEnrollmentEmails_Integration(t *testing.T) {
	// Integration test: sending all three types of emails for a complete workflow
	mockClient := &MockDataRelayClient{}
	db := setupTestDB(t)
	cursoRepo := repository.NewCursoRepository(db)
	orgaoRepo := repository.NewOrgaoSnapshotRepository(db)
	citizenRepo := repository.NewCitizenSnapshotRepository(db)

	service := NewEmailNotificationService(
		mockClient,
		cursoRepo,
		orgaoRepo,
		citizenRepo,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Integration Test User",
		Email: "integration@test.com",
		CPF:   "12312312312",
	}

	curso := &models.Curso{
		ID:           100,
		Titulo:       "Integration Test Course",
		Organization: "Test Org",
		Modalidade:   models.ModalidadePresencial,
	}

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()

	// 1. Created email
	err := service.SendEnrollmentCreatedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Failed to send created email: %v", err)
	}

	// 2. Approved email
	err = service.SendEnrollmentApprovedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Failed to send approved email: %v", err)
	}

	// 3. Rejected email
	err = service.SendEnrollmentRejectedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Failed to send rejected email: %v", err)
	}

	// 4. Schedule change email
	err = service.SendScheduleChangedEmail(ctx, inscricao, curso)
	if err != nil {
		t.Fatalf("Failed to send schedule change email: %v", err)
	}

	// 5. Received application
	err = service.SendCandidaturaEnviadaEmail(ctx, candidatura)
	if err != nil {
		t.Fatalf("Failed to send received application email: %v", err)
	}

	// 6. Approved application
	err = service.SendCandidaturaAprovadaEmail(ctx, candidatura)
	if err != nil {
		t.Fatalf("Failed to send approved application email: %v", err)
	}

	// 7. Failed application
	err = service.SendCandidaturaReprovadaEmail(ctx, candidatura)
	if err != nil {
		t.Fatalf("Failed to send failed application email: %v", err)
	}

	// Verify all seven emails were sent
	if len(mockClient.sentEmails) != 7 {
		t.Fatalf("Expected 7 emails in workflow, got %d", len(mockClient.sentEmails))
	}

	// Verify each email is different
	subjects := make(map[string]bool)
	for _, email := range mockClient.sentEmails {
		subjects[email.Subject] = true
	}

	if len(subjects) != 7 {
		t.Error("Expected 7 different email subjects")
	}
}

func TestSendScheduleChangedEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Usuário Teste",
		Email: "test@example.com",
		CPF:   "12312312312",
	}

	curso := &models.Curso{
		ID:           100,
		Titulo:       "Curso Teste",
		Organization: "Org Teste",
		Modalidade:   models.ModalidadePresencial,
	}

	ctx := context.Background()
	err := service.SendScheduleChangedEmail(ctx, inscricao, curso)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	sentEmail := mockClient.sentEmails[0]

	if len(sentEmail.ToAddresses) != 1 || sentEmail.ToAddresses[0] != "test@example.com" {
		t.Errorf("Expected email to test@example.com, got %v", sentEmail.ToAddresses)
	}

	if !strings.Contains(sentEmail.Subject, "Troca de turma realizada com sucesso") {
		t.Errorf("Expected subject about schedule change, got: %s", sentEmail.Subject)
	}

	if !strings.Contains(sentEmail.Body, "Usuário Teste") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(sentEmail.Body, "Curso Teste") {
		t.Error("Expected body to contain course title")
	}
}

func TestSendScheduleChangedEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Usuário Teste",
		Email: "",
		CPF:   "12312312312",
	}

	curso := &models.Curso{
		ID:           100,
		Titulo:       "Curso Teste",
		Organization: "Org Teste",
		Modalidade:   models.ModalidadePresencial,
	}

	ctx := context.Background()
	err := service.SendScheduleChangedEmail(ctx, inscricao, curso)

	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendScheduleChangedEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Usuário Teste",
		Email: "test@example.com",
		CPF:   "12312312312",
	}

	curso := &models.Curso{
		ID:           100,
		Titulo:       "Curso Teste",
		Organization: "Org Teste",
		Modalidade:   models.ModalidadePresencial,
	}

	ctx := context.Background()
	err := service.SendScheduleChangedEmail(ctx, inscricao, curso)

	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send schedule changed email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendScheduleChangedEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	inscricao := &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Usuário Teste",
		Email: "test@example.com",
		CPF:   "12312312312",
	}

	curso := &models.Curso{
		ID:           100,
		Titulo:       "Curso Teste",
		Organization: "Org Teste",
		Modalidade:   models.ModalidadePresencial,
	}

	ctx := context.Background()
	err := service.SendScheduleChangedEmail(ctx, inscricao, curso)

	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaEnviadaEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaEnviadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	sentEmail := mockClient.sentEmails[0]

	if len(sentEmail.ToAddresses) != 1 || sentEmail.ToAddresses[0] != "test@example.com" {
		t.Errorf("Expected email to test@example.com, got %v", sentEmail.ToAddresses)
	}

	if !strings.Contains(sentEmail.Subject, "Candidatura recebida!") {
		t.Errorf("Expected subject about received application, got: %s", sentEmail.Subject)
	}

	if !strings.Contains(sentEmail.Body, "Usuário Teste") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(sentEmail.Body, "Vaga Teste") {
		t.Error("Expected body to contain position title")
	}
}

func TestSendCandidaturaEnviadaEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := ""

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaEnviadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaEnviadaEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaEnviadaEmail(ctx, candidatura)

	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send received application email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendCandidaturaEnviadaEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaEnviadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaAprovadaEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaAprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	sentEmail := mockClient.sentEmails[0]

	if len(sentEmail.ToAddresses) != 1 || sentEmail.ToAddresses[0] != "test@example.com" {
		t.Errorf("Expected email to test@example.com, got %v", sentEmail.ToAddresses)
	}

	if !strings.Contains(sentEmail.Subject, "Parabéns! 🎉Candidatura aprovada") {
		t.Errorf("Expected subject about received application, got: %s", sentEmail.Subject)
	}

	if !strings.Contains(sentEmail.Body, "Usuário Teste") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(sentEmail.Body, "Vaga Teste") {
		t.Error("Expected body to contain position title")
	}
}

func TestSendCandidaturaAprovadaEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := ""

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaAprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaAprovadaEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaAprovadaEmail(ctx, candidatura)

	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send approved application email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendCandidaturaAprovadaEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaAprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaReprovadaEmail_Success(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaReprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mockClient.sentEmails) != 1 {
		t.Fatalf("Expected 1 email to be sent, got %d", len(mockClient.sentEmails))
	}

	sentEmail := mockClient.sentEmails[0]

	if len(sentEmail.ToAddresses) != 1 || sentEmail.ToAddresses[0] != "test@example.com" {
		t.Errorf("Expected email to test@example.com, got %v", sentEmail.ToAddresses)
	}

	if !strings.Contains(sentEmail.Subject, "Informações sobre sua candidatura") {
		t.Errorf("Expected subject about received application, got: %s", sentEmail.Subject)
	}

	if !strings.Contains(sentEmail.Body, "Usuário Teste") {
		t.Error("Expected body to contain user name")
	}

	if !strings.Contains(sentEmail.Body, "Vaga Teste") {
		t.Error("Expected body to contain position title")
	}
}

func TestSendCandidaturaReprovadaEmail_NoEmail(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := ""

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaReprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error with missing email, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails with missing email address, got %d", len(mockClient.sentEmails))
	}
}

func TestSendCandidaturaReprovadaEmail_DataRelayError(t *testing.T) {
	expectedErr := errors.New("data relay connection failed")
	mockClient := &MockDataRelayClient{
		sendEmailFunc: func(ctx context.Context, req *clients.EmailRequest) error {
			return expectedErr
		},
	}

	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		true,
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaReprovadaEmail(ctx, candidatura)

	if err == nil {
		t.Fatal("Expected error when DataRelay fails")
	}

	if !strings.Contains(err.Error(), "failed to send failed application email") {
		t.Errorf("Expected proper error wrapping, got: %v", err)
	}
}

func TestSendCandidaturaReprovadaEmail_Disabled(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(
		mockClient,
		nil,
		nil,
		nil,
		false, // disabled
		"oportunidades.rio",
	)

	nome := "Usuário Teste"
	email := "test@example.com"

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	vaga := &empregabilidade.Vaga{
		ID:          uuid.New(),
		Titulo:      "Vaga Teste",
		Contratante: empresa,
		OrgaoParceiro: &models.OrgaoSnapshot{
			Name:    "",
			OrgaoID: uuid.New().String(),
		},
	}

	candidatura := &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "55566677788",
		Vaga:  vaga,
	}

	ctx := context.Background()
	err := service.SendCandidaturaReprovadaEmail(ctx, candidatura)

	if err != nil {
		t.Fatalf("Expected no error when disabled, got: %v", err)
	}

	if len(mockClient.sentEmails) != 0 {
		t.Errorf("Expected no emails when disabled, got %d", len(mockClient.sentEmails))
	}
}
