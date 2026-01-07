package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

// CNAEValidationService handles CNAE validation for MEI proposals
type CNAEValidationService struct {
	rmiClient RMIClientInterface
	cache     LegalEntitiesCacheInterface
}

// NewCNAEValidationService creates a new CNAE validation service
func NewCNAEValidationService(rmiClient RMIClientInterface, cache LegalEntitiesCacheInterface) *CNAEValidationService {
	return &CNAEValidationService{
		rmiClient: rmiClient,
		cache:     cache,
	}
}

// CheckCNPJOwnership checks if a CPF owns a specific CNPJ
// Returns true if the CPF owns the CNPJ, false otherwise
func (s *CNAEValidationService) CheckCNPJOwnership(
	ctx context.Context,
	authToken string,
	cpf string,
	cnpj string,
) (bool, error) {
	// Get user's legal entities (with caching)
	legalEntities, err := s.getUserLegalEntities(ctx, authToken, cpf)
	if err != nil {
		log.Printf("[CNPJ_OWNERSHIP_ERROR] CPF=%s Error fetching legal entities: %v", cpf, err)
		return false, err
	}

	// Check if user has any CNPJs
	if len(legalEntities) == 0 {
		return false, nil
	}

	// Find the CNPJ in user's legal entities
	for i := range legalEntities {
		// Normalize CNPJ for comparison (remove formatting)
		normalizedCNPJ := utils.ExtractDigits(legalEntities[i].CNPJ)
		normalizedProvidedCNPJ := utils.ExtractDigits(cnpj)

		if normalizedCNPJ == normalizedProvidedCNPJ {
			return true, nil
		}
	}

	return false, nil
}

// ValidatePropostaForCNAE validates that:
// 1. CNPJ belongs to the user (via CPF from JWT)
// 2. CNPJ has at least one CNAE matching opportunity's allowed CNAEs
func (s *CNAEValidationService) ValidatePropostaForCNAE(
	ctx context.Context,
	authToken string,
	cnpj string,
	opportunityCNAEIDs []string,
) error {
	// 1. Extract CPF from JWT token
	cpf, err := utils.ExtractCPFFromToken(authToken)
	if err != nil {
		log.Printf("[CNAE_VALIDATION_ERROR] Failed to extract CPF from token: %v", err)
		return fmt.Errorf("não foi possível extrair CPF do token de autenticação")
	}

	// 2. Get user's legal entities (with caching)
	legalEntities, err := s.getUserLegalEntities(ctx, authToken, cpf)
	if err != nil {
		log.Printf("[CNAE_VALIDATION_ERROR] CPF=%s Error fetching legal entities: %v", cpf, err)
		return fmt.Errorf("não foi possível validar seus CNPJs no momento. Tente novamente em instantes")
	}

	// 3. Check if user has any CNPJs
	if len(legalEntities) == 0 {
		log.Printf("[CNAE_VALIDATION_FAILED] CPF=%s CNPJ=%s Reason=no_cnpjs", cpf, cnpj)
		return fmt.Errorf("nenhum CNPJ encontrado vinculado ao seu CPF")
	}

	// 4. Find the CNPJ in user's legal entities
	var foundEntity *models.LegalEntity
	for i := range legalEntities {
		// Normalize CNPJ for comparison (remove formatting)
		normalizedCNPJ := utils.ExtractDigits(legalEntities[i].CNPJ)
		normalizedProvidedCNPJ := utils.ExtractDigits(cnpj)

		if normalizedCNPJ == normalizedProvidedCNPJ {
			foundEntity = &legalEntities[i]
			break
		}
	}

	if foundEntity == nil {
		log.Printf("[CNAE_VALIDATION_FAILED] CPF=%s CNPJ=%s Reason=ownership", cpf, cnpj)
		return fmt.Errorf("o CNPJ %s não pertence ao seu CPF", cnpj)
	}

	// 5. Get all CNAEs from the CNPJ (fiscal + secundarias)
	cnpjCNAEs := foundEntity.GetAllCNAEs()

	// 6. Check if any CNPJ CNAE matches any opportunity CNAE
	if !utils.HasMatchingCNAE(cnpjCNAEs, opportunityCNAEIDs) {
		log.Printf("[CNAE_VALIDATION_FAILED] CPF=%s CNPJ=%s Reason=cnae_mismatch CNPJCNAEs=%v OpportunityCNAEs=%v",
			cpf, cnpj, cnpjCNAEs, opportunityCNAEIDs)

		// Format CNAEs for user-friendly error message
		formattedCNAEs := strings.Join(opportunityCNAEIDs, ", ")
		return fmt.Errorf("o CNPJ %s não possui CNAE compatível com esta oportunidade. CNAEs aceitos: %s",
			cnpj, formattedCNAEs)
	}

	// Success!
	log.Printf("[CNAE_VALIDATION_SUCCESS] CPF=%s CNPJ=%s", cpf, cnpj)
	return nil
}

// getUserLegalEntities fetches legal entities with caching
func (s *CNAEValidationService) getUserLegalEntities(ctx context.Context, authToken string, cpf string) ([]models.LegalEntity, error) {
	// Try cache first
	cached, err := s.cache.Get(ctx, cpf)
	if err != nil {
		// Cache error - log but continue with API call
		log.Printf("[CACHE_ERROR] Failed to get from cache for CPF=%s: %v", cpf, err)
	} else if cached != nil {
		// Cache hit
		log.Printf("[CACHE_HIT] CPF=%s Found %d legal entities", cpf, len(cached))
		return cached, nil
	}

	// Cache miss - fetch from RMI API
	log.Printf("[CACHE_MISS] CPF=%s Fetching from RMI API", cpf)
	entities, err := s.rmiClient.GetUserLegalEntities(ctx, authToken, cpf)
	if err != nil {
		return nil, err
	}

	// Store in cache (ignore cache errors - not critical)
	if err := s.cache.Set(ctx, cpf, entities); err != nil {
		log.Printf("[CACHE_ERROR] Failed to set cache for CPF=%s: %v", cpf, err)
	} else {
		log.Printf("[CACHE_SET] CPF=%s Cached %d legal entities", cpf, len(entities))
	}

	return entities, nil
}
