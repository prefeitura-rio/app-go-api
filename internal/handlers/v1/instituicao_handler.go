package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type InstituicaoHandler struct {
	*common.CRUDHandler[models.InstituicaoEnsino]
}

func NewInstituicaoHandler(service *services.InstituicaoService) *InstituicaoHandler {
	return &InstituicaoHandler{
		CRUDHandler: common.NewCRUDHandler[models.InstituicaoEnsino](service, "Instituição"),
	}
}
