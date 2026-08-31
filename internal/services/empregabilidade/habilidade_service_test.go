package empregabilidade

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// --- MOCK REPOSITORY ---

type MockHabilidadeRepository struct {
	mock.Mock
}

func (m *MockHabilidadeRepository) CreateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) (int64, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockHabilidadeRepository) GetHabilidadeByID(ctx context.Context, id int64) (*empregabilidade.Habilidade, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.Habilidade), args.Error(1)
}

func (m *MockHabilidadeRepository) UpdateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockHabilidadeRepository) DeleteHabilidade(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHabilidadeRepository) ListHabilidades(ctx context.Context, filter empregabilidade.HabilidadeFilter, limit, offset int) ([]*empregabilidade.Habilidade, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*empregabilidade.Habilidade), args.Get(1).(int64), args.Error(2)
}

func (m *MockHabilidadeRepository) CreateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) (int64, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockHabilidadeRepository) GetAreaAtuacaoByID(ctx context.Context, id int64) (*empregabilidade.AreaAtuacao, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*empregabilidade.AreaAtuacao), args.Error(1)
}

func (m *MockHabilidadeRepository) UpdateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockHabilidadeRepository) DeleteAreaAtuacao(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHabilidadeRepository) ListAreasAtuacao(ctx context.Context, filter empregabilidade.AreaAtuacaoFilter, limit, offset int) ([]*empregabilidade.AreaAtuacao, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*empregabilidade.AreaAtuacao), args.Get(1).(int64), args.Error(2)
}

func (m *MockHabilidadeRepository) AddHabilidadeAoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoHabilidade) error {
	args := m.Called(ctx, vinculo)
	return args.Error(0)
}

func (m *MockHabilidadeRepository) ListHabilidadesPorCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoHabilidade, error) {
	args := m.Called(ctx, cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*empregabilidade.CurriculoHabilidade), args.Error(1)
}

// --- TESTES UNITÁRIOS PARA HABILIDADES ---

func TestHabilidadeService_CreateHabilidade(t *testing.T) {
	ctx := context.Background()
	var habilidadeID int64 = 10
	entity := &empregabilidade.Habilidade{Nome: "Golang"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("CreateHabilidade", ctx, entity).Return(habilidadeID, nil)

		id, err := svc.CreateHabilidade(ctx, entity)

		assert.NoError(t, err)
		assert.Equal(t, habilidadeID, id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("erro ao criar no banco")

		mockRepo.On("CreateHabilidade", ctx, entity).Return(int64(0), expectedErr)

		id, err := svc.CreateHabilidade(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_GetHabilidadeByID(t *testing.T) {
	ctx := context.Background()
	var habilidadeID int64 = 15
	expectedHabilidade := &empregabilidade.Habilidade{ID: habilidadeID, Nome: "PostgreSQL"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("GetHabilidadeByID", ctx, habilidadeID).Return(expectedHabilidade, nil)

		result, err := svc.GetHabilidadeByID(ctx, habilidadeID)

		assert.NoError(t, err)
		assert.Equal(t, expectedHabilidade, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found or Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("habilidade não encontrada")

		mockRepo.On("GetHabilidadeByID", ctx, habilidadeID).Return(nil, expectedErr)

		result, err := svc.GetHabilidadeByID(ctx, habilidadeID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_UpdateHabilidade(t *testing.T) {
	ctx := context.Background()
	entity := &empregabilidade.Habilidade{ID: 20, Nome: "Golang Avançado"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("UpdateHabilidade", ctx, entity).Return(nil)

		err := svc.UpdateHabilidade(ctx, entity)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("falha ao atualizar")

		mockRepo.On("UpdateHabilidade", ctx, entity).Return(expectedErr)

		err := svc.UpdateHabilidade(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_DeleteHabilidade(t *testing.T) {
	ctx := context.Background()
	var habilidadeID int64 = 25

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("DeleteHabilidade", ctx, habilidadeID).Return(nil)

		err := svc.DeleteHabilidade(ctx, habilidadeID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("falha ao remover")

		mockRepo.On("DeleteHabilidade", ctx, habilidadeID).Return(expectedErr)

		err := svc.DeleteHabilidade(ctx, habilidadeID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_ListHabilidades(t *testing.T) {
	ctx := context.Background()
	filter := empregabilidade.HabilidadeFilter{Search: "Go"}
	expectedList := []*empregabilidade.Habilidade{
		{ID: 1, Nome: "Golang"},
	}

	t.Run("Success with Page Calculation", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		page := 2
		pageSize := 10
		expectedOffset := 10 // (2 - 1) * 10

		mockRepo.On("ListHabilidades", ctx, filter, pageSize, expectedOffset).
			Return(expectedList, int64(1), nil)

		result, total, err := svc.ListHabilidades(ctx, filter, page, pageSize)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("erro ao listar")

		mockRepo.On("ListHabilidades", ctx, filter, 10, 0).
			Return(nil, int64(0), expectedErr)

		result, total, err := svc.ListHabilidades(ctx, filter, 1, 10)

		assert.Error(t, err)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

// --- TESTES UNITÁRIOS PARA ÁREAS DE ATUAÇÃO ---

func TestHabilidadeService_CreateAreaAtuacao(t *testing.T) {
	ctx := context.Background()
	var areaID int64 = 100
	entity := &empregabilidade.AreaAtuacao{Nome: "Tecnologia da Informação"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("CreateAreaAtuacao", ctx, entity).Return(areaID, nil)

		id, err := svc.CreateAreaAtuacao(ctx, entity)

		assert.NoError(t, err)
		assert.Equal(t, areaID, id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("erro ao criar área de atuação no banco")

		mockRepo.On("CreateAreaAtuacao", ctx, entity).Return(int64(0), expectedErr)

		id, err := svc.CreateAreaAtuacao(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_GetAreaAtuacaoByID(t *testing.T) {
	ctx := context.Background()
	var areaID int64 = 100
	expectedArea := &empregabilidade.AreaAtuacao{ID: areaID, Nome: "Tecnologia da Informação"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("GetAreaAtuacaoByID", ctx, areaID).Return(expectedArea, nil)

		result, err := svc.GetAreaAtuacaoByID(ctx, areaID)

		assert.NoError(t, err)
		assert.Equal(t, expectedArea, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found or Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("área de atuação não encontrada")

		mockRepo.On("GetAreaAtuacaoByID", ctx, areaID).Return(nil, expectedErr)

		result, err := svc.GetAreaAtuacaoByID(ctx, areaID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_UpdateAreaAtuacao(t *testing.T) {
	ctx := context.Background()
	entity := &empregabilidade.AreaAtuacao{ID: 100, Nome: "Tecnologia da Informação e Comunicação"}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("UpdateAreaAtuacao", ctx, entity).Return(nil)

		err := svc.UpdateAreaAtuacao(ctx, entity)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("falha ao atualizar área de atuação")

		mockRepo.On("UpdateAreaAtuacao", ctx, entity).Return(expectedErr)

		err := svc.UpdateAreaAtuacao(ctx, entity)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_DeleteAreaAtuacao(t *testing.T) {
	ctx := context.Background()
	var areaID int64 = 100

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("DeleteAreaAtuacao", ctx, areaID).Return(nil)

		err := svc.DeleteAreaAtuacao(ctx, areaID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("falha ao remover área de atuação")

		mockRepo.On("DeleteAreaAtuacao", ctx, areaID).Return(expectedErr)

		err := svc.DeleteAreaAtuacao(ctx, areaID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_ListAreas(t *testing.T) {
	ctx := context.Background()
	filter := empregabilidade.AreaAtuacaoFilter{Search: "Tecnologia"}
	expectedAreas := []*empregabilidade.AreaAtuacao{
		{ID: 100, Nome: "Tecnologia da Informação"},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		page := 3
		pageSize := 5
		expectedOffset := 10 // (3 - 1) * 5

		mockRepo.On("ListAreasAtuacao", ctx, filter, pageSize, expectedOffset).
			Return(expectedAreas, int64(1), nil)

		result, total, err := svc.ListAreasAtuacao(ctx, filter, page, pageSize)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_AddHabilidadeAoCurriculo(t *testing.T) {
	ctx := context.Background()
	vinculo := &empregabilidade.CurriculoHabilidade{
		ID:           1,
		CPF:          "12345678901",
		IDHabilidade: 10,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("AddHabilidadeAoCurriculo", ctx, vinculo).Return(nil)

		err := svc.AddHabilidadeAoCurriculo(ctx, vinculo)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("erro ao vincular habilidade ao currículo")

		mockRepo.On("AddHabilidadeAoCurriculo", ctx, vinculo).Return(expectedErr)

		err := svc.AddHabilidadeAoCurriculo(ctx, vinculo)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestHabilidadeService_ListHabilidadesPorCPF(t *testing.T) {
	ctx := context.Background()
	cpf := "12345678901"
	expectedList := []*empregabilidade.CurriculoHabilidade{
		{ID: 1, CPF: cpf, IDHabilidade: 10},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)

		mockRepo.On("ListHabilidadesPorCPF", ctx, cpf).Return(expectedList, nil)

		result, err := svc.ListHabilidadesPorCPF(ctx, cpf)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, cpf, result[0].CPF)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockHabilidadeRepository)
		svc := NewHabilidadeServiceWithInterface(mockRepo)
		expectedErr := errors.New("erro ao buscar habilidades do CPF")

		mockRepo.On("ListHabilidadesPorCPF", ctx, cpf).Return(nil, expectedErr)

		result, err := svc.ListHabilidadesPorCPF(ctx, cpf)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
