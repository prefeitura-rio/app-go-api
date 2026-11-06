package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cnpjCriteriosBloqueio = "12.345.678/0001-90"

// criterios monta a lista enviada; sem argumentos é a lista vazia ([]), não nil.
func criterios(c ...empregabilidade.CriterioBloqueio) []empregabilidade.CriterioBloqueio {
	return append([]empregabilidade.CriterioBloqueio{}, c...)
}

func TestVagaService_Create_CriteriosBloqueio_DesligadosPorPadrao(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
	vaga := &empregabilidade.Vaga{Titulo: "Repositor", IDContratante: cnpjCriteriosBloqueio}

	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)

	// Nada é gravado quando todos os critérios estão desligados...
	assert.Nil(t, vagaRepo.createdCriteriosBloqueio[id])
	// ...mas a resposta devolve a lista vazia, nunca null.
	require.NotNil(t, vaga.CriteriosBloqueio)
	assert.Empty(t, vaga.CriteriosBloqueio)
}

func TestVagaService_Create_CriteriosBloqueio_ListaVaziaNaoEGravada(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
	vaga := &empregabilidade.Vaga{
		Titulo:            "Repositor",
		IDContratante:     cnpjCriteriosBloqueio,
		CriteriosBloqueio: criterios(),
	}

	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.Nil(t, vagaRepo.createdCriteriosBloqueio[id])
	assert.Empty(t, vaga.CriteriosBloqueio)
}

func TestVagaService_Create_CriteriosBloqueio_GravaLigadosNormalizados(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
	vaga := &empregabilidade.Vaga{
		Titulo:        "Repositor",
		IDContratante: cnpjCriteriosBloqueio,
		CriteriosBloqueio: criterios(
			empregabilidade.CriterioBloqueioIdioma,
			empregabilidade.CriterioBloqueioEscolaridade,
			empregabilidade.CriterioBloqueioIdioma,
		),
	}

	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)

	want := criterios(empregabilidade.CriterioBloqueioEscolaridade, empregabilidade.CriterioBloqueioIdioma)
	assert.Equal(t, want, vagaRepo.createdCriteriosBloqueio[id])
	assert.Equal(t, want, vaga.CriteriosBloqueio)
}

func TestVagaService_Create_CriteriosBloqueio_ChaveDesconhecida(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
	vaga := &empregabilidade.Vaga{
		Titulo:            "Repositor",
		IDContratante:     cnpjCriteriosBloqueio,
		CriteriosBloqueio: criterios("altura"),
	}

	id, err := svc.Create(context.Background(), vaga)
	require.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Empty(t, vagaRepo.vagas, "vaga inválida não pode ser gravada")

	var appErr *apperrors.Error
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.KindInvalid, appErr.Kind)
	assert.Contains(t, err.Error(), `"altura"`)
}

func TestVagaService_Update_CriteriosBloqueio(t *testing.T) {
	ligados := criterios(empregabilidade.CriterioBloqueioEscolaridade, empregabilidade.CriterioBloqueioIdioma)

	tests := []struct {
		name      string
		atual     []empregabilidade.CriterioBloqueio
		enviado   []empregabilidade.CriterioBloqueio
		wantSaved []empregabilidade.CriterioBloqueio
	}{
		{
			name:      "campo ausente mantém o estado gravado",
			atual:     ligados,
			enviado:   nil,
			wantSaved: ligados,
		},
		{
			name:      "campo ausente em vaga sem critérios continua sem gravar",
			atual:     nil,
			enviado:   nil,
			wantSaved: nil,
		},
		{
			name:      "lista vazia desliga tudo e nada é gravado",
			atual:     ligados,
			enviado:   criterios(),
			wantSaved: nil,
		},
		{
			name:      "liga um critério",
			atual:     nil,
			enviado:   criterios(empregabilidade.CriterioBloqueioIdioma),
			wantSaved: criterios(empregabilidade.CriterioBloqueioIdioma),
		},
		{
			name:      "normaliza duplicados e ordem",
			atual:     nil,
			enviado:   criterios(empregabilidade.CriterioBloqueioIdioma, empregabilidade.CriterioBloqueioEscolaridade, empregabilidade.CriterioBloqueioEscolaridade),
			wantSaved: ligados,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
			vagaID := uuid.New()
			vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Repositor",
				Status:            empregabilidade.StatusVagaEmEdicao,
				CriteriosBloqueio: tt.atual,
			}

			err := svc.Update(context.Background(), &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Repositor II",
				CriteriosBloqueio: tt.enviado,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.wantSaved, vagaRepo.vagas[vagaID].CriteriosBloqueio)
		})
	}
}

func TestVagaService_Update_CriteriosBloqueio_ChaveDesconhecida(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
	vagaID := uuid.New()
	original := &empregabilidade.Vaga{
		ID:                vagaID,
		Titulo:            "Repositor",
		Status:            empregabilidade.StatusVagaEmEdicao,
		CriteriosBloqueio: criterios(empregabilidade.CriterioBloqueioEscolaridade),
	}
	vagaRepo.vagas[vagaID] = original

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                vagaID,
		Titulo:            "Repositor II",
		CriteriosBloqueio: criterios("altura"),
	})
	require.Error(t, err)

	var appErr *apperrors.Error
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.KindInvalid, appErr.Kind)
	assert.Same(t, original, vagaRepo.vagas[vagaID], "vaga não pode ser alterada")
}
