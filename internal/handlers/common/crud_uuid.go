package common

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/httputil"
)

type CRUDServiceUUID[T any] interface {
	Create(ctx context.Context, entity *T) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*T, int, error)
}

type CRUDHandlerUUID[T any] struct {
	service    CRUDServiceUUID[T]
	entityName string
}

func NewCRUDHandlerUUID[T any](service CRUDServiceUUID[T], entityName string) *CRUDHandlerUUID[T] {
	return &CRUDHandlerUUID[T]{
		service:    service,
		entityName: entityName,
	}
}

func (h *CRUDHandlerUUID[T]) Create(c *gin.Context) {
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
		SetID(uuid.UUID)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	httputil.Created(c, entity)
}

func (h *CRUDHandlerUUID[T]) GetByID(c *gin.Context) {
	id, ok := httputil.ParseUUID(c, "id")
	if !ok {
		return
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

	httputil.Success(c, entity)
}

func (h *CRUDHandlerUUID[T]) Update(c *gin.Context) {
	id, ok := httputil.ParseUUID(c, "id")
	if !ok {
		return
	}

	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		httputil.BadRequest(c, "Dados inválidos: "+err.Error())
		return
	}

	type IDSetter interface {
		SetID(uuid.UUID)
	}
	if setter, ok := any(&entity).(IDSetter); ok {
		setter.SetID(id)
	}

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	httputil.Success(c, entity)
}

func (h *CRUDHandlerUUID[T]) Delete(c *gin.Context) {
	id, ok := httputil.ParseUUID(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	httputil.Deleted(c, h.entityName)
}

func (h *CRUDHandlerUUID[T]) List(c *gin.Context) {
	page, pageSize := httputil.Pagination(c)

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		httputil.Error(c, apperrors.FromDatabaseError(err))
		return
	}

	httputil.Paginated(c, entities, page, pageSize, total)
}
