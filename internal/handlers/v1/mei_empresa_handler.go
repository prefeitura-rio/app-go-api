package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type MEIEmpresaHandler struct {
	service *services.MEIEmpresaService
}

func NewMEIEmpresaHandler(service *services.MEIEmpresaService) *MEIEmpresaHandler {
	return &MEIEmpresaHandler{
		service: service,
	}
}

// @Summary      Registrar MEI empresa
// @Description  Cria um novo registro de MEI empresa
// @Tags         mei-empresas
// @Accept       json
// @Produce      json
// @Param        request  body      models.MEIEmpresa  true  "Dados da MEI empresa"
// @Success      201      {object}  models.MEIEmpresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/mei-empresas [post]
func (h *MEIEmpresaHandler) Create(c *gin.Context) {
	var meiEmpresa models.MEIEmpresa
	if err := c.ShouldBindJSON(&meiEmpresa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &meiEmpresa)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meiEmpresa.ID = id
	c.JSON(http.StatusCreated, meiEmpresa)
}

// @Summary      Obter MEI empresa por ID
// @Description  Retorna uma MEI empresa pelo seu ID
// @Tags         mei-empresas
// @Produce      json
// @Param        id   path      int  true  "ID da MEI empresa"
// @Success      200  {object}  models.MEIEmpresa
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/mei-empresas/{id} [get]
func (h *MEIEmpresaHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	meiEmpresa, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar MEI empresa: " + err.Error()})
		return
	}

	if meiEmpresa == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MEI empresa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, meiEmpresa)
}

// @Summary      Buscar MEI empresa por CNPJ
// @Description  Retorna uma MEI empresa pelo seu CNPJ
// @Tags         mei-empresas
// @Produce      json
// @Param        cnpj   path      string  true  "CNPJ da MEI empresa"
// @Success      200  {object}  models.MEIEmpresa
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/mei-empresas/cnpj/{cnpj} [get]
func (h *MEIEmpresaHandler) GetByCNPJ(c *gin.Context) {
	cnpj := c.Param("cnpj")

	meiEmpresa, err := h.service.GetByCNPJ(c.Request.Context(), cnpj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar MEI empresa: " + err.Error()})
		return
	}

	if meiEmpresa == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MEI empresa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, meiEmpresa)
}

// @Summary      Atualizar MEI empresa
// @Description  Atualiza os dados de uma MEI empresa existente
// @Tags         mei-empresas
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "ID da MEI empresa"
// @Param        request  body      models.MEIEmpresa  true  "Dados da MEI empresa"
// @Success      200      {object}  models.MEIEmpresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/mei-empresas/{id} [put]
func (h *MEIEmpresaHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var meiEmpresa models.MEIEmpresa
	if err := c.ShouldBindJSON(&meiEmpresa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	meiEmpresa.ID = id
	if err := h.service.Update(c.Request.Context(), &meiEmpresa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, meiEmpresa)
}

// @Summary      Listar MEI empresas
// @Description  Retorna uma lista paginada de MEI empresas
// @Tags         mei-empresas
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10, max: 1000)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/mei-empresas [get]
func (h *MEIEmpresaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := make(map[string]interface{})

	meiEmpresas, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar MEI empresas: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": meiEmpresas,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}
