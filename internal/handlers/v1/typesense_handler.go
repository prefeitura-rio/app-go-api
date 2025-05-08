package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TypesenseHandler contém os handlers para operações do Typesense
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

// @Summary      Criar coleção
// @Description  Cria uma nova coleção no Typesense
// @Tags         collections
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateCollectionRequest  true  "Dados da coleção"
// @Success      201      {object}  models.CollectionResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /typesense/collections [post]
func (h *TypesenseHandler) CreateCollection(c *gin.Context) {
	var req models.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, "Requisição inválida", err)
		return
	}

	collection, err := h.service.CreateCollection(c.Request.Context(), req)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao criar coleção", err)
		return
	}

	c.JSON(http.StatusCreated, collection)
}

// @Summary      Obter coleção
// @Description  Obtém uma coleção pelo nome
// @Tags         collections
// @Produce      json
// @Param        name  path      string  true  "Nome da coleção"
// @Success      200   {object}  models.CollectionResponse
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /typesense/collections/{name} [get]
func (h *TypesenseHandler) GetCollection(c *gin.Context) {
	collectionName := c.Param("name")
	if collectionName == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção é obrigatório", nil)
		return
	}

	collection, err := h.service.GetCollection(c.Request.Context(), collectionName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao obter coleção", err)
		return
	}

	c.JSON(http.StatusOK, collection)
}

// @Summary      Listar coleções
// @Description  Lista todas as coleções no Typesense
// @Tags         collections
// @Produce      json
// @Success      200  {array}   models.CollectionResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /typesense/collections [get]
func (h *TypesenseHandler) ListCollections(c *gin.Context) {
	collections, err := h.service.ListCollections(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao listar coleções", err)
		return
	}

	c.JSON(http.StatusOK, collections)
}

// @Summary      Excluir coleção
// @Description  Exclui uma coleção pelo nome
// @Tags         collections
// @Param        name  path  string  true  "Nome da coleção"
// @Success      204  "No Content"
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /typesense/collections/{name} [delete]
func (h *TypesenseHandler) DeleteCollection(c *gin.Context) {
	collectionName := c.Param("name")
	if collectionName == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção é obrigatório", nil)
		return
	}

	err := h.service.DeleteCollection(c.Request.Context(), collectionName)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao excluir coleção", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Inserir/atualizar documento
// @Description  Insere ou atualiza um documento em uma coleção
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        collection  path      string                       true  "Nome da coleção"
// @Param        request     body      models.UpsertDocumentRequest  true  "Dados do documento"
// @Success      200         {object}  models.Document
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /typesense/collections/{collection}/documents [post]
func (h *TypesenseHandler) UpsertDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	if collectionName == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção é obrigatório", nil)
		return
	}

	var req models.UpsertDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, "Requisição inválida", err)
		return
	}

	// Valor padrão para action é "upsert" se não for especificado
	action := req.ActionOnMatch
	if action == "" {
		action = "upsert"
	}

	document, err := h.service.UpsertDocument(c.Request.Context(), collectionName, req.Document, action)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao inserir/atualizar documento", err)
		return
	}

	c.JSON(http.StatusOK, document)
}

// @Summary      Obter documento
// @Description  Obtém um documento pelo ID
// @Tags         documents
// @Produce      json
// @Param        collection  path      string  true  "Nome da coleção"
// @Param        id          path      string  true  "ID do documento"
// @Success      200         {object}  models.Document
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /typesense/collections/{collection}/documents/{id} [get]
func (h *TypesenseHandler) GetDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	documentID := c.Param("id")

	if collectionName == "" || documentID == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção e ID do documento são obrigatórios", nil)
		return
	}

	document, err := h.service.GetDocument(c.Request.Context(), collectionName, documentID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao obter documento", err)
		return
	}

	c.JSON(http.StatusOK, document)
}

// @Summary      Excluir documento
// @Description  Exclui um documento pelo ID
// @Tags         documents
// @Param        collection  path  string  true  "Nome da coleção"
// @Param        id          path  string  true  "ID do documento"
// @Success      204         "No Content"
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /typesense/collections/{collection}/documents/{id} [delete]
func (h *TypesenseHandler) DeleteDocument(c *gin.Context) {
	collectionName := c.Param("collection")
	documentID := c.Param("id")

	if collectionName == "" || documentID == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção e ID do documento são obrigatórios", nil)
		return
	}

	err := h.service.DeleteDocument(c.Request.Context(), collectionName, documentID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao excluir documento", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary      Buscar documentos
// @Description  Busca documentos em uma coleção
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        collection  path      string                 true  "Nome da coleção"
// @Param        params      body      models.SearchParameters true  "Parâmetros de busca"
// @Success      200         {object}  models.SearchDocumentsResponse
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /typesense/collections/{collection}/documents/search [post]
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

// @Summary      Importar documentos
// @Description  Importa múltiplos documentos em batch para uma coleção
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        collection  path      string        true  "Nome da coleção"
// @Param        action      query     string        false "Ação em caso de documento existente (default: create)"
// @Param        documents   body      []models.Document true  "Lista de documentos"
// @Success      200         {object}  map[string]int
// @Failure      400         {object}  models.ErrorResponse
// @Failure      500         {object}  models.ErrorResponse
// @Router       /typesense/collections/{collection}/documents/import [post]
func (h *TypesenseHandler) ImportDocuments(c *gin.Context) {
	collectionName := c.Param("collection")
	if collectionName == "" {
		handleError(c, http.StatusBadRequest, "Nome da coleção é obrigatório", nil)
		return
	}

	action := c.DefaultQuery("action", "create")

	// Ler body como bytes
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		handleError(c, http.StatusBadRequest, "Erro ao ler corpo da requisição", err)
		return
	}

	// Processar como array de documentos
	var documents []models.Document
	if err := json.Unmarshal(body, &documents); err != nil {
		handleError(c, http.StatusBadRequest, "Formato de documentos inválido", err)
		return
	}

	count, err := h.service.ImportDocuments(c.Request.Context(), collectionName, documents, action)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Erro ao importar documentos", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imported_count": count,
	})
}