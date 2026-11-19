package empregabilidade_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	emprepo "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// Backfill do bloqueio PcD nas vagas exclusivas migradas (DEV-656). Executa o
// SQL exato da migration contra um banco com as migrations aplicadas (just
// migrate-up), numa transação desfeita ao fim.

const migrationBackfillBloqueioPCD = "20261119130000_backfill_criterio_bloqueio_pcd_emp_vagas.sql"

// vagaComPCD cria uma vaga com a opção PcD e o criterios_bloqueio cru (SQL),
// sem passar pelo repositório, como as vagas que a migration de opcao_pcd
// deixou para trás.
func vagaComPCD(t *testing.T, tx *gorm.DB, titulo, opcao string, criteriosSQL *string) uuid.UUID {
	t.Helper()
	var regime, modelo string
	require.NoError(t, tx.Exec(`INSERT INTO emp_empresas (cnpj) VALUES ('99999999000353') ON CONFLICT DO NOTHING`).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_regimes_contratacao LIMIT 1`).Scan(&regime).Error)
	require.NoError(t, tx.Raw(`SELECT id FROM emp_modelos_trabalho LIMIT 1`).Scan(&modelo).Error)
	var id string
	require.NoError(t, tx.Raw(`INSERT INTO emp_vagas (titulo, descricao, id_contratante, id_regime_contratacao, id_modelo_trabalho, opcao_pcd, criterios_bloqueio)
		VALUES (?, 'Descrição', '99999999000353', ?, ?, ?, CAST(? AS jsonb)) RETURNING id`,
		titulo, regime, modelo, opcao, criteriosSQL).Scan(&id).Error)
	return uuid.MustParse(id)
}

func criteriosGravados(t *testing.T, tx *gorm.DB, id uuid.UUID) *string {
	t.Helper()
	var criterios *string
	require.NoError(t, tx.Raw(`SELECT criterios_bloqueio::text FROM emp_vagas WHERE id = ?`, id).Scan(&criterios).Error)
	return criterios
}

func texto(s string) *string { return &s }

func TestVagaBloqueioPCDBackfillIntegration(t *testing.T) {
	tx := bancoCurriculosIntegrationTx(t)

	exclusivaNula := vagaComPCD(t, tx, "Exclusiva sem critérios", "exclusiva", nil)
	exclusivaJSONNull := vagaComPCD(t, tx, "Exclusiva com null", "exclusiva", texto(`null`))
	exclusivaVazia := vagaComPCD(t, tx, "Exclusiva com lista vazia", "exclusiva", texto(`[]`))
	exclusivaSemPCD := vagaComPCD(t, tx, "Exclusiva com outros critérios", "exclusiva", texto(`["escolaridade", "idioma"]`))
	exclusivaComPCD := vagaComPCD(t, tx, "Exclusiva já com PcD", "exclusiva", texto(`["pcd", "escolaridade"]`))
	aberta := vagaComPCD(t, tx, "Aberta", "aberta", nil)
	abertaComCriterio := vagaComPCD(t, tx, "Aberta com critério", "aberta", texto(`["idioma"]`))
	preferencial := vagaComPCD(t, tx, "Preferencial", "preferencial", nil)

	esperado := map[uuid.UUID]*string{
		exclusivaNula:     texto(`["pcd"]`),
		exclusivaJSONNull: texto(`["pcd"]`),
		exclusivaVazia:    texto(`["pcd"]`),
		exclusivaSemPCD:   texto(`["escolaridade", "idioma", "pcd"]`),
		exclusivaComPCD:   texto(`["pcd", "escolaridade"]`),
		aberta:            nil,
		abertaComCriterio: texto(`["idioma"]`),
		preferencial:      nil,
	}

	for rodada := 1; rodada <= 2; rodada++ {
		for _, comando := range comandosDaMigration(t, migrationBackfillBloqueioPCD) {
			require.NoError(t, tx.Exec(comando).Error)
		}
		for id, criterios := range esperado {
			assert.Equal(t, criterios, criteriosGravados(t, tx, id), "rodada %d, vaga %s", rodada, id)
		}
	}

	t.Run("GET da vaga devolve a chave pcd", func(t *testing.T) {
		repo := emprepo.NewVagaRepository(tx)
		vaga, err := repo.GetByID(context.Background(), exclusivaSemPCD)
		require.NoError(t, err)
		assert.Equal(t, []empregabilidade.CriterioBloqueio{
			empregabilidade.CriterioBloqueioEscolaridade,
			empregabilidade.CriterioBloqueioIdioma,
			empregabilidade.CriterioBloqueioPCD,
		}, vaga.CriteriosBloqueio)

		vaga, err = repo.GetByID(context.Background(), preferencial)
		require.NoError(t, err)
		assert.NotContains(t, vaga.CriteriosBloqueio, empregabilidade.CriterioBloqueioPCD)
	})
}

// O Down não desfaz o backfill: não há como saber quais vagas tinham a chave
// antes, e tirá-la desligaria o bloqueio que alguém ligou no Admin.
func TestVagaBloqueioPCDBackfill_DownNaoAlteraDados(t *testing.T) {
	conteudo, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", migrationBackfillBloqueioPCD))
	require.NoError(t, err)

	partes := strings.Split(string(conteudo), "-- +goose Down")
	require.Len(t, partes, 2, "a migration tem uma seção Up e uma Down")
	require.Contains(t, partes[0], "-- +goose Up")
	for _, linha := range strings.Split(partes[1], "\n") {
		linha = strings.TrimSpace(linha)
		assert.True(t, linha == "" || strings.HasPrefix(linha, "--") || linha == "SELECT 1;",
			"o Down só pode ter comentários e um no-op: %q", linha)
	}
}
