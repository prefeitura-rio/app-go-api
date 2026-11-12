package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// Áreas de atuação, conhecimentos de cada área e comportamentos/atitudes da
// vaga (DEV-524), nas tabelas da DEV-528. As FKs dessas tabelas já recusam
// área repetida na vaga, conhecimento fora da área no catálogo e conhecimento
// sem a área registrada na vaga; o service valida antes para devolver um erro
// claro.

// gravarHabilidadesCompetencias insere, na criação da vaga, as listas que
// vieram (nil ou vazia não grava nada). As áreas entram antes dos
// conhecimentos, que dependem delas.
func gravarHabilidadesCompetencias(tx *gorm.DB, entity *empregabilidade.Vaga) error {
	if entity.AreasAtuacao != nil {
		if err := inserirAreasAtuacao(tx, entity.ID, *entity.AreasAtuacao); err != nil {
			return err
		}
	}
	if entity.ComportamentoAtitudesIDs != nil {
		if err := inserirComportamentos(tx, entity.ID, *entity.ComportamentoAtitudesIDs); err != nil {
			return err
		}
	}
	return nil
}

// substituirHabilidadesCompetencias troca, na edição, o conjunto gravado pela
// lista que veio: apaga o que estava e insere o novo. Lista ausente (nil) não
// mexe no que está gravado.
func substituirHabilidadesCompetencias(tx *gorm.DB, entity *empregabilidade.Vaga) error {
	if entity.AreasAtuacao != nil {
		if err := tx.Exec("DELETE FROM emp_vagas_conhecimentos WHERE id_vaga = ?", entity.ID).Error; err != nil {
			return fmt.Errorf("erro ao remover conhecimentos da vaga: %w", err)
		}
		if err := tx.Exec("DELETE FROM emp_vagas_areas_atuacao WHERE id_vaga = ?", entity.ID).Error; err != nil {
			return fmt.Errorf("erro ao remover áreas de atuação da vaga: %w", err)
		}
		if err := inserirAreasAtuacao(tx, entity.ID, *entity.AreasAtuacao); err != nil {
			return err
		}
	}
	if entity.ComportamentoAtitudesIDs != nil {
		if err := tx.Exec("DELETE FROM emp_vagas_comportamento_atitudes WHERE id_vaga = ?", entity.ID).Error; err != nil {
			return fmt.Errorf("erro ao remover comportamentos/atitudes da vaga: %w", err)
		}
		if err := inserirComportamentos(tx, entity.ID, *entity.ComportamentoAtitudesIDs); err != nil {
			return err
		}
	}
	return nil
}

func inserirAreasAtuacao(tx *gorm.DB, vagaID uuid.UUID, areas []empregabilidade.VagaAreaAtuacao) error {
	for _, area := range areas {
		if err := tx.Exec(
			"INSERT INTO emp_vagas_areas_atuacao (id_vaga, id_area_atuacao) VALUES (?, ?)",
			vagaID, area.IDAreaAtuacao,
		).Error; err != nil {
			return fmt.Errorf("erro ao gravar área de atuação da vaga: %w", err)
		}
	}
	for _, area := range areas {
		for _, habilidadeID := range area.ConhecimentoIDs {
			if err := tx.Exec(
				"INSERT INTO emp_vagas_conhecimentos (id_vaga, id_area_atuacao, id_habilidade) VALUES (?, ?, ?)",
				vagaID, area.IDAreaAtuacao, habilidadeID,
			).Error; err != nil {
				return fmt.Errorf("erro ao gravar conhecimento da vaga: %w", err)
			}
		}
	}
	return nil
}

func inserirComportamentos(tx *gorm.DB, vagaID uuid.UUID, ids []int64) error {
	for _, id := range ids {
		if err := tx.Exec(
			"INSERT INTO emp_vagas_comportamento_atitudes (id_vaga, id_comportamento_atitudes) VALUES (?, ?)",
			vagaID, id,
		).Error; err != nil {
			return fmt.Errorf("erro ao gravar comportamento/atitude da vaga: %w", err)
		}
	}
	return nil
}

type conhecimentoDaVagaRow struct {
	IDAreaAtuacao int64
	ID            int64
	Nome          string
}

// GetHabilidadesCompetencias lê as áreas (com os conhecimentos de cada uma) e
// os comportamentos da vaga, por nome. Vaga sem registro, ou marcada como
// excluída, lê listas vazias.
func (r *VagaRepository) GetHabilidadesCompetencias(ctx context.Context, vagaID uuid.UUID) (*empregabilidade.VagaHabilidadesCompetencias, error) {
	db := r.db.WithContext(ctx)

	var areas []empregabilidade.CatalogoItem
	if err := db.Raw(`SELECT a.id, a.nome
		FROM emp_vagas_areas_atuacao va
		JOIN emp_areas_atuacao a ON a.id = va.id_area_atuacao
		JOIN emp_vagas v ON v.id = va.id_vaga AND v.deleted_at IS NULL
		WHERE va.id_vaga = ?
		ORDER BY a.nome, a.id`, vagaID).Scan(&areas).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar áreas de atuação da vaga: %w", err)
	}

	var conhecimentos []conhecimentoDaVagaRow
	if err := db.Raw(`SELECT vc.id_area_atuacao, h.id, h.nome
		FROM emp_vagas_conhecimentos vc
		JOIN emp_habilidades h ON h.id = vc.id_habilidade
		JOIN emp_vagas v ON v.id = vc.id_vaga AND v.deleted_at IS NULL
		WHERE vc.id_vaga = ?
		ORDER BY h.nome, h.id`, vagaID).Scan(&conhecimentos).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar conhecimentos da vaga: %w", err)
	}

	comportamentos := []empregabilidade.CatalogoItem{}
	if err := db.Raw(`SELECT c.id, c.nome
		FROM emp_vagas_comportamento_atitudes vca
		JOIN emp_comportamento_atitudes c ON c.id = vca.id_comportamento_atitudes
		JOIN emp_vagas v ON v.id = vca.id_vaga AND v.deleted_at IS NULL
		WHERE vca.id_vaga = ?
		ORDER BY c.nome, c.id`, vagaID).Scan(&comportamentos).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar comportamentos/atitudes da vaga: %w", err)
	}

	porArea := make(map[int64][]empregabilidade.CatalogoItem, len(areas))
	for _, c := range conhecimentos {
		porArea[c.IDAreaAtuacao] = append(porArea[c.IDAreaAtuacao], empregabilidade.CatalogoItem{ID: c.ID, Nome: c.Nome})
	}

	out := &empregabilidade.VagaHabilidadesCompetencias{
		AreasAtuacao:           make([]empregabilidade.VagaAreaAtuacao, 0, len(areas)),
		ComportamentosAtitudes: comportamentos,
	}
	for _, a := range areas {
		itens := porArea[a.ID]
		area := empregabilidade.VagaAreaAtuacao{
			IDAreaAtuacao:   a.ID,
			Nome:            a.Nome,
			ConhecimentoIDs: make([]int64, 0, len(itens)),
			Conhecimentos:   make([]empregabilidade.CatalogoItem, 0, len(itens)),
		}
		for _, item := range itens {
			area.ConhecimentoIDs = append(area.ConhecimentoIDs, item.ID)
			area.Conhecimentos = append(area.Conhecimentos, item)
		}
		out.AreasAtuacao = append(out.AreasAtuacao, area)
	}
	if out.ComportamentosAtitudes == nil {
		out.ComportamentosAtitudes = []empregabilidade.CatalogoItem{}
	}
	return out, nil
}

// AreasAtuacaoExistentes devolve, dos ids informados, os que existem no
// catálogo de áreas de atuação.
func (r *VagaRepository) AreasAtuacaoExistentes(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	var found []int64
	if err := r.db.WithContext(ctx).Raw("SELECT id FROM emp_areas_atuacao WHERE id IN ?", ids).Scan(&found).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar áreas de atuação: %w", err)
	}
	return found, nil
}

type parAreaHabilidadeRow struct {
	IDAreaAtuacao int64
	IDHabilidade  int64
}

// ConhecimentosDasAreas devolve, para cada área informada, quais dos
// conhecimentos informados pertencem a ela no catálogo
// (area_atuacao_habilidade).
func (r *VagaRepository) ConhecimentosDasAreas(ctx context.Context, areaIDs, habilidadeIDs []int64) (map[int64]map[int64]bool, error) {
	out := map[int64]map[int64]bool{}
	if len(areaIDs) == 0 || len(habilidadeIDs) == 0 {
		return out, nil
	}
	var pares []parAreaHabilidadeRow
	if err := r.db.WithContext(ctx).Raw(
		"SELECT id_area_atuacao, id_habilidade FROM area_atuacao_habilidade WHERE id_area_atuacao IN ? AND id_habilidade IN ?",
		areaIDs, habilidadeIDs,
	).Scan(&pares).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar conhecimentos das áreas de atuação: %w", err)
	}
	for _, p := range pares {
		if out[p.IDAreaAtuacao] == nil {
			out[p.IDAreaAtuacao] = map[int64]bool{}
		}
		out[p.IDAreaAtuacao][p.IDHabilidade] = true
	}
	return out, nil
}

// ComportamentosExistentes devolve, dos ids informados, os que existem no
// catálogo de comportamentos/atitudes.
func (r *VagaRepository) ComportamentosExistentes(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	var found []int64
	if err := r.db.WithContext(ctx).Raw("SELECT id FROM emp_comportamento_atitudes WHERE id IN ?", ids).Scan(&found).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar comportamentos/atitudes: %w", err)
	}
	return found, nil
}
