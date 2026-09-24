package empregabilidade

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// BancoCurriculoServiceInterface é o que o handler usa do BancoCurriculoService.
type BancoCurriculoServiceInterface interface {
	List(ctx context.Context, filter empregabilidade.BancoCurriculoFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoItem, int64, error)
	GetDetalhe(ctx context.Context, cpf string) (*empregabilidade.BancoCurriculoDetalhe, error)
}

type BancoCurriculoHandler struct {
	service BancoCurriculoServiceInterface
}

func NewBancoCurriculoHandler(service BancoCurriculoServiceInterface) *BancoCurriculoHandler {
	return &BancoCurriculoHandler{service: service}
}

// @Summary      Listar banco de currículos
// @Description  Lista todos os currículos da base, do mais recente para o mais antigo. Exige admin ou role go:curriculos:admin / go:curriculos:editor.
// @Tags         empregabilidade-banco-curriculos
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10, máximo: 100)"
// @Param        search    query     string  false  "Busca parcial por nome, nome social ou CPF"
// @Success      200       {object}  empregabilidade.BancoCurriculoListResponse
// @Failure      403       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/banco-curriculos [get]
func (h *BancoCurriculoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := empregabilidade.BancoCurriculoFilter{Search: c.Query("search")}
	items, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		log.Printf("[BancoCurriculoHandler] List failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar currículos"})
		return
	}

	c.JSON(http.StatusOK, empregabilidade.BancoCurriculoListResponse{
		Data: items,
		Meta: empregabilidade.BancoCurriculoListMeta{Page: page, PageSize: pageSize, Total: total},
	})
}

// @Summary      Detalhar currículo do banco
// @Description  Retorna a ficha de um currículo: dados pessoais do cadastro do cidadão e o currículo completo. Exige admin ou role go:curriculos:admin / go:curriculos:editor.
// @Tags         empregabilidade-banco-curriculos
// @Produce      json
// @Param        cpf   path      string  true  "CPF do cidadão, com ou sem máscara"
// @Success      200   {object}  empregabilidade.BancoCurriculoDetalhe
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/banco-curriculos/{cpf} [get]
func (h *BancoCurriculoHandler) GetByCPF(c *gin.Context) {
	cpf := cpfSomenteDigitos(c.Param("cpf"))
	if len(cpf) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CPF inválido"})
		return
	}

	detalhe, err := h.service.GetDetalhe(c.Request.Context(), cpf)
	if err != nil {
		log.Printf("[BancoCurriculoHandler] GetByCPF failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar currículo"})
		return
	}
	if detalhe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Currículo não encontrado"})
		return
	}

	c.JSON(http.StatusOK, detalhe)
}

func cpfSomenteDigitos(valor string) string {
	var b strings.Builder
	for _, r := range valor {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
