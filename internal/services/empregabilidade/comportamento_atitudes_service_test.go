package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	service "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ==========================================
// MOCK REPOSITORIES
// ==========================================

type MockComportamentoAtitudesRepository struct {
	mock.Mock
}

func (m *MockComportamentoAtitudesRepository) CreateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) (int64, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockComportamentoAtitudesRepository) GetComportamentoAtitudesByID(ctx context.Context, id int64) (*empregabilidade.ComportamentoAtitudes, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.ComportamentoAtitudes), args.Error(1)
}

func (m *MockComportamentoAtitudesRepository) UpdateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockComportamentoAtitudesRepository) DeleteComportamentoAtitudes(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockComportamentoAtitudesRepository) ListComportamentoAtitudes(ctx context.Context, filter empregabilidade.ComportamentoAtitudesFilter, limit, offset int) ([]*empregabilidade.ComportamentoAtitudes, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*empregabilidade.ComportamentoAtitudes), args.Get(1).(int64), args.Error(2)
}

// Validação em tempo de compilação da interface
var _ service.ComportamentoAtitudesRepositoryInterface = (*MockComportamentoAtitudesRepository)(nil)

// ==========================================
// TESTES DO COMPORTAMENTO ATITUDES SERVICE
// ==========================================

func TestNewComportamentoAtitudesService(t *testing.T) {
	mockRepo := new(MockComportamentoAtitudesRepository)
	svc := service.NewComportamentoAtitudesService(mockRepo)
	assert.NotNil(t, svc)
}

func TestComportamentoAtitudesService_CreateComportamentoAtitudes(t *testing.T) {
	ctx := context.Background()
	var comportamentoID int64 = 10
	entity := &empregabilidade.ComportamentoAtitudes{Nome: "Liderança"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)

		mockRepo.On("CreateComportamentoAtitudes", ctx, entity).Return(comportamentoID, nil)

		id, err := svc.CreateComportamentoAtitudes(ctx, entity)

		assert.NoError(t, err)
		assert.Equal(t, comportamentoID, id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)
		expectedErr := errors.New("erro ao criar no banco")

		mockRepo.On("CreateComportamentoAtitudes", ctx, entity).Return(int64(0), expectedErr)

		id, err := svc.CreateComportamentoAtitudes(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestComportamentoAtitudesService_GetComportamentoAtitudesByID(t *testing.T) {
	ctx := context.Background()
	var comportamentoID int64 = 15
	expectedEntity := &empregabilidade.ComportamentoAtitudes{ID: comportamentoID, Nome: "Trabalho em Equipe"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)

		mockRepo.On("GetComportamentoAtitudesByID", ctx, comportamentoID).Return(expectedEntity, nil)

		result, err := svc.GetComportamentoAtitudesByID(ctx, comportamentoID)

		assert.NoError(t, err)
		assert.Equal(t, expectedEntity, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found or Error", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)
		expectedErr := errors.New("comportamento não encontrado")

		mockRepo.On("GetComportamentoAtitudesByID", ctx, comportamentoID).Return(nil, expectedErr)

		result, err := svc.GetComportamentoAtitudesByID(ctx, comportamentoID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestComportamentoAtitudesService_UpdateComportamentoAtitudes(t *testing.T) {
	ctx := context.Background()
	entity := &empregabilidade.ComportamentoAtitudes{ID: 20, Nome: "Proatividade Avançada"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)

		mockRepo.On("UpdateComportamentoAtitudes", ctx, entity).Return(nil)

		err := svc.UpdateComportamentoAtitudes(ctx, entity)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)
		expectedErr := errors.New("falha ao atualizar")

		mockRepo.On("UpdateComportamentoAtitudes", ctx, entity).Return(expectedErr)

		err := svc.UpdateComportamentoAtitudes(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestComportamentoAtitudesService_DeleteComportamentoAtitudes(t *testing.T) {
	ctx := context.Background()
	var comportamentoID int64 = 25

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)

		mockRepo.On("DeleteComportamentoAtitudes", ctx, comportamentoID).Return(nil)

		err := svc.DeleteComportamentoAtitudes(ctx, comportamentoID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)
		expectedErr := errors.New("falha ao remover")

		mockRepo.On("DeleteComportamentoAtitudes", ctx, comportamentoID).Return(expectedErr)

		err := svc.DeleteComportamentoAtitudes(ctx, comportamentoID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestComportamentoAtitudesService_ListComportamentoAtitudes(t *testing.T) {
	ctx := context.Background()
	filter := empregabilidade.ComportamentoAtitudesFilter{Search: "Lid"}
	expectedList := []*empregabilidade.ComportamentoAtitudes{
		{ID: 1, Nome: "Liderança"},
	}

	t.Run("Success with Page Calculation", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)

		page := 2
		pageSize := 10
		expectedOffset := 10

		mockRepo.On("ListComportamentoAtitudes", ctx, filter, pageSize, expectedOffset).
			Return(expectedList, int64(1), nil)

		result, total, err := svc.ListComportamentoAtitudes(ctx, filter, page, pageSize)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockComportamentoAtitudesRepository)
		svc := service.NewComportamentoAtitudesService(mockRepo)
		expectedErr := errors.New("erro ao listar")

		mockRepo.On("ListComportamentoAtitudes", ctx, filter, 10, 0).
			Return(nil, int64(0), expectedErr)

		result, total, err := svc.ListComportamentoAtitudes(ctx, filter, 1, 10)

		assert.Error(t, err)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
