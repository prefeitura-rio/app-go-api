package common

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

type CRUDService[T any] interface {
	Create(ctx context.Context, entity *T) (int, error)
	GetByID(ctx context.Context, id int) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*T, int, error)
}

type CRUDHandler[T any] struct {
	service    CRUDService[T]
	entityName string
	cache      *cache.ReferenceDataCache // Optional cache for static reference data
}

func NewCRUDHandler[T any](service CRUDService[T], entityName string) *CRUDHandler[T] {
	return &CRUDHandler[T]{
		service:    service,
		entityName: entityName,
		cache:      nil,
	}
}

// WithCache adds cache support to the CRUD handler
// Should be used for static reference data that rarely changes
func (h *CRUDHandler[T]) WithCache(cache *cache.ReferenceDataCache) *CRUDHandler[T] {
	h.cache = cache
	return h
}

func (h *CRUDHandler[T]) Create(c *gin.Context) {
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	type IDSetter interface {
		SetID(int)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	// Invalidate cache after create (if caching is enabled)
	if h.cache != nil {
		_ = h.cache.Invalidate(c.Request.Context())
	}

	c.JSON(http.StatusCreated, entity)
}

func (h *CRUDHandler[T]) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Try cache first (if caching is enabled)
	if h.cache != nil {
		cachedData, err := h.cache.GetByID(c.Request.Context(), id)
		if err == nil && cachedData != nil {
			var entity T
			if err := json.Unmarshal(cachedData, &entity); err == nil {
				c.JSON(http.StatusOK, entity)
				return
			}
			// If unmarshal fails, fall through to database query
		}
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": h.entityName + " não encontrada"})
		return
	}

	// Cache the result (if caching is enabled)
	if h.cache != nil {
		_ = h.cache.SetByID(c.Request.Context(), id, entity)
	}

	c.JSON(http.StatusOK, entity)
}

func (h *CRUDHandler[T]) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	type IDSetter interface {
		SetID(int)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	// Invalidate cache after update (if caching is enabled)
	if h.cache != nil {
		_ = h.cache.InvalidateByID(c.Request.Context(), id)
	}

	c.JSON(http.StatusOK, entity)
}

func (h *CRUDHandler[T]) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	// Invalidate cache after delete (if caching is enabled)
	if h.cache != nil {
		_ = h.cache.InvalidateByID(c.Request.Context(), id)
	}

	c.JSON(http.StatusOK, gin.H{"message": h.entityName + " excluída com sucesso"})
}

func (h *CRUDHandler[T]) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	// For static reference data with caching, only cache the default listing (page 1, no filters)
	// This covers the most common use case (getting all categorias/acessibilidades/escolaridades)
	if h.cache != nil && page == 1 && pageSize == 10 {
		cachedData, err := h.cache.GetList(c.Request.Context())
		if err == nil && cachedData != nil {
			// Cache returns the full response structure
			type cachedResponse struct {
				Data []*T  `json:"data"`
				Meta gin.H `json:"meta"`
			}
			var cached cachedResponse
			if err := json.Unmarshal(cachedData, &cached); err == nil {
				c.JSON(http.StatusOK, gin.H{
					"data": cached.Data,
					"meta": cached.Meta,
				})
				return
			}
			// If unmarshal fails, fall through to database query
		}
	}

	filter := make(map[string]interface{})

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	response := gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	}

	// Cache the result for default listing (if caching is enabled)
	if h.cache != nil && page == 1 && pageSize == 10 {
		_ = h.cache.SetList(c.Request.Context(), response)
	}

	c.JSON(http.StatusOK, response)
}
