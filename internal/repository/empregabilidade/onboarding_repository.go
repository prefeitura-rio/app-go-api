package empregabilidade

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type OnboardingRepository struct {
	db *gorm.DB
}

func NewOnboardingRepository(db *gorm.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

func (r *OnboardingRepository) Upsert(ctx context.Context, entity *empregabilidade.Onboarding) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao salvar onboarding: %w", result.Error)
	}
	return nil
}

func (r *OnboardingRepository) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error) {
	var entity empregabilidade.Onboarding
	result := r.db.WithContext(ctx).First(&entity, "cpf = ?", cpf)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar onboarding: %w", result.Error)
	}
	return &entity, nil
}

func (r *OnboardingRepository) MarkFirstLoginCompleted(ctx context.Context, cpf string) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Onboarding{}).
		Where("cpf = ?", cpf).
		Update("is_empregabilidade_first_login", false)
	if result.Error != nil {
		return fmt.Errorf("erro ao marcar primeiro login como concluído: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		entity := &empregabilidade.Onboarding{
			CPF:                         cpf,
			IsEmpregabilidadeFirstLogin: false,
		}
		if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
			return fmt.Errorf("erro ao criar registro de onboarding: %w", err)
		}
	}

	return nil
}
