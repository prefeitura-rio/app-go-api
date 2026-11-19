package v1

import (
	"github.com/gin-gonic/gin"
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

// @Summary      Criar empresa
// @Description  Cria uma nova empresa
// @Tags         empresas
// @Accept       json
// @Produce      json
// @Param        request  body      models.Empresa  true  "Dados da empresa"
// @Success      201      {object}  models.Empresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empresas [post]
func (h *EmpresaHandler) Create(c *gin.Context) {
	h.CRUDHandler.Create(c)
}

// @Summary      Listar empresas
// @Description  Retorna lista paginada de empresas
// @Tags         empresas
// @Produce      json
// @Param        page      query     int  false  "Número da página (default: 1)"
// @Param        pageSize  query     int  false  "Tamanho da página (default: 10)"
// @Success      200       {object}  object{data=[]models.Empresa,meta=object{page=int,page_size=int,total=int}}
// @Failure      500       {object}  models.ErrorResponse
// @Router       /api/v1/empresas [get]
func (h *EmpresaHandler) List(c *gin.Context) {
	h.CRUDHandler.List(c)
}

// @Summary      Buscar empresa
// @Description  Retorna uma empresa pelo ID
// @Tags         empresas
// @Produce      json
// @Param        id   path      int  true  "ID da empresa"
// @Success      200  {object}  models.Empresa
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [get]
func (h *EmpresaHandler) GetByID(c *gin.Context) {
	h.CRUDHandler.GetByID(c)
}

// @Summary      Atualizar empresa
// @Description  Atualiza uma empresa existente
// @Tags         empresas
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "ID da empresa"
// @Param        request  body      models.Empresa  true  "Dados da empresa"
// @Success      200      {object}  models.Empresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [put]
func (h *EmpresaHandler) Update(c *gin.Context) {
	h.CRUDHandler.Update(c)
}

// @Summary      Excluir empresa
// @Description  Remove uma empresa pelo ID
// @Tags         empresas
// @Produce      json
// @Param        id   path      int  true  "ID da empresa"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [delete]
func (h *EmpresaHandler) Delete(c *gin.Context) {
	h.CRUDHandler.Delete(c)
}
