package empregabilidade_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizarCriteriosBloqueio(t *testing.T) {
	tests := []struct {
		name    string
		input   []empregabilidade.CriterioBloqueio
		want    []empregabilidade.CriterioBloqueio
		wantErr string
	}{
		{name: "nil fica desligado", input: nil, want: nil},
		{name: "lista vazia não é gravada", input: []empregabilidade.CriterioBloqueio{}, want: nil},
		{
			name:  "um critério ligado",
			input: []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioIdioma},
			want:  []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioIdioma},
		},
		{
			name: "duplicados removidos e ordem fixa",
			input: []empregabilidade.CriterioBloqueio{
				empregabilidade.CriterioBloqueioIdioma,
				empregabilidade.CriterioBloqueioEscolaridade,
				empregabilidade.CriterioBloqueioIdioma,
			},
			want: []empregabilidade.CriterioBloqueio{
				empregabilidade.CriterioBloqueioEscolaridade,
				empregabilidade.CriterioBloqueioIdioma,
			},
		},
		{
			name:    "chave desconhecida é recusada",
			input:   []empregabilidade.CriterioBloqueio{"escolaridade", "altura"},
			wantErr: `critério de bloqueio desconhecido: "altura"`,
		},
		{
			name:    "chave vazia é recusada",
			input:   []empregabilidade.CriterioBloqueio{""},
			wantErr: `critério de bloqueio desconhecido: ""`,
		},
		{
			name:    "chave com caixa diferente é recusada",
			input:   []empregabilidade.CriterioBloqueio{"Idioma"},
			wantErr: `critério de bloqueio desconhecido: "Idioma"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := empregabilidade.NormalizarCriteriosBloqueio(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCriterioBloqueio_IsValid(t *testing.T) {
	assert.True(t, empregabilidade.CriterioBloqueioEscolaridade.IsValid())
	assert.True(t, empregabilidade.CriterioBloqueioIdioma.IsValid())
	assert.True(t, empregabilidade.CriterioBloqueioPCD.IsValid())
	assert.False(t, empregabilidade.CriterioBloqueio("altura").IsValid())
	assert.False(t, empregabilidade.CriterioBloqueio("").IsValid())
}

func TestVaga_AfterFind_CriteriosBloqueioDesligadosViramListaVazia(t *testing.T) {
	v := &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Repositor"}
	require.NoError(t, v.AfterFind(nil))

	require.NotNil(t, v.CriteriosBloqueio)
	assert.Empty(t, v.CriteriosBloqueio)

	body, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"criterios_bloqueio":[]`)
}

func TestVaga_AfterFind_PreservaCriteriosBloqueioLigados(t *testing.T) {
	v := &empregabilidade.Vaga{
		ID:                uuid.New(),
		Titulo:            "Repositor",
		CriteriosBloqueio: []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioEscolaridade},
	}
	require.NoError(t, v.AfterFind(nil))

	body, err := json.Marshal(v)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"criterios_bloqueio":["escolaridade"]`)
}

func TestVaga_UnmarshalCriteriosBloqueio(t *testing.T) {
	t.Run("campo ausente fica nil", func(t *testing.T) {
		var v empregabilidade.Vaga
		require.NoError(t, json.Unmarshal([]byte(`{"titulo":"x"}`), &v))
		assert.Nil(t, v.CriteriosBloqueio)
	})
	t.Run("lista vazia fica vazia e não nil", func(t *testing.T) {
		var v empregabilidade.Vaga
		require.NoError(t, json.Unmarshal([]byte(`{"criterios_bloqueio":[]}`), &v))
		require.NotNil(t, v.CriteriosBloqueio)
		assert.Empty(t, v.CriteriosBloqueio)
	})
	t.Run("lista com critérios", func(t *testing.T) {
		var v empregabilidade.Vaga
		require.NoError(t, json.Unmarshal([]byte(`{"criterios_bloqueio":["idioma"]}`), &v))
		assert.Equal(t, []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioIdioma}, v.CriteriosBloqueio)
	})
}
