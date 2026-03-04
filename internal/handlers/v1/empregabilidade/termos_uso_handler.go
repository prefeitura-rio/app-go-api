package empregabilidade

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type TermosUsoHandler struct {
	service *services.TermosUsoService
}

func NewTermosUsoHandler(service *services.TermosUsoService) *TermosUsoHandler {
	return &TermosUsoHandler{service: service}
}

func checkTermosOwnership(c *gin.Context) bool {
	cpf := c.Param("cpf")
	if middlewares.IsAdmin(c) {
		return true
	}
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return false
	}
	if userCPF != cpf {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode acessar seus próprios dados"})
		return false
	}
	return true
}

// @Summary      Verificar aceite dos termos
// @Description  Verifica se o usuário aceitou os termos de uso
// @Tags         empregabilidade-termos-uso
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  map[string]bool
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/termos-uso/{cpf} [get]
func (h *TermosUsoHandler) HasAcceptedTerms(c *gin.Context) {
	if !checkTermosOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	hasAccepted, err := h.service.HasAcceptedTerms(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_accepted": hasAccepted})
}

// @Summary      Aceitar termos de uso
// @Description  Registra o aceite dos termos de uso pelo usuário autenticado
// @Tags         empregabilidade-termos-uso
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/termos-uso/{cpf}/accept [put]
func (h *TermosUsoHandler) AcceptTerms(c *gin.Context) {
	if !checkTermosOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	if err := h.service.AcceptTerms(c.Request.Context(), cpf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Termos de uso aceitos com sucesso"})
}

// @Summary      Buscar detalhes dos termos
// @Description  Retorna os detalhes do aceite dos termos de uso
// @Tags         empregabilidade-termos-uso
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  empregabilidade.TermosUso
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/termos-uso/{cpf}/details [get]
func (h *TermosUsoHandler) GetDetails(c *gin.Context) {
	if !checkTermosOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	termos, err := h.service.GetByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if termos == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Termos de uso não encontrados"})
		return
	}

	c.JSON(http.StatusOK, termos)
}
