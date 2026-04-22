package empregabilidade

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockDisponibilidadeRepository struct {
	mock.Mock
}

func (m *MockDisponibilidadeRepository) Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockDisponibilidadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.Disponibilidade), args.Error(1)
}

func (m *MockDisponibilidadeRepository) Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockDisponibilidadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDisponibilidadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*empregabilidade.Disponibilidade), args.Int(1), args.Error(2)
}

type MockEtapaRepository struct {
	mock.Mock
}

func (m *MockEtapaRepository) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockEtapaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.Etapa), args.Error(1)
}

func (m *MockEtapaRepository) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockEtapaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEtapaRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	args := m.Called(ctx, vagaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*empregabilidade.Etapa), args.Error(1)
}

func (m *MockEtapaRepository) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	args := m.Called(ctx, vagaID)
	return args.Error(0)
}

type MockOnboardingRepositoryImpl struct {
	mock.Mock
}

func (m *MockOnboardingRepositoryImpl) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error) {
	args := m.Called(ctx, cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.Onboarding), args.Error(1)
}

func (m *MockOnboardingRepositoryImpl) Upsert(ctx context.Context, entity *empregabilidade.Onboarding) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockOnboardingRepositoryImpl) MarkFirstLoginCompleted(ctx context.Context, cpf string) error {
	args := m.Called(ctx, cpf)
	return args.Error(0)
}

type MockTermosUsoRepositoryImpl struct {
	mock.Mock
}

func (m *MockTermosUsoRepositoryImpl) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error) {
	args := m.Called(ctx, cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.TermosUso), args.Error(1)
}

func (m *MockTermosUsoRepositoryImpl) Upsert(ctx context.Context, entity *empregabilidade.TermosUso) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTermosUsoRepositoryImpl) AcceptTerms(ctx context.Context, cpf string) error {
	args := m.Called(ctx, cpf)
	return args.Error(0)
}

// DisponibilidadeService Tests
func TestDisponibilidadeService_Create(t *testing.T) {
	id := uuid.New()
	entity := &empregabilidade.Disponibilidade{
		Descricao: "Integral",
	}

	mockRepo := new(MockDisponibilidadeRepository)
	mockRepo.On("Create", mock.Anything, entity).Return(id, nil)

	service := NewDisponibilidadeServiceWithInterface(mockRepo)
	resultID, err := service.Create(context.Background(), entity)

	assert.NoError(t, err)
	assert.Equal(t, id, resultID)
}

func TestDisponibilidadeService_GetByID(t *testing.T) {
	id := uuid.New()
	entity := &empregabilidade.Disponibilidade{
		ID:        id,
		Descricao: "Integral",
	}

	mockRepo := new(MockDisponibilidadeRepository)
	mockRepo.On("GetByID", mock.Anything, id).Return(entity, nil)

	service := NewDisponibilidadeServiceWithInterface(mockRepo)
	result, err := service.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, entity, result)
}

func TestDisponibilidadeService_Update(t *testing.T) {
	entity := &empregabilidade.Disponibilidade{
		ID:        uuid.New(),
		Descricao: "Meio Período",
	}

	mockRepo := new(MockDisponibilidadeRepository)
	mockRepo.On("Update", mock.Anything, entity).Return(nil)

	service := NewDisponibilidadeServiceWithInterface(mockRepo)
	err := service.Update(context.Background(), entity)

	assert.NoError(t, err)
}

func TestDisponibilidadeService_Delete(t *testing.T) {
	id := uuid.New()

	mockRepo := new(MockDisponibilidadeRepository)
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	service := NewDisponibilidadeServiceWithInterface(mockRepo)
	err := service.Delete(context.Background(), id)

	assert.NoError(t, err)
}

func TestDisponibilidadeService_List(t *testing.T) {
	entities := []*empregabilidade.Disponibilidade{
		{ID: uuid.New(), Descricao: "Integral"},
		{ID: uuid.New(), Descricao: "Meio Período"},
	}

	mockRepo := new(MockDisponibilidadeRepository)
	mockRepo.On("List", mock.Anything, mock.Anything, 10, 0).Return(entities, 2, nil)

	service := NewDisponibilidadeServiceWithInterface(mockRepo)
	results, total, err := service.List(context.Background(), nil, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, entities, results)
	assert.Equal(t, 2, total)
}

// EtapaService Tests
func TestEtapaService_Create(t *testing.T) {
	id := uuid.New()
	vagaID := uuid.New()
	entity := &empregabilidade.Etapa{
		IDVaga: vagaID,
		Titulo: "Entrevista",
		Ordem:  1,
	}

	mockRepo := new(MockEtapaRepository)
	mockRepo.On("Create", mock.Anything, entity).Return(id, nil)

	service := NewEtapaServiceWithInterface(mockRepo)
	resultID, err := service.Create(context.Background(), entity)

	assert.NoError(t, err)
	assert.Equal(t, id, resultID)
}

func TestEtapaService_GetByID(t *testing.T) {
	id := uuid.New()
	entity := &empregabilidade.Etapa{
		ID:     id,
		Titulo: "Teste Técnico",
		Ordem:  2,
	}

	mockRepo := new(MockEtapaRepository)
	mockRepo.On("GetByID", mock.Anything, id).Return(entity, nil)

	service := NewEtapaServiceWithInterface(mockRepo)
	result, err := service.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, entity, result)
}

func TestEtapaService_ListByVaga(t *testing.T) {
	vagaID := uuid.New()
	etapas := []*empregabilidade.Etapa{
		{ID: uuid.New(), IDVaga: vagaID, Titulo: "Entrevista", Ordem: 1},
		{ID: uuid.New(), IDVaga: vagaID, Titulo: "Teste", Ordem: 2},
	}

	mockRepo := new(MockEtapaRepository)
	mockRepo.On("ListByVaga", mock.Anything, vagaID).Return(etapas, nil)

	service := NewEtapaServiceWithInterface(mockRepo)
	results, err := service.ListByVaga(context.Background(), vagaID)

	assert.NoError(t, err)
	assert.Equal(t, etapas, results)
}

func TestEtapaService_DeleteByVaga(t *testing.T) {
	vagaID := uuid.New()

	mockRepo := new(MockEtapaRepository)
	mockRepo.On("DeleteByVaga", mock.Anything, vagaID).Return(nil)

	service := NewEtapaServiceWithInterface(mockRepo)
	err := service.DeleteByVaga(context.Background(), vagaID)

	assert.NoError(t, err)
}

func TestEtapaService_Delete(t *testing.T) {
	id := uuid.New()

	mockRepo := new(MockEtapaRepository)
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	service := NewEtapaServiceWithInterface(mockRepo)
	err := service.Delete(context.Background(), id)

	assert.NoError(t, err)
}

// OnboardingService Tests
func TestOnboardingService_GetByCPF(t *testing.T) {
	cpf := "12345678901"
	onboarding := &empregabilidade.Onboarding{
		CPF:                         cpf,
		IsEmpregabilidadeFirstLogin: true,
	}

	mockRepo := new(MockOnboardingRepositoryImpl)
	mockRepo.On("GetByCPF", mock.Anything, cpf).Return(onboarding, nil)

	service := NewOnboardingServiceWithInterface(mockRepo)
	result, err := service.GetByCPF(context.Background(), cpf)

	assert.NoError(t, err)
	assert.Equal(t, onboarding, result)
}

func TestOnboardingService_Upsert(t *testing.T) {
	onboarding := &empregabilidade.Onboarding{
		CPF:                         "12345678901",
		IsEmpregabilidadeFirstLogin: false,
	}

	mockRepo := new(MockOnboardingRepositoryImpl)
	mockRepo.On("Upsert", mock.Anything, onboarding).Return(nil)

	service := NewOnboardingServiceWithInterface(mockRepo)
	err := service.Upsert(context.Background(), onboarding)

	assert.NoError(t, err)
}

func TestOnboardingService_MarkFirstLoginCompleted(t *testing.T) {
	cpf := "12345678901"

	mockRepo := new(MockOnboardingRepositoryImpl)
	mockRepo.On("MarkFirstLoginCompleted", mock.Anything, cpf).Return(nil)

	service := NewOnboardingServiceWithInterface(mockRepo)
	err := service.MarkFirstLoginCompleted(context.Background(), cpf)

	assert.NoError(t, err)
}

func TestOnboardingService_IsFirstLogin(t *testing.T) {
	tests := []struct {
		name           string
		cpf            string
		onboarding     *empregabilidade.Onboarding
		repoErr        error
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "returns error when GetByCPF fails",
			cpf:            "12345678901",
			onboarding:     nil,
			repoErr:        errors.New("not found"),
			expectedResult: false,
			expectedError:  errors.New("not found"),
		},
		{
			name: "returns true when IsEmpregabilidadeFirstLogin is true",
			cpf:  "12345678901",
			onboarding: &empregabilidade.Onboarding{
				IsEmpregabilidadeFirstLogin: true,
			},
			repoErr:        nil,
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "returns false when IsEmpregabilidadeFirstLogin is false",
			cpf:  "12345678901",
			onboarding: &empregabilidade.Onboarding{
				IsEmpregabilidadeFirstLogin: false,
			},
			repoErr:        nil,
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockOnboardingRepositoryImpl)
			mockRepo.On("GetByCPF", mock.Anything, tt.cpf).Return(tt.onboarding, tt.repoErr)

			service := NewOnboardingServiceWithInterface(mockRepo)
			result, err := service.IsFirstLogin(context.Background(), tt.cpf)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

// TermosUsoService Tests
func TestTermosUsoService_GetByCPF(t *testing.T) {
	cpf := "12345678901"
	termos := &empregabilidade.TermosUso{
		CPF:         cpf,
		UserConsent: true,
	}

	mockRepo := new(MockTermosUsoRepositoryImpl)
	mockRepo.On("GetByCPF", mock.Anything, cpf).Return(termos, nil)

	service := NewTermosUsoServiceWithInterface(mockRepo)
	result, err := service.GetByCPF(context.Background(), cpf)

	assert.NoError(t, err)
	assert.Equal(t, termos, result)
}

func TestTermosUsoService_Upsert(t *testing.T) {
	termos := &empregabilidade.TermosUso{
		CPF:         "12345678901",
		UserConsent: true,
	}

	mockRepo := new(MockTermosUsoRepositoryImpl)
	mockRepo.On("Upsert", mock.Anything, termos).Return(nil)

	service := NewTermosUsoServiceWithInterface(mockRepo)
	err := service.Upsert(context.Background(), termos)

	assert.NoError(t, err)
}

func TestTermosUsoService_AcceptTerms(t *testing.T) {
	cpf := "12345678901"

	mockRepo := new(MockTermosUsoRepositoryImpl)
	mockRepo.On("AcceptTerms", mock.Anything, cpf).Return(nil)

	service := NewTermosUsoServiceWithInterface(mockRepo)
	err := service.AcceptTerms(context.Background(), cpf)

	assert.NoError(t, err)
}

func TestTermosUsoService_HasAcceptedTerms(t *testing.T) {
	tests := []struct {
		name           string
		cpf            string
		termos         *empregabilidade.TermosUso
		repoErr        error
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "returns error when GetByCPF fails",
			cpf:            "12345678901",
			termos:         nil,
			repoErr:        errors.New("not found"),
			expectedResult: false,
			expectedError:  errors.New("not found"),
		},
		{
			name: "returns true when UserConsent is true",
			cpf:  "12345678901",
			termos: &empregabilidade.TermosUso{
				UserConsent: true,
			},
			repoErr:        nil,
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "returns false when UserConsent is false",
			cpf:  "12345678901",
			termos: &empregabilidade.TermosUso{
				UserConsent: false,
			},
			repoErr:        nil,
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTermosUsoRepositoryImpl)
			mockRepo.On("GetByCPF", mock.Anything, tt.cpf).Return(tt.termos, tt.repoErr)

			service := NewTermosUsoServiceWithInterface(mockRepo)
			result, err := service.HasAcceptedTerms(context.Background(), tt.cpf)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

// Error handling tests
func TestLookupServices_ErrorHandling(t *testing.T) {
	t.Run("DisponibilidadeService handles repository errors", func(t *testing.T) {
		mockRepo := new(MockDisponibilidadeRepository)
		mockRepo.On("List", mock.Anything, mock.Anything, 10, 0).
			Return(nil, 0, errors.New("database error"))

		service := NewDisponibilidadeServiceWithInterface(mockRepo)
		results, total, err := service.List(context.Background(), nil, 1, 10)

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, 0, total)
	})

	t.Run("EtapaService handles ListByVaga errors", func(t *testing.T) {
		vagaID := uuid.New()
		mockRepo := new(MockEtapaRepository)
		mockRepo.On("ListByVaga", mock.Anything, vagaID).
			Return(nil, errors.New("database error"))

		service := NewEtapaServiceWithInterface(mockRepo)
		results, err := service.ListByVaga(context.Background(), vagaID)

		assert.Error(t, err)
		assert.Nil(t, results)
	})

	t.Run("OnboardingService propagates errors from GetByCPF in IsFirstLogin", func(t *testing.T) {
		mockRepo := new(MockOnboardingRepositoryImpl)
		mockRepo.On("GetByCPF", mock.Anything, "12345678901").
			Return(nil, errors.New("database error"))

		service := NewOnboardingServiceWithInterface(mockRepo)
		_, err := service.IsFirstLogin(context.Background(), "12345678901")

		assert.Error(t, err)
	})
}
