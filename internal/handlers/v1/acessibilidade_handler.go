package v1

import (
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type AcessibilidadeHandler struct {
	*common.CRUDHandler[models.Acessibilidade]
}

func NewAcessibilidadeHandler(service *services.AcessibilidadeService) *AcessibilidadeHandler {
	return &AcessibilidadeHandler{
		CRUDHandler: common.NewCRUDHandler[models.Acessibilidade](service, "Acessibilidade"),
	}
}

// WithCache adds cache support to the handler
func (h *AcessibilidadeHandler) WithCache(cache *cache.ReferenceDataCache) *AcessibilidadeHandler {
	h.CRUDHandler.WithCache(cache)
	return h
}
