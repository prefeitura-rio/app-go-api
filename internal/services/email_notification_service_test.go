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
	// Test when cursoRepo is nil or schedule doesn't exist
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

	// When repo is nil, should return nil
	if info != nil {
		t.Errorf("Expected nil schedule info when repo is nil, got: %v", info)
	}
}

func TestGetScheduleInfo_RemoteSchedule(t *testing.T) {
	// Test when schedule can't be found
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

	// When repo is nil, should return nil
	if info != nil {
		t.Errorf("Expected nil schedule info when repo is nil, got: %v", info)
	}
}

func TestGetScheduleInfo_NoScheduleID(t *testing.T) {
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

	// Verify all three emails were sent
	if len(mockClient.sentEmails) != 3 {
		t.Fatalf("Expected 3 emails in workflow, got %d", len(mockClient.sentEmails))
	}

	// Verify each email is different
	subjects := make(map[string]bool)
	for _, email := range mockClient.sentEmails {
		subjects[email.Subject] = true
	}

	if len(subjects) != 3 {
		t.Error("Expected 3 different email subjects")
	}
}
