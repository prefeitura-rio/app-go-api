package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type CurriculoRepository struct {
	db *gorm.DB
}

func NewCurriculoRepository(db *gorm.DB) *CurriculoRepository {
	return &CurriculoRepository{db: db}
}

// Formação Acadêmica

func (r *CurriculoRepository) CreateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar formação: %w", err)
	}
	return entity.ID, nil
}

func (r *CurriculoRepository) GetFormacaoByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoFormacao, error) {
	var entity empregabilidade.CurriculoFormacao
	result := r.db.WithContext(ctx).Preload("Escolaridade").First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar formação: %w", result.Error)
	}
	return &entity, nil
}

func (r *CurriculoRepository) UpdateFormacao(ctx context.Context, entity *empregabilidade.CurriculoFormacao) error {
	result := r.db.WithContext(ctx).Omit("created_at").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar formação: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) DeleteFormacao(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.CurriculoFormacao{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir formação: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) ListFormacoesByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoFormacao, error) {
	var entities []*empregabilidade.CurriculoFormacao
	result := r.db.WithContext(ctx).Preload("Escolaridade").Where("cpf = ?", cpf).Order("ano_conclusao DESC").Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar formações: %w", result.Error)
	}
	return entities, nil
}

// Idiomas

func (r *CurriculoRepository) CreateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar idioma: %w", err)
	}
	return entity.ID, nil
}

func (r *CurriculoRepository) GetIdiomaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoIdioma, error) {
	var entity empregabilidade.CurriculoIdioma
	result := r.db.WithContext(ctx).Preload("Idioma").Preload("Nivel").First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar idioma: %w", result.Error)
	}
	return &entity, nil
}

func (r *CurriculoRepository) UpdateIdioma(ctx context.Context, entity *empregabilidade.CurriculoIdioma) error {
	result := r.db.WithContext(ctx).Omit("created_at").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar idioma: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) DeleteIdioma(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.CurriculoIdioma{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir idioma: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) ListIdiomasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoIdioma, error) {
	var entities []*empregabilidade.CurriculoIdioma
	result := r.db.WithContext(ctx).Preload("Idioma").Preload("Nivel").Where("cpf = ?", cpf).Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar idiomas: %w", result.Error)
	}
	return entities, nil
}

// Cursos Complementares

func (r *CurriculoRepository) CreateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar curso complementar: %w", err)
	}
	return entity.ID, nil
}

func (r *CurriculoRepository) GetCursoComplementarByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoCursoComplementar, error) {
	var entity empregabilidade.CurriculoCursoComplementar
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar curso complementar: %w", result.Error)
	}
	return &entity, nil
}

func (r *CurriculoRepository) UpdateCursoComplementar(ctx context.Context, entity *empregabilidade.CurriculoCursoComplementar) error {
	result := r.db.WithContext(ctx).Omit("created_at").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar curso complementar: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) DeleteCursoComplementar(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.CurriculoCursoComplementar{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir curso complementar: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) ListCursosComplementaresByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoCursoComplementar, error) {
	var entities []*empregabilidade.CurriculoCursoComplementar
	result := r.db.WithContext(ctx).Where("cpf = ?", cpf).Order("ano_conclusao DESC").Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar cursos complementares: %w", result.Error)
	}
	return entities, nil
}

// Experiências

func (r *CurriculoRepository) CreateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar experiência: %w", err)
	}
	return entity.ID, nil
}

func (r *CurriculoRepository) GetExperienciaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoExperiencia, error) {
	var entity empregabilidade.CurriculoExperiencia
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar experiência: %w", result.Error)
	}
	return &entity, nil
}

func (r *CurriculoRepository) UpdateExperiencia(ctx context.Context, entity *empregabilidade.CurriculoExperiencia) error {
	result := r.db.WithContext(ctx).Omit("created_at").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar experiência: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) DeleteExperiencia(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.CurriculoExperiencia{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir experiência: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) ListExperienciasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoExperiencia, error) {
	var entities []*empregabilidade.CurriculoExperiencia
	result := r.db.WithContext(ctx).Where("cpf = ?", cpf).Order("eh_trabalho_atual DESC, created_at DESC").Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar experiências: %w", result.Error)
	}
	return entities, nil
}

// Conquistas

func (r *CurriculoRepository) CreateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) (uuid.UUID, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar conquista: %w", err)
	}
	return entity.ID, nil
}

func (r *CurriculoRepository) GetConquistaByID(ctx context.Context, id uuid.UUID) (*empregabilidade.CurriculoConquista, error) {
	var entity empregabilidade.CurriculoConquista
	result := r.db.WithContext(ctx).Preload("TipoConquista").First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar conquista: %w", result.Error)
	}
	return &entity, nil
}

func (r *CurriculoRepository) UpdateConquista(ctx context.Context, entity *empregabilidade.CurriculoConquista) error {
	result := r.db.WithContext(ctx).Omit("created_at").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar conquista: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) DeleteConquista(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.CurriculoConquista{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir conquista: %w", result.Error)
	}
	return nil
}

func (r *CurriculoRepository) ListConquistasByCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoConquista, error) {
	var entities []*empregabilidade.CurriculoConquista
	result := r.db.WithContext(ctx).Preload("TipoConquista").Where("cpf = ?", cpf).Order("created_at DESC").Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar conquistas: %w", result.Error)
	}
	return entities, nil
}

func (r *CurriculoRepository) GetPerfilByCPF(ctx context.Context, cpf string) (*empregabilidade.CurriculoPerfil, error) {
	var entity empregabilidade.CurriculoPerfil
	result := r.db.WithContext(ctx).First(&entity, "cpf = ?", cpf)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar perfil: %w", result.Error)
	}
	return &entity, nil
}

// ReplaceAll methods (bulk section save)

func (r *CurriculoRepository) ReplaceAllFormacoesByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoFormacao) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoFormacao{}).Error; err != nil {
			return fmt.Errorf("erro ao remover formações: %w", err)
		}
		if len(items) > 0 {
			for _, item := range items {
				item.CPF = cpf
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("erro ao inserir formações: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) ReplaceAllFormacaoAccordionByCPF(ctx context.Context, cpf string, formacoes []*empregabilidade.CurriculoFormacao, idiomas []*empregabilidade.CurriculoIdioma) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoFormacao{}).Error; err != nil {
			return fmt.Errorf("erro ao remover formações: %w", err)
		}
		if len(formacoes) > 0 {
			for _, item := range formacoes {
				item.CPF = cpf
			}
			if err := tx.Create(&formacoes).Error; err != nil {
				return fmt.Errorf("erro ao inserir formações: %w", err)
			}
		}
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoIdioma{}).Error; err != nil {
			return fmt.Errorf("erro ao remover idiomas: %w", err)
		}
		if len(idiomas) > 0 {
			for _, item := range idiomas {
				item.CPF = cpf
			}
			if err := tx.Create(&idiomas).Error; err != nil {
				return fmt.Errorf("erro ao inserir idiomas: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) ReplaceAllExperienciasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoExperiencia) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoExperiencia{}).Error; err != nil {
			return fmt.Errorf("erro ao remover experiências: %w", err)
		}
		if len(items) > 0 {
			for _, item := range items {
				item.CPF = cpf
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("erro ao inserir experiências: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) ReplaceAllExperienciaProfissionalAccordionByCPF(ctx context.Context, cpf string, experiencias []*empregabilidade.CurriculoExperiencia, conquistas []*empregabilidade.CurriculoConquista, resumoProfissional string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoExperiencia{}).Error; err != nil {
			return fmt.Errorf("erro ao remover experiências: %w", err)
		}
		if len(experiencias) > 0 {
			for _, item := range experiencias {
				item.CPF = cpf
			}
			if err := tx.Create(&experiencias).Error; err != nil {
				return fmt.Errorf("erro ao inserir experiências: %w", err)
			}
		}
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoConquista{}).Error; err != nil {
			return fmt.Errorf("erro ao remover conquistas: %w", err)
		}
		if len(conquistas) > 0 {
			for _, item := range conquistas {
				item.CPF = cpf
			}
			if err := tx.Create(&conquistas).Error; err != nil {
				return fmt.Errorf("erro ao inserir conquistas: %w", err)
			}
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cpf"}},
			DoUpdates: clause.AssignmentColumns([]string{"resumo_profissional", "updated_at"}),
		}).Create(&empregabilidade.CurriculoPerfil{
			CPF:                cpf,
			ResumoProfissional: resumoProfissional,
		}).Error; err != nil {
			return fmt.Errorf("erro ao salvar resumo profissional: %w", err)
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) AreaAtuacaoExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("erro ao buscar área de atuação: %w", err)
	}
	return count > 0, nil
}

func (r *CurriculoRepository) PivotIDsDaArea(ctx context.Context, areaID int64, habilidadeIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64, len(habilidadeIDs))
	if len(habilidadeIDs) == 0 {
		return out, nil
	}
	var rows []empregabilidade.AreaAtuacaoHabilidade
	if err := r.db.WithContext(ctx).
		Where("id_area_atuacao = ? AND id_habilidade IN ?", areaID, habilidadeIDs).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar conhecimentos da área: %w", err)
	}
	for _, row := range rows {
		out[row.IDHabilidade] = row.ID
	}
	return out, nil
}

func (r *CurriculoRepository) ExistingComportamentoIDs(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	var found []int64
	if err := r.db.WithContext(ctx).
		Model(&empregabilidade.ComportamentoAtitudes{}).
		Where("id IN ?", ids).
		Pluck("id", &found).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar comportamentos/atitudes: %w", err)
	}
	return found, nil
}

func (r *CurriculoRepository) ReplaceHabilidadesCompetencias(ctx context.Context, cpf string, areaID *int64, pivotIDs, comportamentoIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureCurriculo(tx, cpf); err != nil {
			return err
		}
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoAreaAtuacaoHabilidade{}).Error; err != nil {
			return fmt.Errorf("erro ao remover conhecimentos do currículo: %w", err)
		}
		if areaID == nil {
			if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoAreaAtuacao{}).Error; err != nil {
				return fmt.Errorf("erro ao remover área de atuação do currículo: %w", err)
			}
		} else if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cpf"}},
			DoUpdates: clause.AssignmentColumns([]string{"id_area_atuacao", "updated_at"}),
		}).Create(&empregabilidade.CurriculoAreaAtuacao{
			CPF:           cpf,
			IDAreaAtuacao: *areaID,
		}).Error; err != nil {
			return fmt.Errorf("erro ao gravar área de atuação do currículo: %w", err)
		}

		if len(pivotIDs) > 0 {
			links := make([]*empregabilidade.CurriculoAreaAtuacaoHabilidade, len(pivotIDs))
			for i, id := range pivotIDs {
				links[i] = &empregabilidade.CurriculoAreaAtuacaoHabilidade{
					CPF:                     cpf,
					IDAreaAtuacaoHabilidade: id,
				}
			}
			if err := tx.Create(&links).Error; err != nil {
				return fmt.Errorf("erro ao gravar conhecimentos do currículo: %w", err)
			}
		}

		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoComportamentoAtitudes{}).Error; err != nil {
			return fmt.Errorf("erro ao remover comportamentos/atitudes do currículo: %w", err)
		}
		if len(comportamentoIDs) > 0 {
			links := make([]*empregabilidade.CurriculoComportamentoAtitudes, len(comportamentoIDs))
			for i, id := range comportamentoIDs {
				links[i] = &empregabilidade.CurriculoComportamentoAtitudes{
					CPF:                     cpf,
					IDComportamentoAtitudes: id,
				}
			}
			if err := tx.Create(&links).Error; err != nil {
				return fmt.Errorf("erro ao gravar comportamentos/atitudes do currículo: %w", err)
			}
		}
		return nil
	})
}

func (r *CurriculoRepository) GetHabilidadesCompetencias(ctx context.Context, cpf string) (*empregabilidade.HabilidadesCompetencias, error) {
	out := &empregabilidade.HabilidadesCompetencias{
		Conhecimentos:          []*empregabilidade.CatalogoItem{},
		ComportamentosAtitudes: []*empregabilidade.CatalogoItem{},
	}

	var area empregabilidade.CurriculoAreaAtuacao
	err := r.db.WithContext(ctx).Preload("AreaAtuacao").Where("cpf = ?", cpf).First(&area).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("erro ao buscar área de atuação do currículo: %w", err)
	}
	if err == nil && area.AreaAtuacao != nil {
		out.AreaAtuacao = &empregabilidade.CatalogoItem{ID: area.AreaAtuacao.ID, Nome: area.AreaAtuacao.Nome}
	}

	if err := r.db.WithContext(ctx).
		Table("emp_habilidades").
		Select("emp_habilidades.id, emp_habilidades.nome").
		Joins("JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id").
		Joins("JOIN emp_curriculo_area_atuacao_habilidade c ON c.id_area_atuacao_habilidade = aah.id").
		Where("c.cpf = ?", cpf).
		Order("emp_habilidades.nome ASC").
		Scan(&out.Conhecimentos).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar conhecimentos do currículo: %w", err)
	}
	if out.Conhecimentos == nil {
		out.Conhecimentos = []*empregabilidade.CatalogoItem{}
	}

	if err := r.db.WithContext(ctx).
		Table("emp_comportamento_atitudes").
		Select("emp_comportamento_atitudes.id, emp_comportamento_atitudes.nome").
		Joins("JOIN emp_curriculo_comportamento_atitudes c ON c.id_comportamento_atitudes = emp_comportamento_atitudes.id").
		Where("c.cpf = ?", cpf).
		Order("emp_comportamento_atitudes.nome ASC").
		Scan(&out.ComportamentosAtitudes).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar comportamentos/atitudes do currículo: %w", err)
	}
	if out.ComportamentosAtitudes == nil {
		out.ComportamentosAtitudes = []*empregabilidade.CatalogoItem{}
	}
	return out, nil
}

func (r *CurriculoRepository) ReplaceAllConquistasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoConquista) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoConquista{}).Error; err != nil {
			return fmt.Errorf("erro ao remover conquistas: %w", err)
		}
		if len(items) > 0 {
			for _, item := range items {
				item.CPF = cpf
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("erro ao inserir conquistas: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) ReplaceAllIdiomasByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoIdioma) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoIdioma{}).Error; err != nil {
			return fmt.Errorf("erro ao remover idiomas: %w", err)
		}
		if len(items) > 0 {
			for _, item := range items {
				item.CPF = cpf
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("erro ao inserir idiomas: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

func (r *CurriculoRepository) ReplaceAllCursosComplementaresByCPF(ctx context.Context, cpf string, items []*empregabilidade.CurriculoCursoComplementar) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cpf = ?", cpf).Delete(&empregabilidade.CurriculoCursoComplementar{}).Error; err != nil {
			return fmt.Errorf("erro ao remover cursos complementares: %w", err)
		}
		if len(items) > 0 {
			for _, item := range items {
				item.CPF = cpf
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("erro ao inserir cursos complementares: %w", err)
			}
		}
		return ensureCurriculo(tx, cpf)
	})
}

// Situação e Interesses

func (r *CurriculoRepository) UpsertSituacaoInteresses(ctx context.Context, entity *empregabilidade.CurriculoSituacaoInteresses) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Upsert pela chave, sem tocar em created_at: o Save regravava a coluna
		// com o zero do Go a cada novo salvamento.
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "cpf"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"id_situacao",
				"tempo_procurando_emprego",
				"id_disponibilidade",
				"ids_tipos_vinculo_preferencia",
				"updated_at",
			}),
		}).Create(entity).Error; err != nil {
			return err
		}
		return ensureCurriculo(tx, entity.CPF)
	})
	if err != nil {
		return fmt.Errorf("erro ao salvar situação e interesses: %w", err)
	}
	return nil
}

func (r *CurriculoRepository) GetSituacaoInteressesByCPF(ctx context.Context, cpf string) (*empregabilidade.CurriculoSituacaoInteresses, error) {
	var entity empregabilidade.CurriculoSituacaoInteresses
	result := r.db.WithContext(ctx).
		Preload("Situacao").
		Preload("Disponibilidade").
		First(&entity, "cpf = ?", cpf)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar situação e interesses: %w", result.Error)
	}
	return &entity, nil
}

