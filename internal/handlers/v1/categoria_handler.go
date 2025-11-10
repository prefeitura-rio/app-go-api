package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type CategoriaHandler struct {
	*common.CRUDHandler[models.Categoria]
}

func NewCategoriaHandler(service *services.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{
		CRUDHandler: common.NewCRUDHandler[models.Categoria](service, "Categoria"),
	}
}
