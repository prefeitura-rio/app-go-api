package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
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

// WithCache adds cache support to the handler
func (h *EscolaridadeHandler) WithCache(cache *cache.ReferenceDataCache) *EscolaridadeHandler {
	h.CRUDHandler.WithCache(cache)
	return h
}

// @Summary      Criar escolaridade
// @Description  Cria um novo nível de escolaridade
// @Tags         escolaridades
// @Accept       json
// @Produce      json
// @Param        request  body      models.Escolaridade  true  "Dados da escolaridade"
// @Success      201      {object}  models.Escolaridade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades [post]
func (h *EscolaridadeHandler) Create(c *gin.Context) {
	h.CRUDHandler.Create(c)
}

// @Summary      Listar escolaridades
// @Description  Retorna lista paginada de níveis de escolaridade
// @Tags         escolaridades
// @Produce      json
// @Param        page      query     int  false  "Número da página (default: 1)"
// @Param        pageSize  query     int  false  "Tamanho da página (default: 10)"
// @Success      200       {object}  object{data=[]models.Escolaridade,meta=object{page=int,page_size=int,total=int}}
// @Failure      500       {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades [get]
func (h *EscolaridadeHandler) List(c *gin.Context) {
	h.CRUDHandler.List(c)
}

// @Summary      Buscar escolaridade
// @Description  Retorna uma escolaridade pelo ID
// @Tags         escolaridades
// @Produce      json
// @Param        id   path      int  true  "ID da escolaridade"
// @Success      200  {object}  models.Escolaridade
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [get]
func (h *EscolaridadeHandler) GetByID(c *gin.Context) {
	h.CRUDHandler.GetByID(c)
}

// @Summary      Atualizar escolaridade
// @Description  Atualiza uma escolaridade existente
// @Tags         escolaridades
// @Accept       json
// @Produce      json
// @Param        id       path      int                  true  "ID da escolaridade"
// @Param        request  body      models.Escolaridade  true  "Dados da escolaridade"
// @Success      200      {object}  models.Escolaridade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [put]
func (h *EscolaridadeHandler) Update(c *gin.Context) {
	h.CRUDHandler.Update(c)
}

// @Summary      Excluir escolaridade
// @Description  Remove uma escolaridade pelo ID
// @Tags         escolaridades
// @Produce      json
// @Param        id   path      int  true  "ID da escolaridade"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [delete]
func (h *EscolaridadeHandler) Delete(c *gin.Context) {
	h.CRUDHandler.Delete(c)
}
