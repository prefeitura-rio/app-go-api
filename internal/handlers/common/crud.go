package common

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/httputil"
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
	cache      *cache.ReferenceDataCache
}

func NewCRUDHandler[T any](service CRUDService[T], entityName string) *CRUDHandler[T] {
	return &CRUDHandler[T]{
		service:    service,
		entityName: entityName,
		cache:      nil,
	}
}

func (h *CRUDHandler[T]) WithCache(cache *cache.ReferenceDataCache) *CRUDHandler[T] {
	h.cache = cache
	return h
}

func (h *CRUDHandler[T]) Create(c *gin.Context) {
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		httputil.BadRequest(c, "Dados inválidos: "+err.Error())
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	type IDSetter interface {
		SetID(int)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	if h.cache != nil {
		_ = h.cache.Invalidate(c.Request.Context())
	}

	httputil.Created(c, entity)
}

func (h *CRUDHandler[T]) GetByID(c *gin.Context) {
	id, ok := httputil.ParseIntID(c, "id")
	if !ok {
		return
	}

	if h.cache != nil {
		cachedData, err := h.cache.GetByID(c.Request.Context(), id)
		if err == nil && cachedData != nil {
			var entity T
			if err := json.Unmarshal(cachedData, &entity); err == nil {
				httputil.Success(c, entity)
				return
			}
		}
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	if entity == nil {
		httputil.NotFound(c, h.entityName+" não encontrado(a)")
		return
	}

	if h.cache != nil {
		_ = h.cache.SetByID(c.Request.Context(), id, entity)
	}

	httputil.Success(c, entity)
}

func (h *CRUDHandler[T]) Update(c *gin.Context) {
	id, ok := httputil.ParseIntID(c, "id")
	if !ok {
		return
	}

	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		httputil.BadRequest(c, "Dados inválidos: "+err.Error())
		return
	}

	type IDSetter interface {
		SetID(int)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	if h.cache != nil {
		_ = h.cache.InvalidateByID(c.Request.Context(), id)
	}

	httputil.Success(c, entity)
}

func (h *CRUDHandler[T]) Delete(c *gin.Context) {
	id, ok := httputil.ParseIntID(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	if h.cache != nil {
		_ = h.cache.InvalidateByID(c.Request.Context(), id)
	}

	httputil.Deleted(c, h.entityName)
}

func (h *CRUDHandler[T]) List(c *gin.Context) {
	page, pageSize := httputil.Pagination(c)

	if h.cache != nil && page == 1 && pageSize == 10 {
		cachedData, err := h.cache.GetList(c.Request.Context())
		if err == nil && cachedData != nil {
			type cachedResponse struct {
				Data []*T   `json:"data"`
				Meta gin.H  `json:"meta"`
			}
			var cached cachedResponse
			if err := json.Unmarshal(cachedData, &cached); err == nil {
				c.JSON(http.StatusOK, gin.H{
					"data": cached.Data,
					"meta": cached.Meta,
				})
				return
			}
		}
	}

	filter := make(map[string]interface{})

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
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

	if h.cache != nil && page == 1 && pageSize == 10 {
		_ = h.cache.SetList(c.Request.Context(), response)
	}

	c.JSON(http.StatusOK, response)
}
