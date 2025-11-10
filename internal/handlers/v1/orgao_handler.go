package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type OrgaoHandler struct {
	*common.CRUDHandler[models.Orgao]
}

func NewOrgaoHandler(service *services.OrgaoService) *OrgaoHandler {
	return &OrgaoHandler{
		CRUDHandler: common.NewCRUDHandler[models.Orgao](service, "Órgão"),
	}
}
