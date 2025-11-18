package v1

import (
	"github.com/gin-gonic/gin"
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

// @Summary      Criar acessibilidade
// @Description  Cria uma nova opção de acessibilidade
// @Tags         acessibilidades
// @Accept       json
// @Produce      json
// @Param        request  body      models.Acessibilidade  true  "Dados da acessibilidade"
// @Success      201      {object}  models.Acessibilidade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades [post]
func (h *AcessibilidadeHandler) Create(c *gin.Context) {
	h.CRUDHandler.Create(c)
}

// @Summary      Listar acessibilidades
// @Description  Retorna lista paginada de opções de acessibilidade
// @Tags         acessibilidades
// @Produce      json
// @Param        page      query     int  false  "Número da página (default: 1)"
// @Param        pageSize  query     int  false  "Tamanho da página (default: 10)"
// @Success      200       {object}  object{data=[]models.Acessibilidade,meta=object{page=int,page_size=int,total=int}}
// @Failure      500       {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades [get]
func (h *AcessibilidadeHandler) List(c *gin.Context) {
	h.CRUDHandler.List(c)
}

// @Summary      Buscar acessibilidade
// @Description  Retorna uma acessibilidade pelo ID
// @Tags         acessibilidades
// @Produce      json
// @Param        id   path      int  true  "ID da acessibilidade"
// @Success      200  {object}  models.Acessibilidade
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [get]
func (h *AcessibilidadeHandler) GetByID(c *gin.Context) {
	h.CRUDHandler.GetByID(c)
}

// @Summary      Atualizar acessibilidade
// @Description  Atualiza uma acessibilidade existente
// @Tags         acessibilidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "ID da acessibilidade"
// @Param        request  body      models.Acessibilidade  true  "Dados da acessibilidade"
// @Success      200      {object}  models.Acessibilidade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [put]
func (h *AcessibilidadeHandler) Update(c *gin.Context) {
	h.CRUDHandler.Update(c)
}

// @Summary      Excluir acessibilidade
// @Description  Remove uma acessibilidade pelo ID
// @Tags         acessibilidades
// @Produce      json
// @Param        id   path      int  true  "ID da acessibilidade"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [delete]
func (h *AcessibilidadeHandler) Delete(c *gin.Context) {
	h.CRUDHandler.Delete(c)
}
