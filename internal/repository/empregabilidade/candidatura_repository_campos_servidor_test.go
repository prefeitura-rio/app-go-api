package empregabilidade

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// instanteRecente casa com um time.Time gerado agora pela API, e nunca com o
// instante forjado no corpo.
type instanteRecente struct{ forjado time.Time }

func (m instanteRecente) Match(v driver.Value) bool {
	t, ok := v.(time.Time)
	return ok && !t.Equal(m.forjado) && time.Since(t) < time.Minute
}

// Create nunca grava os campos de servidor que vieram no corpo: o id é gerado
// pelo banco, a candidatura não nasce removida e os instantes são da API; o
// evento candidatura_criada leva o id gerado (DEV-775).
func TestCandidaturaRepository_Create_IgnoraCamposDoServidor(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	vagaID := uuid.New()
	idForjado := uuid.New()
	idGerado := uuid.New()
	forjado := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	candidatura := &empregabilidade.Candidatura{
		ID:        idForjado,
		CPF:       cpfSintetico,
		IDVaga:    vagaID,
		Status:    empregabilidade.StatusCandidaturaEnviada,
		DeletedAt: gorm.DeletedAt{Time: forjado, Valid: true},
		CreatedAt: forjado,
		UpdatedAt: forjado,
	}

	// Sem a coluna "id" no INSERT: o banco gera o id (default gen_random_uuid()).
	insert := regexp.QuoteMeta(`INSERT INTO "emp_candidaturas" ("cpf","nome","email","id_vaga","status","id_etapa_atual","status_anterior","respostas_info_complementares","curriculo_snapshot","origem","associado_por_cpf","ultimo_evento_visto","deleted_at","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)
	args := make([]driver.Value, 0, 15)
	for i := 0; i < 12; i++ {
		args = append(args, sqlmock.AnyArg())
	}
	args = append(args, nil, instanteRecente{forjado}, instanteRecente{forjado})

	mock.ExpectBegin()
	mock.ExpectQuery("^" + insert + "$").
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(idGerado))
	esperarEventos(mock, eventoEsperado{
		candidatura: idGerado, vaga: vagaID,
		tipo: empregabilidade.EventoCandidaturaCriada, ator: empregabilidade.AtorCandidato,
		statusNovo: empregabilidade.StatusCandidaturaEnviada,
	})
	mock.ExpectCommit()

	id, err := repo.Create(ctxComOrigem(empregabilidade.AtorCandidato, ""), candidatura)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, idGerado, id)
	assert.Equal(t, idGerado, candidatura.ID)
	assert.False(t, candidatura.DeletedAt.Valid, "a candidatura não pode nascer removida")
}
