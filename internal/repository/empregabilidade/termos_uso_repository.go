package empregabilidade

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type TermosUsoRepository struct {
	db *gorm.DB
}

func NewTermosUsoRepository(db *gorm.DB) *TermosUsoRepository {
	return &TermosUsoRepository{db: db}
}

func (r *TermosUsoRepository) Upsert(ctx context.Context, entity *empregabilidade.TermosUso) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao salvar termos de uso: %w", result.Error)
	}
	return nil
}

func (r *TermosUsoRepository) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error) {
	var entity empregabilidade.TermosUso
	result := r.db.WithContext(ctx).First(&entity, "cpf = ?", cpf)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar termos de uso: %w", result.Error)
	}
	return &entity, nil
}

func (r *TermosUsoRepository) AcceptTerms(ctx context.Context, cpf string) error {
	now := time.Now()
	entity := &empregabilidade.TermosUso{
		CPF:         cpf,
		UserConsent: true,
		AceitoEm:    &now,
	}

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "cpf"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_consent", "aceito_em", "updated_at"}),
		}).
		Create(entity)

	if result.Error != nil {
		return fmt.Errorf("erro ao aceitar termos de uso: %w", result.Error)
	}

	return nil
}
