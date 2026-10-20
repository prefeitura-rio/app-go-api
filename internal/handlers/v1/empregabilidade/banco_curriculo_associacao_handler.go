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

// AssociacaoCandidatoServiceInterface é o que o handler usa do AssociacaoCandidatoService.
type AssociacaoCandidatoServiceInterface interface {
	Associar(ctx context.Context, params services.AssociarParams) ([]*empregabilidade.AssociacaoVaga, error)
}

type AssociacaoCandidatoHandler struct {
	service AssociacaoCandidatoServiceInterface
}

func NewAssociacaoCandidatoHandler(service AssociacaoCandidatoServiceInterface) *AssociacaoCandidatoHandler {
	return &AssociacaoCandidatoHandler{service: service}
}

// @Summary      Associar candidato a vagas
// @Description  Associa o candidato do banco de currículos a uma ou mais vagas, criando a candidatura em cada uma. Valem as regras da candidatura do app: currículo mínimo e termos aceitos (422 quando faltam), vaga recebendo candidaturas e requisitos da vaga. O resultado vem por vaga. Exige acesso ao banco de currículos (admin ou go:curriculos:*) e, fora os admins totais, go:empregabilidade:admin ou editor de empregabilidade; editores só associam a vagas da própria secretaria. O candidato não recebe o e-mail de candidatura enviada.
// @Tags         empregabilidade-banco-curriculos
// @Accept       json
// @Produce      json
// @Param        cpf      path      string                                true  "CPF do cidadão, com ou sem máscara"
// @Param        request  body      empregabilidade.AssociarVagasRequest  true  "Vagas a associar (até 50)"
// @Success      200      {object}  empregabilidade.AssociarVagasResponse
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      422      {object}  empregabilidade.AssociacaoRecusadaResponse
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/banco-curriculos/{cpf}/associacoes [post]
func (h *AssociacaoCandidatoHandler) Associar(c *gin.Context) {
	cpf := cpfSomenteDigitos(c.Param("cpf"))
	if len(cpf) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CPF inválido"})
		return
	}

	var req empregabilidade.AssociarVagasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Informe de 1 a 50 vagas"})
		return
	}

	resultados, err := h.service.Associar(c.Request.Context(), services.AssociarParams{
		CPF:              cpf,
		IDsVagas:         req.IDsVagas,
		OperadorCPF:      middlewares.GetUserCPF(c),
		OrgaosPermitidos: middlewares.GetBancoCurriculosOrgaosPermitidos(c),
	})
	if err != nil {
		var naoAssociavel *services.CandidatoNaoAssociavelError
		if errors.As(err, &naoAssociavel) {
			c.JSON(http.StatusUnprocessableEntity, empregabilidade.AssociacaoRecusadaResponse{
				Error:      "O currículo não tem o mínimo para se candidatar",
				Pendencias: naoAssociavel.Pendencias,
			})
			return
		}
		log.Printf("[AssociacaoCandidatoHandler] Associar failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao associar candidato"})
		return
	}
	if resultados == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Currículo não encontrado"})
		return
	}

	c.JSON(http.StatusOK, empregabilidade.AssociarVagasResponse{Resultados: resultados})
}
