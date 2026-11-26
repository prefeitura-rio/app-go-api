package empregabilidade_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// tipo_etapa (DEV-688 / PREF-687): uma das cinco classificações; o título da
// etapa continua livre.

func TestTiposEtapa_ListaFechada(t *testing.T) {
	assert.Equal(t, []string{"analise_curriculo", "entrevista", "teste", "proposta", "outro"}, empregabilidade.TiposEtapa)
}

func TestNormalizarTipoEtapa(t *testing.T) {
	str := func(s string) *string { return &s }
	tests := []struct {
		name     string
		tipo     *string
		esperado *string
		invalido bool
	}{
		{"ausente", nil, nil, false},
		{"vazio vira ausente", str(""), nil, false},
		{"só espaços vira ausente", str("   "), nil, false},
		{"analise_curriculo", str("analise_curriculo"), str("analise_curriculo"), false},
		{"entrevista", str("entrevista"), str("entrevista"), false},
		{"teste", str("teste"), str("teste"), false},
		{"proposta", str("proposta"), str("proposta"), false},
		{"outro", str("outro"), str("outro"), false},
		{"espaços e maiúsculas", str("  Entrevista "), str("entrevista"), false},
		{"código fora da lista", str("triagem"), nil, true},
		{"título no lugar do código", str("Análise de currículo"), nil, true},
		{"injeção", str("outro'; DROP TABLE emp_etapas;--"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := empregabilidade.NormalizarTipoEtapa(tt.tipo)
			if tt.invalido {
				require.ErrorIs(t, err, empregabilidade.ErrTipoEtapaInvalido)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.esperado, got)
		})
	}
}

func TestInferirTipoEtapa(t *testing.T) {
	tests := []struct {
		titulo   string
		esperado string
	}{
		{"Análise de currículo", empregabilidade.TipoEtapaAnaliseCurriculo},
		{"ANÁLISE CURRICULAR", empregabilidade.TipoEtapaAnaliseCurriculo},
		{"Analise de curriculo", empregabilidade.TipoEtapaAnaliseCurriculo},
		{"Triagem", empregabilidade.TipoEtapaAnaliseCurriculo},
		{"Entrevista com o RH", empregabilidade.TipoEtapaEntrevista},
		{"entrevista final", empregabilidade.TipoEtapaEntrevista},
		{"Teste prático", empregabilidade.TipoEtapaTeste},
		{"Prova de conhecimentos", empregabilidade.TipoEtapaTeste},
		{"Avaliação técnica", empregabilidade.TipoEtapaTeste},
		{"Dinâmica de grupo", empregabilidade.TipoEtapaTeste},
		{"Proposta", empregabilidade.TipoEtapaProposta},
		{"Oferta de emprego", empregabilidade.TipoEtapaProposta},
		{"Contratação", empregabilidade.TipoEtapaProposta},
		{"Exame admissional", empregabilidade.TipoEtapaProposta},
		{"Currículo e entrevista", empregabilidade.TipoEtapaAnaliseCurriculo},
		{"Pré-entrevista", empregabilidade.TipoEtapaEntrevista},
		{"Testes", empregabilidade.TipoEtapaTeste},
		{"Inscrição", empregabilidade.TipoEtapaOutro},
		// Os padrões casam só no começo de uma palavra (revisão 6.1).
		{"Comprovação de documentos", empregabilidade.TipoEtapaOutro},
		{"Aprovação final", empregabilidade.TipoEtapaOutro},
		{"Aprovado", empregabilidade.TipoEtapaOutro},
		{"Contestação", empregabilidade.TipoEtapaOutro},
		{"Reavaliação", empregabilidade.TipoEtapaOutro},
		{"", empregabilidade.TipoEtapaOutro},
		{"   ", empregabilidade.TipoEtapaOutro},
	}
	for _, tt := range tests {
		t.Run(tt.titulo, func(t *testing.T) {
			assert.Equal(t, tt.esperado, empregabilidade.InferirTipoEtapa(tt.titulo))
		})
	}
}

func TestEtapa_BeforeCreate_InfereTipoDoTitulo(t *testing.T) {
	etapa := &empregabilidade.Etapa{Titulo: "Entrevista com o gestor"}
	require.NoError(t, etapa.BeforeCreate(nil))
	require.NotNil(t, etapa.TipoEtapa)
	assert.Equal(t, empregabilidade.TipoEtapaEntrevista, *etapa.TipoEtapa)
	assert.Equal(t, "Entrevista com o gestor", etapa.Titulo, "o título continua livre")
}

func TestEtapa_BeforeCreate_MantemTipoInformado(t *testing.T) {
	tipo := empregabilidade.TipoEtapaProposta
	etapa := &empregabilidade.Etapa{Titulo: "Entrevista", TipoEtapa: &tipo}
	require.NoError(t, etapa.BeforeCreate(nil))
	assert.Equal(t, empregabilidade.TipoEtapaProposta, *etapa.TipoEtapa)
}

// A migration que classifica as etapas existentes usa as mesmas regras, na
// mesma ordem, que a inferência das etapas novas.
func TestRegrasTipoEtapa_IguaisAsDaMigration(t *testing.T) {
	conteudo, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20261121140000_classifica_tipo_etapa_emp_etapas.sql"))
	require.NoError(t, err)
	up := strings.Split(string(conteudo), "-- +goose Down")[0]

	ultimo := -1
	for _, regra := range empregabilidade.RegrasTipoEtapaPorTitulo {
		clausula := fmt.Sprintf(`WHEN titulo ~* '\m(%s)' THEN '%s'`, regra.Padrao, regra.Tipo)
		pos := strings.Index(up, clausula)
		require.GreaterOrEqual(t, pos, 0, "migration sem a cláusula %q", clausula)
		assert.Greater(t, pos, ultimo, "cláusula fora de ordem: %q", clausula)
		ultimo = pos
	}
	assert.Contains(t, up, "ELSE 'outro'")
	assert.Equal(t, len(empregabilidade.RegrasTipoEtapaPorTitulo), strings.Count(up, "WHEN titulo ~*"))
}
