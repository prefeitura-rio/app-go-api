package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type MEIEmpresaService struct {
	repo     *repository.MEIEmpresaRepository
	cnaeRepo *repository.CNAERepository
}

func NewMEIEmpresaService(repo *repository.MEIEmpresaRepository, cnaeRepo *repository.CNAERepository) *MEIEmpresaService {
	return &MEIEmpresaService{
		repo:     repo,
		cnaeRepo: cnaeRepo,
	}
}

func (s *MEIEmpresaService) Create(ctx context.Context, meiEmpresa *models.MEIEmpresa) (int, error) {
	// Validar campos obrigatórios
	if err := meiEmpresa.Validate(); err != nil {
		return 0, err
	}

	if err := s.ValidateCNPJ(meiEmpresa.CNPJ); err != nil {
		return 0, err
	}

	// Verificar se CNPJ já existe
	existing, err := s.repo.GetByCNPJ(ctx, meiEmpresa.CNPJ)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, errors.New("CNPJ já cadastrado")
	}

	// Validar que todos os CNAEs existem
	for _, cnae := range meiEmpresa.CNAEs {
		existingCNAE, err := s.cnaeRepo.GetByID(ctx, cnae.ID)
		if err != nil {
			return 0, err
		}
		if existingCNAE == nil {
			return 0, fmt.Errorf("CNAE ID %d não encontrado", cnae.ID)
		}
	}

	return s.repo.Create(ctx, meiEmpresa)
}

func (s *MEIEmpresaService) GetByID(ctx context.Context, id int) (*models.MEIEmpresa, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MEIEmpresaService) GetByCNPJ(ctx context.Context, cnpj string) (*models.MEIEmpresa, error) {
	return s.repo.GetByCNPJ(ctx, cnpj)
}

func (s *MEIEmpresaService) Update(ctx context.Context, meiEmpresa *models.MEIEmpresa) error {
	// Validar campos obrigatórios
	if err := meiEmpresa.Validate(); err != nil {
		return err
	}

	if err := s.ValidateCNPJ(meiEmpresa.CNPJ); err != nil {
		return err
	}

	// Validar que todos os CNAEs existem
	for _, cnae := range meiEmpresa.CNAEs {
		existingCNAE, err := s.cnaeRepo.GetByID(ctx, cnae.ID)
		if err != nil {
			return err
		}
		if existingCNAE == nil {
			return fmt.Errorf("CNAE ID %d não encontrado", cnae.ID)
		}
	}

	return s.repo.Update(ctx, meiEmpresa)
}

func (s *MEIEmpresaService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *MEIEmpresaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.MEIEmpresa, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

func (s *MEIEmpresaService) ValidateCNPJ(cnpj string) error {
	// Remove caracteres não numéricos
	re := regexp.MustCompile(`[^\d]`)
	cleanCNPJ := re.ReplaceAllString(cnpj, "")

	// Verifica se tem 14 dígitos
	if len(cleanCNPJ) != 14 {
		return errors.New("CNPJ deve conter 14 dígitos")
	}

	// Validação básica - pode ser expandida com algoritmo completo de validação
	return nil
}
