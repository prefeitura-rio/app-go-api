package empregabilidade

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

const bancoCurriculosFrom = `
	FROM emp_curriculos c
	LEFT JOIN citizen_snapshots cs ON cs.cpf = c.cpf`

// A profissão vem do emprego marcado como atual; sem ele, do cargo com mais tempo
// de experiência. O formulário não pede datas e salva todos os empregos com o
// mesmo created_at, então não há como saber qual é o mais recente. A ficha usa a
// mesma regra (profissaoDasExperiencias, no service).
const bancoCurriculosSelect = `
	SELECT
		c.cpf,
		c.created_at AS data_inclusao,
		NULLIF(cs.nome, '') AS nome,
		NULLIF(cs.nome_social, '') AS nome_social,
		NULLIF(cs.escolaridade, '') AS escolaridade,
		(
			SELECT NULLIF(e.cargo, '')
			FROM emp_curriculo_experiencias e
			WHERE e.cpf = c.cpf
			ORDER BY e.eh_trabalho_atual DESC, e.tempo_experiencia_meses DESC NULLS LAST, e.cargo ASC
			LIMIT 1
		) AS profissao`

// ListBancoCurriculos lista todos os currículos da base, com os dados pessoais do
// cadastro do cidadão quando existirem, do mais recente para o mais antigo.
func (r *CurriculoRepository) ListBancoCurriculos(ctx context.Context, filter empregabilidade.BancoCurriculoFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoItem, int64, error) {
	var where string
	var args []interface{}
	if search := strings.TrimSpace(filter.Search); search != "" {
		where = ` WHERE (unaccent(cs.nome) ILIKE unaccent(?) OR unaccent(cs.nome_social) ILIKE unaccent(?))`
		term := "%" + search + "%"
		args = append(args, term, term)
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*)"+bancoCurriculosFrom+where, args...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar currículos: %w", err)
	}

	items := []*empregabilidade.BancoCurriculoItem{}
	if total == 0 {
		return items, 0, nil
	}

	query := bancoCurriculosSelect + bancoCurriculosFrom + where + ` ORDER BY c.created_at DESC, c.cpf ASC LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao listar currículos: %w", err)
	}

	return items, total, nil
}

// GetCurriculoByCPF retorna o registro raiz do currículo, ou nil se o cidadão não
// tem currículo.
func (r *CurriculoRepository) GetCurriculoByCPF(ctx context.Context, cpf string) (*empregabilidade.Curriculo, error) {
	var entity empregabilidade.Curriculo
	result := r.db.WithContext(ctx).First(&entity, "cpf = ?", cpf)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar currículo: %w", result.Error)
	}
	return &entity, nil
}

// ensureCurriculo registra o currículo raiz na primeira escrita do cidadão.
// Escritas seguintes não alteram nada, preservando a data de inclusão original.
func ensureCurriculo(tx *gorm.DB, cpf string) error {
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cpf"}},
		DoNothing: true,
	}).Create(&empregabilidade.Curriculo{CPF: cpf}).Error
	if err != nil {
		return fmt.Errorf("erro ao registrar currículo: %w", err)
	}
	return nil
}
