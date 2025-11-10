package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type EscolaridadeHandler struct {
	*common.CRUDHandler[models.Escolaridade]
}

func NewEscolaridadeHandler(service *services.EscolaridadeService) *EscolaridadeHandler {
	return &EscolaridadeHandler{
		CRUDHandler: common.NewCRUDHandler[models.Escolaridade](service, "Escolaridade"),
	}
}
