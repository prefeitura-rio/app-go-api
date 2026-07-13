package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar formação: %w", result.Error)
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
	result := r.db.WithContext(ctx).Save(entity)
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
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar idioma: %w", result.Error)
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
	result := r.db.WithContext(ctx).Save(entity)
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
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar curso complementar: %w", result.Error)
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
	result := r.db.WithContext(ctx).Save(entity)
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
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar experiência: %w", result.Error)
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
	result := r.db.WithContext(ctx).Save(entity)
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
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar conquista: %w", result.Error)
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
	result := r.db.WithContext(ctx).Save(entity)
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
		return nil
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
		return nil
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
		return nil
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
		if err := tx.Save(&empregabilidade.CurriculoPerfil{
			CPF:                cpf,
			ResumoProfissional: resumoProfissional,
		}).Error; err != nil {
			return fmt.Errorf("erro ao salvar resumo profissional: %w", err)
		}
		return nil
	})
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
		return nil
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
		return nil
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
		return nil
	})
}

// Situação e Interesses

func (r *CurriculoRepository) UpsertSituacaoInteresses(ctx context.Context, entity *empregabilidade.CurriculoSituacaoInteresses) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao salvar situação e interesses: %w", result.Error)
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
