package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TypesenseHandler contém os handlers para operações de busca no Typesense
type TypesenseHandler struct {
	service *services.TypesenseService
}

// NewTypesenseHandler cria uma nova instância do TypesenseHandler
func NewTypesenseHandler() (*TypesenseHandler, error) {
	service, err := services.NewTypesenseService()
	if err != nil {
		return nil, err
	}
	return &TypesenseHandler{
		service: service,
	}, nil
}

// handleError é uma função utilitária para manipular erros em respostas HTTP
func handleError(c *gin.Context, status int, message string, err error) {
	response := models.ErrorResponse{
		Message: message,
		Code:    status,
	}
	if err != nil {
		response.Message = message + ": " + err.Error()
	}
	c.JSON(status, response)
}

// @Summary      Buscar em múltiplas coleções
// @Description  Realiza busca em várias coleções simultaneamente
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        params  body  models.MultiCollectionSearchParameters  true  "Parâmetros de busca"
// @Success      200     {object}  models.MultiCollectionSearchResponse
// @Failure      400     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /api/v1/typesense/multi-search [post]
func (h *TypesenseHandler) SearchMultiCollection(c *gin.Context) {
	var params models.MultiCollectionSearchParameters
	if err := c.ShouldBindJSON(&params); err != nil {
		handleError(c, http.StatusBadRequest, "Parâmetros de busca inválidos", err)
		return
	}

	// Validar se pelo menos uma coleção foi especificada
	if len(params.Collections) == 0 {
		handleError(c, http.StatusBadRequest, "É necessário especificar pelo menos uma coleção", nil)
		return
	}

	// Validar parâmetros mínimos de busca
	if params.Params.Q == "" || params.Params.QueryBy == "" {
		handleError(c, http.StatusBadRequest, "Termo de busca (q) e campos para busca (query_by) são obrigatórios", nil)
		return
	}

	// Executar busca em múltiplas coleções
	results, err := h.service.SearchMultiCollection(c.Request.Context(), params)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao realizar busca em múltiplas coleções", err)
		return
	}

	c.JSON(http.StatusOK, results)
}

// SearchDocuments busca documentos em uma coleção específica
// @Summary      Buscar documentos em uma coleção
// @Description  Busca documentos em uma coleção
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        collection  path      string                 true  "Nome da coleção"
// @Param        params      body      models.SearchParameters true  "Parâmetros de busca"
// @Success      200         {object}  models.SearchDocumentsResponse
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /api/v1/typesense/collections/{collection}/documents/search [post]
func (h *TypesenseHandler) SearchDocuments(c *gin.Context) {
	collectionName := c.Param("collection")
	if collectionName == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção é obrigatório", nil)
		return
	}

	var params models.SearchParameters
	if err := c.ShouldBindJSON(&params); err != nil {
		handleError(c, http.StatusBadRequest, "Parâmetros de busca inválidos", err)
		return
	}

	results, err := h.service.SearchDocuments(c.Request.Context(), collectionName, params)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao buscar documentos", err)
		return
	}

	c.JSON(http.StatusOK, results)
}
