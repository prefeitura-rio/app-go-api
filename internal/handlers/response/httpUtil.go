package response

import (
	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// DTOs / Contratos Genéricos

type ErrorResponse struct {
	Error string `json:"error" example:"Mensagem de erro aqui"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Operação realizada com sucesso"`
}

type Paginated[T any] struct {
	Data     []T   `json:"data"`
	Total    int64 `json:"total" example:"42"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"pageSize" example:"20"`
}

type ListHabilidadesPaginatedResponse = Paginated[[]empregabilidade.Habilidade]

type ListAreasAtuacaoPaginatedResponse = Paginated[[]empregabilidade.AreaAtuacao]

// Helpers HTTP

// Error envia uma resposta JSON padronizada para erros.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

// Success envia uma resposta JSON padronizada para mensagens de sucesso.
func Success(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, SuccessResponse{
		Message: message,
	})
}

// PaginatedJSON envia uma resposta JSON padronizada com estrutura de paginação.
func PaginatedJSON[T any](c *gin.Context, statusCode int, data []T, total int64, page, pageSize int) {
	if data == nil {
		data = make([]T, 0)
	}

	c.JSON(statusCode, Paginated[T]{
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
