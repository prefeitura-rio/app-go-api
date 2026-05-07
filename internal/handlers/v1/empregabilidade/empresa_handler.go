package empregabilidade

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type EmpresaHandler struct {
	service              *services.EmpresaService
	cnpjConsultaService  *services.CNPJConsultaService
}

func NewEmpresaHandler(service *services.EmpresaService) *EmpresaHandler {
	return &EmpresaHandler{service: service}
}

func (h *EmpresaHandler) WithCNPJConsulta(cnpjConsultaService *services.CNPJConsultaService) *EmpresaHandler {
	h.cnpjConsultaService = cnpjConsultaService
	return h
}

// @Summary      Criar empresa
// @Description  Cria uma nova empresa
// @Tags         empregabilidade-empresas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Empresa  true  "Dados da empresa"
// @Success      201      {object}  empregabilidade.Empresa
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas [post]
func (h *EmpresaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Empresa
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cnpj, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.CNPJ = cnpj
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar empresas
// @Description  Retorna lista paginada de empresas com filtros opcionais
// @Tags         empregabilidade-empresas
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10)"
// @Param        search    query     string  false  "Busca parcial em razão social ou nome fantasia"
// @Param        cnpj      query     string  false  "Filtrar por CNPJ exato"
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas [get]
func (h *EmpresaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := empregabilidade.EmpresaFilter{
		Search: c.Query("search"),
		CNPJ:   c.Query("cnpj"),
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Buscar empresa
// @Description  Retorna uma empresa pelo CNPJ
// @Tags         empregabilidade-empresas
// @Produce      json
// @Param        cnpj   path      string  true  "CNPJ da empresa"
// @Success      200    {object}  empregabilidade.Empresa
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas/{cnpj} [get]
func (h *EmpresaHandler) GetByID(c *gin.Context) {
	cnpj := c.Param("cnpj")

	entity, err := h.service.GetByID(c.Request.Context(), cnpj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar empresa
// @Description  Atualiza uma empresa existente
// @Tags         empregabilidade-empresas
// @Accept       json
// @Produce      json
// @Param        cnpj     path      string                   true  "CNPJ da empresa"
// @Param        request  body      empregabilidade.Empresa  true  "Dados da empresa"
// @Success      200      {object}  empregabilidade.Empresa
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas/{cnpj} [put]
func (h *EmpresaHandler) Update(c *gin.Context) {
	cnpj := c.Param("cnpj")

	var entity empregabilidade.Empresa
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.CNPJ = cnpj

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir empresa
// @Description  Remove uma empresa pelo CNPJ
// @Tags         empregabilidade-empresas
// @Produce      json
// @Param        cnpj   path      string  true  "CNPJ da empresa"
// @Success      200    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas/{cnpj} [delete]
func (h *EmpresaHandler) Delete(c *gin.Context) {
	cnpj := c.Param("cnpj")

	if err := h.service.Delete(c.Request.Context(), cnpj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Empresa excluída com sucesso"})
}

// @Summary      Consultar CNPJ
// @Description  Consulta dados de uma empresa pelo CNPJ no RMI (Registro Mercantil Integrado)
// @Tags         empregabilidade-empresas
// @Produce      json
// @Param        cnpj   path      string  true  "CNPJ da empresa (apenas dígitos, 14 caracteres)"
// @Success      200    {object}  models.LegalEntityConsultaResponse
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /api/v1/empregabilidade/empresas/consulta-cnpj/{cnpj} [get]
func (h *EmpresaHandler) ConsultaCNPJ(c *gin.Context) {
	if h.cnpjConsultaService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Serviço de consulta CNPJ não disponível"})
		return
	}

	cnpj := c.Param("cnpj")
	if cnpj == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ é obrigatório"})
		return
	}

	result, err := h.cnpjConsultaService.ConsultarCNPJ(c.Request.Context(), cnpj)
	if err != nil {
		if errors.Is(err, clients.ErrCNPJNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "CNPJ não encontrado"})
			return
		}
		if errors.Is(err, clients.ErrCNPJAccessDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado ao CNPJ"})
			return
		}
		if errors.Is(err, clients.ErrInvalidCNPJ) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ inválido: deve conter 14 dígitos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
