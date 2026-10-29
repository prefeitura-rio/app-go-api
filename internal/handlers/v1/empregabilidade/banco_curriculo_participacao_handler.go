package empregabilidade

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ParticipacaoBancoServiceInterface é o que o handler usa do ParticipacaoBancoService.
type ParticipacaoBancoServiceInterface interface {
	Get(ctx context.Context, cpf string) (*empregabilidade.ParticipacaoBancoStatus, error)
	Atualizar(ctx context.Context, cpf string, disponivel bool) (*empregabilidade.ParticipacaoBancoStatus, error)
}

// ParticipacaoBancoHandler atende o card do banco de currículos no app do cidadão.
type ParticipacaoBancoHandler struct {
	service ParticipacaoBancoServiceInterface
}

func NewParticipacaoBancoHandler(service ParticipacaoBancoServiceInterface) *ParticipacaoBancoHandler {
	return &ParticipacaoBancoHandler{service: service}
}

// @Summary      Consultar participação no banco de currículos
// @Description  Diz se o currículo do cidadão está disponível para os recrutadores no banco de currículos, o que falta para o currículo mínimo (o mesmo exigido para se candidatar) e quando o currículo foi atualizado. Só o próprio cidadão ou um admin.
// @Tags         empregabilidade-curriculo-banco
// @Produce      json
// @Param        cpf  path      string  true  "CPF do cidadão"
// @Success      200  {object}  empregabilidade.ParticipacaoBancoStatus
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/banco-curriculos [get]
func (h *ParticipacaoBancoHandler) Get(c *gin.Context) {
	if !checkTermosOwnership(c) {
		return
	}

	status, err := h.service.Get(c.Request.Context(), cpfSomenteDigitos(c.Param("cpf")))
	if err != nil {
		log.Printf("[ParticipacaoBancoHandler] Get failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar o banco de currículos"})
		return
	}
	c.JSON(http.StatusOK, status)
}

// @Summary      Ativar ou desativar o currículo no banco de currículos
// @Description  O cidadão escolhe se os recrutadores podem consultar o currículo no banco de currículos. Vale para qualquer currículo, completo ou não (422 quando a pessoa ainda não salvou nenhuma parte). Só o próprio cidadão.
// @Tags         empregabilidade-curriculo-banco
// @Accept       json
// @Produce      json
// @Param        cpf      path      string                                             true  "CPF do cidadão"
// @Param        request  body      empregabilidade.AtualizarParticipacaoBancoRequest  true  "Disponível ou não"
// @Success      200      {object}  empregabilidade.ParticipacaoBancoStatus
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      422      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/banco-curriculos [put]
func (h *ParticipacaoBancoHandler) Atualizar(c *gin.Context) {
	cpf := cpfSomenteDigitos(c.Param("cpf"))
	// A escolha é do cidadão: nem o admin muda pelo outro.
	userCPF := cpfSomenteDigitos(middlewares.GetUserCPF(c))
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}
	if userCPF != cpf {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode alterar o seu próprio currículo"})
		return
	}

	var req empregabilidade.AtualizarParticipacaoBancoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Informe se o currículo fica disponível"})
		return
	}

	status, err := h.service.Atualizar(c.Request.Context(), cpf, *req.Disponivel)
	if err != nil {
		if errors.Is(err, services.ErrSemCurriculoParaBanco) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Crie o currículo antes de escolher se ele aparece no banco de currículos"})
			return
		}
		log.Printf("[ParticipacaoBancoHandler] Atualizar failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar o banco de currículos"})
		return
	}
	c.JSON(http.StatusOK, status)
}
