package empregabilidade

import (
	"context"
	"fmt"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type CNPJConsultaService struct {
	rmiClient    *clients.RMIClient
	tokenManager *auth.ServiceAccountTokenManager
}

func NewCNPJConsultaService(
	rmiClient *clients.RMIClient,
	tokenManager *auth.ServiceAccountTokenManager,
) *CNPJConsultaService {
	return &CNPJConsultaService{
		rmiClient:    rmiClient,
		tokenManager: tokenManager,
	}
}

// ConsultarCNPJ fetches legal entity data from RMI API by CNPJ
// Returns a simplified response for frontend consumption
func (s *CNPJConsultaService) ConsultarCNPJ(ctx context.Context, cnpj string) (*models.LegalEntityConsultaResponse, error) {
	if s.tokenManager == nil {
		return nil, fmt.Errorf("service account não configurado")
	}

	token, err := s.tokenManager.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter token de serviço: %w", err)
	}

	legalEntity, err := s.rmiClient.GetLegalEntityByCNPJ(ctx, token, cnpj)
	if err != nil {
		return nil, err
	}

	return legalEntity.ToConsultaResponse(), nil
}
