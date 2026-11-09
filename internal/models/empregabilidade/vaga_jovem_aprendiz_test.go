package empregabilidade_test

import (
	"encoding/json"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJovemAprendiz_FaixaFixa(t *testing.T) {
	assert.Equal(t, 14, empregabilidade.IdadeMinimaJovemAprendiz)
	assert.Equal(t, 24, empregabilidade.IdadeMaximaJovemAprendiz)
}

func TestAtendeJovemAprendiz(t *testing.T) {
	idade := func(v int) *int { return &v }
	tests := []struct {
		name  string
		idade *int
		want  bool
	}{
		{name: "13 anos fica fora", idade: idade(13), want: false},
		{name: "14 anos entra", idade: idade(14), want: true},
		{name: "18 anos entra", idade: idade(18), want: true},
		{name: "24 anos entra", idade: idade(24), want: true},
		{name: "25 anos fica fora", idade: idade(25), want: false},
		{name: "sem idade conhecida não atende", idade: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, empregabilidade.AtendeJovemAprendiz(tt.idade))
		})
	}
}

func TestRequisitoJovemAprendiz(t *testing.T) {
	assert.Equal(t, "Jovem Aprendiz — ter entre 14 e 24 anos", empregabilidade.RequisitoJovemAprendiz())
}

func TestCriterioBloqueioJovemAprendiz(t *testing.T) {
	assert.Equal(t, empregabilidade.CriterioBloqueio("jovem_aprendiz"), empregabilidade.CriterioBloqueioJovemAprendiz)
	assert.True(t, empregabilidade.CriterioBloqueioJovemAprendiz.IsValid())

	got, err := empregabilidade.NormalizarCriteriosBloqueio([]empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioJovemAprendiz,
		empregabilidade.CriterioBloqueioExperienciaCarteira,
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioJovemAprendiz,
	})
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.CriterioBloqueio{
		empregabilidade.CriterioBloqueioEscolaridade,
		empregabilidade.CriterioBloqueioExperienciaCarteira,
		empregabilidade.CriterioBloqueioJovemAprendiz,
	}, got)
}

// "Maiores de 18" continua sendo idade_minima: não há chave de bloqueio para
// ela, e idade_minima não é aceita em criterios_bloqueio.
func TestCriterioBloqueio_IdadeMinimaNaoEhChave(t *testing.T) {
	assert.False(t, empregabilidade.CriterioBloqueio("idade_minima").IsValid())
}

func TestVaga_JSON_JovemAprendiz(t *testing.T) {
	ligado := true
	raw, err := json.Marshal(empregabilidade.Vaga{JovemAprendiz: &ligado})
	require.NoError(t, err)
	var campos map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &campos))
	assert.JSONEq(t, `true`, string(campos["jovem_aprendiz"]))

	var v empregabilidade.Vaga
	require.NoError(t, json.Unmarshal([]byte(`{"titulo":"Dev"}`), &v))
	assert.Nil(t, v.JovemAprendiz, "ausente no corpo fica nil (o PUT mantém o gravado)")
	require.NoError(t, json.Unmarshal([]byte(`{"jovem_aprendiz":false}`), &v))
	require.NotNil(t, v.JovemAprendiz)
	assert.False(t, *v.JovemAprendiz)

	assert.Error(t, json.Unmarshal([]byte(`{"jovem_aprendiz":"sim"}`), &v), "só booleano")
	assert.Error(t, json.Unmarshal([]byte(`{"jovem_aprendiz":14}`), &v), "sem número livre")
}

// Vaga lida sem valor (linha anterior à coluna) e vaga nova sem o campo
// nascem com Jovem Aprendiz desligado.
func TestVaga_Hooks_JovemAprendizDesligadoPorPadrao(t *testing.T) {
	lida := &empregabilidade.Vaga{}
	require.NoError(t, lida.AfterFind(nil))
	require.NotNil(t, lida.JovemAprendiz)
	assert.False(t, *lida.JovemAprendiz)

	nova := &empregabilidade.Vaga{}
	require.NoError(t, nova.BeforeCreate(nil))
	require.NotNil(t, nova.JovemAprendiz)
	assert.False(t, *nova.JovemAprendiz)

	ligado := true
	mantida := &empregabilidade.Vaga{JovemAprendiz: &ligado}
	require.NoError(t, mantida.AfterFind(nil))
	assert.True(t, *mantida.JovemAprendiz)
}
