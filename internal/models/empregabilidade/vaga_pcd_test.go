package empregabilidade_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func acessibilidade(a empregabilidade.AcessibilidadePCD) *empregabilidade.AcessibilidadePCD {
	return &a
}

func TestOpcaoPCD_IsValid(t *testing.T) {
	assert.True(t, empregabilidade.OpcaoPCDAberta.IsValid())
	assert.True(t, empregabilidade.OpcaoPCDPreferencial.IsValid())
	assert.True(t, empregabilidade.OpcaoPCDExclusiva.IsValid())

	for _, invalida := range []string{"", "Exclusiva", "exclusivo_pcd", "para_pcd", "todas"} {
		assert.False(t, empregabilidade.OpcaoPCD(invalida).IsValid(), invalida)
	}
}

// De/para da história PREF-632: a "Vaga PcD" desmarcada (sem acessibilidade)
// vira aberta, a marcada vira a modalidade escolhida. para_pcd (marcada, sem
// modalidade) vira preferencial, que não bloqueia ninguém.
func TestOpcaoPCDDeAcessibilidade(t *testing.T) {
	tests := []struct {
		name  string
		input *empregabilidade.AcessibilidadePCD
		want  empregabilidade.OpcaoPCD
	}{
		{name: "vaga PcD desmarcada vira aberta", input: nil, want: empregabilidade.OpcaoPCDAberta},
		{name: "preferencial", input: acessibilidade(empregabilidade.AcessibilidadePreferencialPCD), want: empregabilidade.OpcaoPCDPreferencial},
		{name: "exclusiva", input: acessibilidade(empregabilidade.AcessibilidadeExclusivoPCD), want: empregabilidade.OpcaoPCDExclusiva},
		{name: "marcada sem modalidade vira preferencial", input: acessibilidade(empregabilidade.AcessibilidadeParaPCD), want: empregabilidade.OpcaoPCDPreferencial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, empregabilidade.OpcaoPCDDeAcessibilidade(tt.input))
		})
	}
}

func TestOpcaoPCD_AcessibilidadeLegada(t *testing.T) {
	assert.Nil(t, empregabilidade.OpcaoPCDAberta.AcessibilidadeLegada())
	assert.Equal(t, acessibilidade(empregabilidade.AcessibilidadePreferencialPCD), empregabilidade.OpcaoPCDPreferencial.AcessibilidadeLegada())
	assert.Equal(t, acessibilidade(empregabilidade.AcessibilidadeExclusivoPCD), empregabilidade.OpcaoPCDExclusiva.AcessibilidadeLegada())
}

func TestCriterioBloqueioPCD(t *testing.T) {
	assert.Equal(t, empregabilidade.CriterioBloqueio("pcd"), empregabilidade.CriterioBloqueioPCD)
	assert.True(t, empregabilidade.CriterioBloqueioPCD.IsValid())

	got, err := empregabilidade.NormalizarCriteriosBloqueio([]empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioPCD,
		empregabilidade.CriterioBloqueioIdioma,
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioPCD,
	})
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioIdioma,
		empregabilidade.CriterioBloqueioPCD,
	}, got)
}

func TestVaga_OpcaoPCD_SempreDevolvida(t *testing.T) {
	t.Run("vaga lida sem opção devolve aberta", func(t *testing.T) {
		v := &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Repositor"}
		require.NoError(t, v.AfterFind(nil))
		assert.Equal(t, empregabilidade.OpcaoPCDAberta, v.OpcaoPCD)

		body, err := json.Marshal(v)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"opcao_pcd":"aberta"`)
	})

	t.Run("vaga lida com opção preserva", func(t *testing.T) {
		v := &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Repositor", OpcaoPCD: empregabilidade.OpcaoPCDExclusiva}
		require.NoError(t, v.AfterFind(nil))
		assert.Equal(t, empregabilidade.OpcaoPCDExclusiva, v.OpcaoPCD)
	})
}

func TestVaga_BeforeCreate_OpcaoPCDPadraoAberta(t *testing.T) {
	v := &empregabilidade.Vaga{Titulo: "Repositor"}
	require.NoError(t, v.BeforeCreate(nil))
	assert.Equal(t, empregabilidade.OpcaoPCDAberta, v.OpcaoPCD)

	v = &empregabilidade.Vaga{Titulo: "Repositor", OpcaoPCD: empregabilidade.OpcaoPCDPreferencial}
	require.NoError(t, v.BeforeCreate(nil))
	assert.Equal(t, empregabilidade.OpcaoPCDPreferencial, v.OpcaoPCD)
}

func TestVaga_UnmarshalOpcaoPCD(t *testing.T) {
	var v empregabilidade.Vaga
	require.NoError(t, json.Unmarshal([]byte(`{"titulo":"x"}`), &v))
	assert.Equal(t, empregabilidade.OpcaoPCD(""), v.OpcaoPCD, "campo ausente fica vazio")

	require.NoError(t, json.Unmarshal([]byte(`{"opcao_pcd":"exclusiva"}`), &v))
	assert.Equal(t, empregabilidade.OpcaoPCDExclusiva, v.OpcaoPCD)
}
