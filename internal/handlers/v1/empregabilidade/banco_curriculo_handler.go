package empregabilidade

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// BancoCurriculoServiceInterface é o que o handler usa do BancoCurriculoService.
type BancoCurriculoServiceInterface interface {
	List(ctx context.Context, filter empregabilidade.BancoCurriculoFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoItem, int64, error)
	GetDetalhe(ctx context.Context, cpf string) (*empregabilidade.BancoCurriculoDetalhe, error)
	ListCandidaturas(ctx context.Context, cpf string, filter empregabilidade.BancoCurriculoCandidaturaFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoCandidatura, int64, error)
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
// @Param        genero               query  []string  false  "Gêneros autodeclarados (repetível ou separado por vírgula)"  collectionFormat(multi)
// @Param        idade_min            query  int       false  "Idade mínima"
// @Param        idade_max            query  int       false  "Idade máxima"
// @Param        tempo_experiencia    query  string    false  "Soma das experiências"  Enums(sem_experiencia, ate_2_anos, 2_a_4_anos, 4_a_6_anos, 6_a_8_anos, 8_a_10_anos, mais_de_10_anos)
// @Param        experiencia_carteira query  string    false  "Experiência comprovada em carteira"  Enums(sim, 6_meses)
// @Param        escolaridade         query  string    false  "Escolaridade do cadastro (exata)"
// @Param        id_idioma            query  string    false  "Idioma do currículo"
// @Param        id_nivel_idioma      query  string    false  "Nível mínimo do idioma (exige id_idioma)"
// @Param        bairro               query  string    false  "Bairro do cadastro"
// @Param        pcd                  query  string    false  "Pessoa com deficiência"  Enums(sim, nao)
// @Success      200       {object}  empregabilidade.BancoCurriculoListResponse
// @Failure      400       {object}  map[string]string
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

	candidato, err := lerFiltroCandidato(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	filter := empregabilidade.BancoCurriculoFilter{Search: c.Query("search"), Candidato: candidato}
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

// @Summary      Histórico de candidaturas do currículo
// @Description  Lista as vagas a que o cidadão se candidatou ou foi associado, da candidatura mais recente para a mais antiga. O status da vaga considera a data limite. Exige admin ou role go:curriculos:admin / go:curriculos:editor.
// @Tags         empregabilidade-banco-curriculos
// @Produce      json
// @Param        cpf       path      string  true   "CPF do cidadão, com ou sem máscara"
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10, máximo: 100)"
// @Param        search    query     string  false  "Busca parcial pelo título da vaga"
// @Success      200       {object}  empregabilidade.BancoCurriculoCandidaturaListResponse
// @Failure      400       {object}  map[string]string
// @Failure      403       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/banco-curriculos/{cpf}/candidaturas [get]
func (h *BancoCurriculoHandler) ListCandidaturas(c *gin.Context) {
	cpf := cpfSomenteDigitos(c.Param("cpf"))
	if len(cpf) != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CPF inválido"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := empregabilidade.BancoCurriculoCandidaturaFilter{Search: c.Query("search")}
	items, total, err := h.service.ListCandidaturas(c.Request.Context(), cpf, filter, page, pageSize)
	if err != nil {
		log.Printf("[BancoCurriculoHandler] ListCandidaturas failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar candidaturas"})
		return
	}

	c.JSON(http.StatusOK, empregabilidade.BancoCurriculoCandidaturaListResponse{
		Data: items,
		Meta: empregabilidade.BancoCurriculoListMeta{Page: page, PageSize: pageSize, Total: total},
	})
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
