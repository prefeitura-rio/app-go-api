package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type EmpresaHandler struct {
	*common.CRUDHandler[models.Empresa]
}

func NewEmpresaHandler(service *services.EmpresaService) *EmpresaHandler {
	return &EmpresaHandler{
		CRUDHandler: common.NewCRUDHandler[models.Empresa](service, "Empresa"),
	}
}
