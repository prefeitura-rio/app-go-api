package empregabilidade_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// capturaCriacaoRepo guarda o estado da candidatura no instante em que o
// service a entrega ao repositório.
type capturaCriacaoRepo struct {
	*MockCandidaturaRepo
	recebida empregabilidade.Candidatura
}

func (r *capturaCriacaoRepo) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	r.recebida = *entity
	return r.MockCandidaturaRepo.Create(ctx, entity)
}

// Os campos de servidor vindos do corpo (id, deleted_at, created_at,
// updated_at) nunca chegam ao repositório: o id é do banco e a candidatura não
// nasce removida, nos dois caminhos de criação (DEV-775).
func TestCandidaturaService_Create_DescartaCamposDoServidor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		criar func(*services.CandidaturaService, *empregabilidade.Candidatura) (uuid.UUID, error)
	}{
		{"Create", func(s *services.CandidaturaService, c *empregabilidade.Candidatura) (uuid.UUID, error) {
			return s.Create(context.Background(), c)
		}},
		{"CreateSemNotificacao", func(s *services.CandidaturaService, c *empregabilidade.Candidatura) (uuid.UUID, error) {
			return s.CreateSemNotificacao(context.Background(), c)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &capturaCriacaoRepo{MockCandidaturaRepo: NewMockCandidaturaRepo()}
			vagaRepo := NewMockVagaRepo()
			vagaID, _ := novaVagaComEtapa(vagaRepo)
			service := services.NewCandidaturaService(repo, vagaRepo, NewMockCurriculoService(), nil, nil, newDisabledEmailSvc())

			idForjado := uuid.New()
			forjado := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
			id, err := tc.criar(service, &empregabilidade.Candidatura{
				ID:        idForjado,
				CPF:       "12345678900",
				IDVaga:    vagaID,
				DeletedAt: gorm.DeletedAt{Time: forjado, Valid: true},
				CreatedAt: forjado,
				UpdatedAt: forjado,
			})

			require.NoError(t, err)
			assert.Equal(t, uuid.Nil, repo.recebida.ID, "o id não pode vir do corpo")
			assert.False(t, repo.recebida.DeletedAt.Valid, "a candidatura não pode nascer removida")
			assert.True(t, repo.recebida.CreatedAt.IsZero(), "created_at é da API")
			assert.True(t, repo.recebida.UpdatedAt.IsZero(), "updated_at é da API")
			assert.NotEqual(t, idForjado, id)
		})
	}
}
