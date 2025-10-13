package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type PropostaMEIService struct {
	repo              *repository.PropostaMEIRepository
	oportunidadeRepo  *repository.OportunidadeMEIRepository
	meiEmpresaRepo    *repository.MEIEmpresaRepository
}

func NewPropostaMEIService(
	repo *repository.PropostaMEIRepository,
	oportunidadeRepo *repository.OportunidadeMEIRepository,
	meiEmpresaRepo *repository.MEIEmpresaRepository,
) *PropostaMEIService {
	return &PropostaMEIService{
		repo:             repo,
		oportunidadeRepo: oportunidadeRepo,
		meiEmpresaRepo:   meiEmpresaRepo,
	}
}

func (s *PropostaMEIService) Create(ctx context.Context, proposta *models.PropostaMEI) (uuid.UUID, error) {
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

	// Validar que a MEI empresa existe
	meiEmpresa, err := s.meiEmpresaRepo.GetByID(ctx, proposta.MEIEmpresaID)
	if err != nil {
		return uuid.Nil, err
	}
	if meiEmpresa == nil {
		return uuid.Nil, errors.New("MEI empresa não encontrada")
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
	return s.repo.GetByID(ctx, id)
}

func (s *PropostaMEIService) Update(ctx context.Context, proposta *models.PropostaMEI) error {
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
	return s.repo.ListByOportunidade(ctx, oportunidadeID, nomeEmpresa, cnpj, status, pageSize, offset)
}

func (s *PropostaMEIService) ListByMEIEmpresa(ctx context.Context, meiEmpresaID int, page, pageSize int) ([]*models.PropostaMEI, int, error) {
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
