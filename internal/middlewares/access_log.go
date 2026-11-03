package middlewares

import (
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// redactedValue substitui no log o valor de um parâmetro de query que pode ter dado pessoal.
	redactedValue = "[omitido]"
	// unmatchedRoute substitui no log o caminho de uma requisição que não casou com nenhuma rota.
	unmatchedRoute = "(rota não encontrada)"
)

// safeQueryParams são os parâmetros de query cujo valor nunca é dado pessoal (paginação e ordenação).
var safeQueryParams = map[string]bool{
	"page":      true,
	"limit":     true,
	"offset":    true,
	"per_page":  true,
	"page_size": true,
	"sort":      true,
	"sort_by":   true,
	"order":     true,
	"order_by":  true,
}

// AccessLogger substitui o gin.Logger: grava uma linha por requisição no mesmo formato
// ("[GIN] data | status | latência | IP | método caminho"), mas sem dados pessoais. O caminho
// é o template da rota (/api/v1/enrollments/user/:cpf, não o CPF) e os valores da query string
// são omitidos, exceto os de paginação e ordenação.
func AccessLogger(out io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}

		_, _ = fmt.Fprintf(out, "[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
			start.Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.Method,
			sanitizedRequestPath(c),
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}

// sanitizedRequestPath devolve o template da rota com a query string sem os valores sensíveis.
func sanitizedRequestPath(c *gin.Context) string {
	route := c.FullPath()
	if route == "" {
		return unmatchedRoute
	}

	query := c.Request.URL.Query()
	if len(query) == 0 {
		return route
	}

	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		escapedKey := url.QueryEscape(key)
		for _, value := range query[key] {
			if safeQueryParams[strings.ToLower(key)] {
				parts = append(parts, escapedKey+"="+url.QueryEscape(value))
			} else {
				parts = append(parts, escapedKey+"="+redactedValue)
			}
		}
	}
	return route + "?" + strings.Join(parts, "&")
}
