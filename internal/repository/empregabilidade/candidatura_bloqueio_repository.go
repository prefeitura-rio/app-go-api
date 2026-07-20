package empregabilidade

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type CandidaturaBloqueioRepository struct {
	db *gorm.DB
}

func NewCandidaturaBloqueioRepository(db *gorm.DB) *CandidaturaBloqueioRepository {
	return &CandidaturaBloqueioRepository{db: db}
}

func (r *CandidaturaBloqueioRepository) Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar bloqueio de candidatura: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *CandidaturaBloqueioRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.CandidaturaBloqueio, int, error) {
	var entities []*empregabilidade.CandidaturaBloqueio
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.CandidaturaBloqueio{})

	// Ordena as chaves para gerar SQL determinística (a iteração de map em Go
	// é aleatória, o que tornava o WHERE — e o teste que o espera — instável).
	keys := make([]string, 0, len(filter))
	for key := range filter {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		db = db.Where(key+" = ?", filter[key])
	}

	db.Count(&total)

	result := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar bloqueios de candidatura: %w", result.Error)
	}

	return entities, int(total), nil
}
