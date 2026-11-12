package empregabilidade

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Vaga lida pela gestão: as listas saem sempre, vazias quando a vaga não tem
// registro (vaga antiga).
func TestVaga_SetHabilidadesCompetencias_VazioSaiComoListasVazias(t *testing.T) {
	v := Vaga{}
	v.SetHabilidadesCompetencias(nil)

	raw, err := json.Marshal(v)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	assert.Equal(t, []any{}, body["areas_atuacao"])
	assert.Equal(t, []any{}, body["comportamento_atitudes_ids"])
	assert.Equal(t, []any{}, body["comportamentos_atitudes"])
}

func TestVaga_SetHabilidadesCompetencias_PreencheIDsENomes(t *testing.T) {
	v := Vaga{}
	v.SetHabilidadesCompetencias(&VagaHabilidadesCompetencias{
		AreasAtuacao: []VagaAreaAtuacao{
			{IDAreaAtuacao: 3, Nome: "Tecnologia", Conhecimentos: []CatalogoItem{{ID: 10, Nome: "Go"}, {ID: 11, Nome: "SQL"}}},
			{IDAreaAtuacao: 4, Nome: "Saúde"},
		},
		ComportamentosAtitudes: []CatalogoItem{{ID: 7, Nome: "Proatividade"}},
	})

	require.NotNil(t, v.AreasAtuacao)
	areas := *v.AreasAtuacao
	require.Len(t, areas, 2)
	assert.Equal(t, []int64{10, 11}, areas[0].ConhecimentoIDs)
	assert.Equal(t, []int64{}, areas[1].ConhecimentoIDs)
	assert.Equal(t, []CatalogoItem{}, areas[1].Conhecimentos)
	assert.Equal(t, []int64{7}, *v.ComportamentoAtitudesIDs)
	assert.Equal(t, []CatalogoItem{{ID: 7, Nome: "Proatividade"}}, *v.ComportamentosAtitudes)
}

// Rotas públicas e do SuperApp não carregam os campos: a vaga sem eles não
// os traz no JSON.
func TestVaga_SemHabilidadesCarregadas_NaoSerializaOsCampos(t *testing.T) {
	raw, err := json.Marshal(Vaga{Titulo: "Pública"})
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	for _, campo := range []string{"areas_atuacao", "comportamento_atitudes_ids", "comportamentos_atitudes"} {
		_, presente := body[campo]
		assert.False(t, presente, "campo %s não deveria sair", campo)
	}
}

// No corpo do POST/PUT, campo ausente fica nil (o PUT mantém o gravado) e
// lista vazia fica vazia (o PUT apaga tudo).
func TestVaga_Bind_DistingueAusenteDeVazio(t *testing.T) {
	var ausente Vaga
	require.NoError(t, json.Unmarshal([]byte(`{"titulo":"x"}`), &ausente))
	assert.Nil(t, ausente.AreasAtuacao)
	assert.Nil(t, ausente.ComportamentoAtitudesIDs)

	var vazio Vaga
	require.NoError(t, json.Unmarshal([]byte(`{"areas_atuacao":[],"comportamento_atitudes_ids":[]}`), &vazio))
	require.NotNil(t, vazio.AreasAtuacao)
	assert.Empty(t, *vazio.AreasAtuacao)
	require.NotNil(t, vazio.ComportamentoAtitudesIDs)
	assert.Empty(t, *vazio.ComportamentoAtitudesIDs)

	var cheio Vaga
	require.NoError(t, json.Unmarshal([]byte(`{"areas_atuacao":[{"id_area_atuacao":3,"conhecimento_ids":[10,11]}],"comportamento_atitudes_ids":[7]}`), &cheio))
	require.NotNil(t, cheio.AreasAtuacao)
	assert.Equal(t, []VagaAreaAtuacao{{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 11}}}, *cheio.AreasAtuacao)
	assert.Equal(t, []int64{7}, *cheio.ComportamentoAtitudesIDs)
}
