package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

func handleLookupError(c *gin.Context, err error) {
	dbErr := utils.ParseDatabaseError(err)
	c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
}

// ModeloTrabalhoHandler

type ModeloTrabalhoHandler struct {
	service *services.ModeloTrabalhoService
}

func NewModeloTrabalhoHandler(service *services.ModeloTrabalhoService) *ModeloTrabalhoHandler {
	return &ModeloTrabalhoHandler{service: service}
}

func (h *ModeloTrabalhoHandler) Create(c *gin.Context) {
	var entity empregabilidade.ModeloTrabalho
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *ModeloTrabalhoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *ModeloTrabalhoHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Modelo de trabalho não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *ModeloTrabalhoHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.ModeloTrabalho
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *ModeloTrabalhoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Modelo de trabalho excluído com sucesso"})
}

// TipoPCDHandler

type TipoPCDHandler struct {
	service *services.TipoPCDService
}

func NewTipoPCDHandler(service *services.TipoPCDService) *TipoPCDHandler {
	return &TipoPCDHandler{service: service}
}

func (h *TipoPCDHandler) Create(c *gin.Context) {
	var entity empregabilidade.TipoPCD
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *TipoPCDHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *TipoPCDHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tipo PCD não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *TipoPCDHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.TipoPCD
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *TipoPCDHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo PCD excluído com sucesso"})
}

// IdiomaHandler

type IdiomaHandler struct {
	service *services.IdiomaService
}

func NewIdiomaHandler(service *services.IdiomaService) *IdiomaHandler {
	return &IdiomaHandler{service: service}
}

func (h *IdiomaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Idioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *IdiomaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *IdiomaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idioma não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *IdiomaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Idioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *IdiomaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idioma excluído com sucesso"})
}

// NivelIdiomaHandler

type NivelIdiomaHandler struct {
	service *services.NivelIdiomaService
}

func NewNivelIdiomaHandler(service *services.NivelIdiomaService) *NivelIdiomaHandler {
	return &NivelIdiomaHandler{service: service}
}

func (h *NivelIdiomaHandler) Create(c *gin.Context) {
	var entity empregabilidade.NivelIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *NivelIdiomaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *NivelIdiomaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nível de idioma não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *NivelIdiomaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.NivelIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *NivelIdiomaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nível de idioma excluído com sucesso"})
}

// EscolaridadeHandler

type EscolaridadeHandler struct {
	service *services.EscolaridadeService
}

func NewEscolaridadeHandler(service *services.EscolaridadeService) *EscolaridadeHandler {
	return &EscolaridadeHandler{service: service}
}

func (h *EscolaridadeHandler) Create(c *gin.Context) {
	var entity empregabilidade.Escolaridade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *EscolaridadeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *EscolaridadeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Escolaridade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *EscolaridadeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Escolaridade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *EscolaridadeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Escolaridade excluída com sucesso"})
}

// TipoConquistaHandler

type TipoConquistaHandler struct {
	service *services.TipoConquistaService
}

func NewTipoConquistaHandler(service *services.TipoConquistaService) *TipoConquistaHandler {
	return &TipoConquistaHandler{service: service}
}

func (h *TipoConquistaHandler) Create(c *gin.Context) {
	var entity empregabilidade.TipoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *TipoConquistaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *TipoConquistaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tipo de conquista não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *TipoConquistaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.TipoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *TipoConquistaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo de conquista excluído com sucesso"})
}

// SituacaoAtualHandler

type SituacaoAtualHandler struct {
	service *services.SituacaoAtualService
}

func NewSituacaoAtualHandler(service *services.SituacaoAtualService) *SituacaoAtualHandler {
	return &SituacaoAtualHandler{service: service}
}

func (h *SituacaoAtualHandler) Create(c *gin.Context) {
	var entity empregabilidade.SituacaoAtual
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *SituacaoAtualHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *SituacaoAtualHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Situação atual não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *SituacaoAtualHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.SituacaoAtual
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *SituacaoAtualHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Situação atual excluída com sucesso"})
}

// DisponibilidadeHandler

type DisponibilidadeHandler struct {
	service *services.DisponibilidadeService
}

func NewDisponibilidadeHandler(service *services.DisponibilidadeService) *DisponibilidadeHandler {
	return &DisponibilidadeHandler{service: service}
}

func (h *DisponibilidadeHandler) Create(c *gin.Context) {
	var entity empregabilidade.Disponibilidade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

func (h *DisponibilidadeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

func (h *DisponibilidadeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Disponibilidade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *DisponibilidadeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Disponibilidade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *DisponibilidadeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Disponibilidade excluída com sucesso"})
}
