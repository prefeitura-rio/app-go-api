package v1

import (
	"github.com/gin-gonic/gin"
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

// @Summary      Criar instituição de ensino
// @Description  Cria uma nova instituição de ensino
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Param        request  body      models.InstituicaoEnsino  true  "Dados da instituição de ensino"
// @Success      201      {object}  models.InstituicaoEnsino
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes [post]
func (h *InstituicaoHandler) Create(c *gin.Context) {
	h.CRUDHandler.Create(c)
}

// @Summary      Listar instituições de ensino
// @Description  Retorna lista paginada de instituições de ensino
// @Tags         instituicoes
// @Produce      json
// @Param        page      query     int  false  "Número da página (default: 1)"
// @Param        pageSize  query     int  false  "Tamanho da página (default: 10)"
// @Success      200       {object}  object{data=[]models.InstituicaoEnsino,meta=object{page=int,page_size=int,total=int}}
// @Failure      500       {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes [get]
func (h *InstituicaoHandler) List(c *gin.Context) {
	h.CRUDHandler.List(c)
}

// @Summary      Buscar instituição de ensino
// @Description  Retorna uma instituição de ensino pelo ID
// @Tags         instituicoes
// @Produce      json
// @Param        id   path      int  true  "ID da instituição de ensino"
// @Success      200  {object}  models.InstituicaoEnsino
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [get]
func (h *InstituicaoHandler) GetByID(c *gin.Context) {
	h.CRUDHandler.GetByID(c)
}

// @Summary      Atualizar instituição de ensino
// @Description  Atualiza uma instituição de ensino existente
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "ID da instituição de ensino"
// @Param        request  body      models.InstituicaoEnsino  true  "Dados da instituição de ensino"
// @Success      200      {object}  models.InstituicaoEnsino
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [put]
func (h *InstituicaoHandler) Update(c *gin.Context) {
	h.CRUDHandler.Update(c)
}

// @Summary      Excluir instituição de ensino
// @Description  Remove uma instituição de ensino pelo ID
// @Tags         instituicoes
// @Produce      json
// @Param        id   path      int  true  "ID da instituição de ensino"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [delete]
func (h *InstituicaoHandler) Delete(c *gin.Context) {
	h.CRUDHandler.Delete(c)
}
