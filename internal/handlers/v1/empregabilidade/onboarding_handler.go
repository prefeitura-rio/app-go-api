package empregabilidade

import (
	"net/http"

	"github.com/gin-gonic/gin"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type OnboardingHandler struct {
	service *services.OnboardingService
}

func NewOnboardingHandler(service *services.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{service: service}
}

// @Summary      Verificar primeiro login
// @Description  Verifica se é o primeiro login do usuário no módulo de empregabilidade
// @Tags         empregabilidade-onboarding
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  map[string]bool
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/onboarding/{cpf} [get]
func (h *OnboardingHandler) IsFirstLogin(c *gin.Context) {
	cpf := c.Param("cpf")

	isFirstLogin, err := h.service.IsFirstLogin(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_first_login": isFirstLogin})
}

// @Summary      Marcar primeiro login como concluído
// @Description  Marca o primeiro login do usuário como concluído
// @Tags         empregabilidade-onboarding
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/onboarding/{cpf}/complete [put]
func (h *OnboardingHandler) MarkFirstLoginCompleted(c *gin.Context) {
	cpf := c.Param("cpf")

	if err := h.service.MarkFirstLoginCompleted(c.Request.Context(), cpf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Primeiro login marcado como concluído"})
}
