package empregabilidade_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Classificação das etapas existentes (DEV-688). Executa o SQL exato da
// migration contra um banco com as migrations aplicadas (just migrate-up), numa
// transação desfeita ao fim.

const migrationClassificaTipoEtapa = "20261121140000_classifica_tipo_etapa_emp_etapas.sql"

// etapaCrua grava a etapa direto no banco (sem o hook do modelo), como as
// etapas anteriores ao DEV-532.
func etapaCrua(t *testing.T, tx *gorm.DB, vaga uuid.UUID, titulo string, ordem int, tipo *string) uuid.UUID {
	t.Helper()
	var id string
	require.NoError(t, tx.Raw(`INSERT INTO emp_etapas (id_vaga, titulo, ordem, tipo_etapa, updated_at) VALUES (?, ?, ?, ?, '2026-01-02T03:04:05Z') RETURNING id`,
		vaga, titulo, ordem, tipo).Scan(&id).Error)
	return uuid.MustParse(id)
}

func TestEtapaTipoEtapaBackfillIntegration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	vaga := vagaComPCD(t, tx, "Vaga com etapas antigas", "aberta", nil)

	type esperado struct {
		titulo   string
		tipo     *string
		inferido bool // sem tipo válido gravado: a migration classifica pelo título
	}
	casos := map[uuid.UUID]esperado{}
	nova := func(titulo string, tipo *string, tipoEsperado *string) {
		semTipo := tipo == nil || !slices.Contains(empregabilidade.TiposEtapa, strings.ToLower(strings.TrimSpace(*tipo)))
		casos[etapaCrua(t, tx, vaga, titulo, len(casos)+1, tipo)] = esperado{titulo, tipoEsperado, semTipo}
	}

	nova("Análise de currículo", nil, texto(empregabilidade.TipoEtapaAnaliseCurriculo))
	nova("ANÁLISE CURRICULAR", nil, texto(empregabilidade.TipoEtapaAnaliseCurriculo))
	nova("Triagem", texto(""), texto(empregabilidade.TipoEtapaAnaliseCurriculo))
	nova("Entrevista com o RH", texto("   "), texto(empregabilidade.TipoEtapaEntrevista))
	nova("Teste prático", nil, texto(empregabilidade.TipoEtapaTeste))
	nova("DINÂMICA DE GRUPO", nil, texto(empregabilidade.TipoEtapaTeste))
	nova("Avaliação técnica", nil, texto(empregabilidade.TipoEtapaTeste))
	nova("Oferta de emprego", nil, texto(empregabilidade.TipoEtapaProposta))
	nova("Exame admissional", nil, texto(empregabilidade.TipoEtapaProposta))
	nova("Currículo e entrevista", nil, texto(empregabilidade.TipoEtapaAnaliseCurriculo))
	nova("Inscrição", nil, texto(empregabilidade.TipoEtapaOutro))
	nova("Pré-entrevista", nil, texto(empregabilidade.TipoEtapaEntrevista))
	nova("Comprovação de documentos", nil, texto(empregabilidade.TipoEtapaOutro))
	nova("Aprovação final", nil, texto(empregabilidade.TipoEtapaOutro))
	nova("Reavaliação", nil, texto(empregabilidade.TipoEtapaOutro))
	nova("Conversa", texto(" Entrevista "), texto(empregabilidade.TipoEtapaEntrevista))
	nova("Entrevista", texto(empregabilidade.TipoEtapaProposta), texto(empregabilidade.TipoEtapaProposta))
	nova("Entrevista", texto("codigo_livre"), texto(empregabilidade.TipoEtapaEntrevista))
	nova("Prova escrita", texto("Triagem"), texto(empregabilidade.TipoEtapaTeste))

	for rodada := 1; rodada <= 2; rodada++ {
		for _, comando := range comandosDaMigration(t, migrationClassificaTipoEtapa) {
			require.NoError(t, tx.Exec(comando).Error)
		}
		for id, e := range casos {
			var gravada struct {
				Titulo    string
				TipoEtapa *string
				UpdatedAt time.Time
			}
			require.NoError(t, tx.Raw(`SELECT titulo, tipo_etapa, updated_at FROM emp_etapas WHERE id = ?`, id).Scan(&gravada).Error)
			assert.Equal(t, e.titulo, gravada.Titulo, "rodada %d: o título não muda", rodada)
			assert.Equal(t, e.tipo, gravada.TipoEtapa, "rodada %d, etapa %q", rodada, e.titulo)
			assert.True(t, gravada.UpdatedAt.Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)), "rodada %d: updated_at não muda", rodada)
			if e.inferido {
				require.NotNil(t, gravada.TipoEtapa)
				assert.Equal(t, empregabilidade.InferirTipoEtapa(e.titulo), *gravada.TipoEtapa,
					"a migration e a API classificam %q igual", e.titulo)
			}
		}
	}
}

// Etapas novas sem tipo, criadas pela API (com a vaga, sozinhas ou na edição da
// vaga), gravam o tipo inferido do título no banco.
func TestEtapaTipoEtapaInferidoIntegration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)
	ctx := context.Background()
	vagaRepo := emprepo.NewVagaRepository(tx)
	etapaRepo := emprepo.NewEtapaRepository(tx)

	var regime, modelo string
	require.NoError(t, tx.Exec(`INSERT INTO emp_empresas (cnpj) VALUES ('99999999000353') ON CONFLICT DO NOTHING`).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_regimes_contratacao LIMIT 1`).Scan(&regime).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_modelos_trabalho LIMIT 1`).Scan(&modelo).Error)

	proposta := empregabilidade.TipoEtapaProposta
	vaga := &empregabilidade.Vaga{
		Titulo: "Vaga com etapas novas", Descricao: "Descrição", IDContratante: "99999999000353",
		IDRegimeContratacao: uuid.MustParse(regime), IDModeloTrabalho: uuid.MustParse(modelo), Status: empregabilidade.StatusVagaEmEdicao,
		Etapas: []empregabilidade.Etapa{
			{Titulo: "Análise de currículo", Ordem: 1},
			{Titulo: "Conversa final", Ordem: 2, TipoEtapa: &proposta},
		},
	}
	id, err := vagaRepo.Create(ctx, vaga)
	require.NoError(t, err)

	avulsa := &empregabilidade.Etapa{IDVaga: id, Titulo: "Teste prático", Ordem: 3}
	_, err = etapaRepo.Create(ctx, avulsa)
	require.NoError(t, err)

	tipos := func() map[string]string {
		etapas, err := etapaRepo.ListByVaga(ctx, id)
		require.NoError(t, err)
		porTitulo := map[string]string{}
		for _, e := range etapas {
			require.NotNil(t, e.TipoEtapa, "etapa %q sem tipo", e.Titulo)
			porTitulo[e.Titulo] = *e.TipoEtapa
		}
		return porTitulo
	}
	assert.Equal(t, map[string]string{
		"Análise de currículo": empregabilidade.TipoEtapaAnaliseCurriculo,
		"Conversa final":       empregabilidade.TipoEtapaProposta,
		"Teste prático":        empregabilidade.TipoEtapaTeste,
	}, tipos())

	salva, err := vagaRepo.GetByID(ctx, id)
	require.NoError(t, err)
	existentes := salva.Etapas
	for i := range existentes {
		existentes[i].TipoEtapa = nil // cliente que não conhece o campo: mantém o gravado
	}
	salva.Etapas = append(existentes, empregabilidade.Etapa{Titulo: "Entrevista com o gestor", Ordem: 4})
	require.NoError(t, vagaRepo.UpdateWithAssociations(ctx, salva))
	assert.Equal(t, map[string]string{
		"Análise de currículo":    empregabilidade.TipoEtapaAnaliseCurriculo,
		"Conversa final":          empregabilidade.TipoEtapaProposta,
		"Teste prático":           empregabilidade.TipoEtapaTeste,
		"Entrevista com o gestor": empregabilidade.TipoEtapaEntrevista,
	}, tipos())
}
