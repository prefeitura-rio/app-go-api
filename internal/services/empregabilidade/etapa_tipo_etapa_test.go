package empregabilidade_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// tipo_etapa (DEV-532): espaços das pontas saem, vazio não altera, acima de 64
// caracteres é erro de validação e nada é gravado.

func strPtr(s string) *string { return &s }

func TestEtapaService_TipoEtapa_Normalizacao(t *testing.T) {
	vaga, etapa := uuid.New(), uuid.New()

	tests := []struct {
		name     string
		tipo     *string
		esperado *string
		invalido bool
	}{
		{"ausente", nil, nil, false},
		{"com espaços nas pontas", strPtr("  entrevista "), strPtr("entrevista"), false},
		{"só espaços vira ausente", strPtr("   "), nil, false},
		{"no limite de 64", strPtr(strings.Repeat("á", 64)), strPtr(strings.Repeat("á", 64)), false},
		{"acima de 64", strPtr(strings.Repeat("a", 65)), nil, true},
	}

	for _, tt := range tests {
		t.Run("create "+tt.name, func(t *testing.T) {
			repo := &mockEtapaRepository{}
			svc := services.NewEtapaServiceWithInterface(repo)
			entity := &empregabilidade.Etapa{IDVaga: vaga, Titulo: "Etapa", TipoEtapa: tt.tipo}

			_, err := svc.Create(context.Background(), entity)
			if tt.invalido {
				require.Error(t, err)
				assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.esperado, entity.TipoEtapa)
		})

		t.Run("update "+tt.name, func(t *testing.T) {
			repo := &mockEtapaRepository{items: []*empregabilidade.Etapa{{ID: etapa, IDVaga: vaga}}}
			svc := services.NewEtapaServiceWithInterface(repo)
			entity := &empregabilidade.Etapa{ID: etapa, IDVaga: vaga, Titulo: "Etapa", TipoEtapa: tt.tipo}

			err := svc.Update(context.Background(), entity)
			if tt.invalido {
				require.Error(t, err)
				assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
				assert.Empty(t, repo.updated, "nada pode ser gravado")
				return
			}
			require.NoError(t, err)
			require.Len(t, repo.updated, 1)
			assert.Equal(t, tt.esperado, repo.updated[0].TipoEtapa)
		})
	}
}

func TestVagaService_TipoEtapaInvalidoNasEtapas(t *testing.T) {
	longo := strings.Repeat("a", 65)

	t.Run("create", func(t *testing.T) {
		svc := services.NewVagaServiceWithInterfaces(NewMockVagaRepoForService(), NewMockEmpresaRepoForService(), nil)
		_, err := svc.Create(context.Background(), &empregabilidade.Vaga{Etapas: []empregabilidade.Etapa{{Titulo: "Etapa", TipoEtapa: &longo}}})
		require.Error(t, err)
		assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
	})

	t.Run("update", func(t *testing.T) {
		svc := services.NewVagaServiceWithInterfaces(NewMockVagaRepoForService(), NewMockEmpresaRepoForService(), nil)
		err := svc.Update(context.Background(), &empregabilidade.Vaga{ID: uuid.New(), Etapas: []empregabilidade.Etapa{{Titulo: "Etapa", TipoEtapa: &longo}}})
		require.Error(t, err)
		assert.True(t, apperrors.IsKind(err, apperrors.KindValidation), "esperado erro de validação, veio %v", err)
	})
}
