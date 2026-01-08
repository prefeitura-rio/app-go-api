package services

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type PropostaMEIService struct {
	repo                  PropostaMEIRepositoryInterface
	oportunidadeRepo      OportunidadeMEIRepositoryInterface
	cnaeValidationService CNAEValidationServiceInterface
	contactInfoService    *ContactInfoService
}

func NewPropostaMEIService(
	repo PropostaMEIRepositoryInterface,
	oportunidadeRepo OportunidadeMEIRepositoryInterface,
	cnaeValidationService CNAEValidationServiceInterface,
	contactInfoService *ContactInfoService,
) *PropostaMEIService {
	return &PropostaMEIService{
		repo:                  repo,
		oportunidadeRepo:      oportunidadeRepo,
		cnaeValidationService: cnaeValidationService,
		contactInfoService:    contactInfoService,
	}
}

func (s *PropostaMEIService) Create(ctx context.Context, proposta *models.PropostaMEI, authToken string) (uuid.UUID, error) {
	// Validar campos da proposta
	if err := proposta.Validate(); err != nil {
		return uuid.Nil, err
	}

	// Validar que a oportunidade existe e está ativa
	oportunidade, err := s.oportunidadeRepo.GetByID(ctx, proposta.OportunidadeMEIID)
	if err != nil {
		return uuid.Nil, err
	}
	if oportunidade == nil {
		return uuid.Nil, errors.New("oportunidade não encontrada")
	}
	if oportunidade.Status != models.StatusOportunidadeActive {
		return uuid.Nil, errors.New("oportunidade não está ativa")
	}

	// NEW: Validate CNAE compatibility
	err = s.cnaeValidationService.ValidatePropostaForCNAE(
		ctx,
		authToken,
		proposta.MEIEmpresaID, // This is the CNPJ
		oportunidade.CNAEIDs,
	)
	if err != nil {
		return uuid.Nil, err // Return validation error to user
	}

	// Verificar se já existe proposta desta empresa para esta oportunidade
	exists, err := s.repo.CheckExistingProposta(ctx, proposta.OportunidadeMEIID, proposta.MEIEmpresaID)
	if err != nil {
		return uuid.Nil, err
	}
	if exists {
		return uuid.Nil, errors.New("já existe uma proposta desta empresa para esta oportunidade")
	}

	// Definir status inicial
	proposta.StatusAdmin = models.StatusPropostaAdminActive
	proposta.StatusCidadao = models.StatusPropostaCidadaoSubmitted

	return s.repo.Create(ctx, proposta)
}

func (s *PropostaMEIService) GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	proposta, err := s.repo.GetByID(ctx, id)
	if err != nil || proposta == nil {
		return proposta, err
	}

	// Enrich with contact info (graceful degradation on error)
	s.enrichPropostaWithContactInfo(ctx, proposta)

	return proposta, nil
}

func (s *PropostaMEIService) Update(ctx context.Context, proposta *models.PropostaMEI) error {
	return s.repo.Update(ctx, proposta)
}

func (s *PropostaMEIService) UpdateProposta(ctx context.Context, id uuid.UUID, oportunidadeID int, valorProposta *float64) error {
	// Buscar proposta existente
	proposta, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if proposta == nil {
		return errors.New("proposta não encontrada")
	}

	// Validar que a proposta pertence à oportunidade especificada
	if proposta.OportunidadeMEIID != oportunidadeID {
		return errors.New("proposta não pertence à oportunidade especificada")
	}

	// Atualizar apenas o valor da proposta
	if valorProposta != nil {
		// Validate that the value is positive
		if *valorProposta < 0 {
			return errors.New("valor_proposta deve ser positivo")
		}
		proposta.ValorProposta = valorProposta
	}

	return s.repo.Update(ctx, proposta)
}

func (s *PropostaMEIService) UpdateStatusCidadao(ctx context.Context, id uuid.UUID, status models.StatusPropostaCidadao) error {
	proposta, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if proposta == nil {
		return errors.New("proposta não encontrada")
	}

	proposta.StatusCidadao = status

	return s.repo.Update(ctx, proposta)
}

func (s *PropostaMEIService) Approve(ctx context.Context, id uuid.UUID) error {
	return s.UpdateStatusCidadao(ctx, id, models.StatusPropostaCidadaoApproved)
}

func (s *PropostaMEIService) Reject(ctx context.Context, id uuid.UUID) error {
	return s.UpdateStatusCidadao(ctx, id, models.StatusPropostaCidadaoRejected)
}

func (s *PropostaMEIService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *PropostaMEIService) ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	offset := (page - 1) * pageSize
	propostas, total, err := s.repo.ListByOportunidade(ctx, oportunidadeID, nomeEmpresa, cnpj, status, pageSize, offset)
	if err != nil {
		return propostas, total, err
	}

	// Enrich all proposals with contact info (parallel)
	s.enrichMultiplePropostasWithContactInfo(ctx, propostas)

	return propostas, total, nil
}

func (s *PropostaMEIService) ListByMEIEmpresa(ctx context.Context, meiEmpresaID string, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByMEIEmpresa(ctx, meiEmpresaID, pageSize, offset)
}

func (s *PropostaMEIService) ListByStatus(ctx context.Context, status models.StatusPropostaCidadao, page, pageSize int) ([]*models.PropostaMEI, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByStatus(ctx, status, pageSize, offset)
}

func (s *PropostaMEIService) UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error) {
	if len(propostaIDs) == 0 {
		return 0, errors.New("nenhuma proposta selecionada")
	}

	if !status.IsValid() {
		return 0, errors.New("status inválido")
	}

	return s.repo.UpdateMultipleStatus(ctx, propostaIDs, status)
}

// enrichPropostaWithContactInfo enriches a single proposta with contact information
// Implements graceful degradation: logs errors but doesn't fail the operation
func (s *PropostaMEIService) enrichPropostaWithContactInfo(ctx context.Context, proposta *models.PropostaMEI) {
	if proposta == nil || proposta.MEIEmpresaID == "" {
		return
	}

	// Skip if contact info service is not configured
	if s.contactInfoService == nil {
		return
	}

	contactInfo, err := s.contactInfoService.GetCNPJOwnerContactInfo(ctx, proposta.MEIEmpresaID)
	if err != nil {
		log.Printf("[CONTACT_INFO_ERROR] Failed to fetch contact info for CNPJ %s: %v",
			proposta.MEIEmpresaID, err)
		return
	}

	// Set contact fields (these are computed fields, not persisted)
	proposta.EmailPessoaFisica = contactInfo.EmailPessoaFisica
	proposta.CelularPessoaFisica = contactInfo.CelularPessoaFisica
}

// enrichMultiplePropostasWithContactInfo enriches multiple propostas in parallel
// Implements graceful degradation: errors don't stop the entire operation
func (s *PropostaMEIService) enrichMultiplePropostasWithContactInfo(ctx context.Context, propostas []*models.PropostaMEI) {
	if len(propostas) == 0 || s.contactInfoService == nil {
		return
	}

	// Extract unique CNPJs
	cnpjSet := make(map[string]bool)
	for _, p := range propostas {
		if p.MEIEmpresaID != "" {
			cnpjSet[p.MEIEmpresaID] = true
		}
	}

	if len(cnpjSet) == 0 {
		return
	}

	// Convert set to slice
	cnpjs := make([]string, 0, len(cnpjSet))
	for cnpj := range cnpjSet {
		cnpjs = append(cnpjs, cnpj)
	}

	// Fetch all contact info in parallel
	contactInfoMap := s.contactInfoService.GetMultipleCNPJOwnerContactInfo(ctx, cnpjs)

	// Enrich each proposta
	for _, proposta := range propostas {
		if contactInfo, found := contactInfoMap[proposta.MEIEmpresaID]; found {
			proposta.EmailPessoaFisica = contactInfo.EmailPessoaFisica
			proposta.CelularPessoaFisica = contactInfo.CelularPessoaFisica
		}
	}
}
