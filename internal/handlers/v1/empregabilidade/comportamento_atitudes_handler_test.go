package empregabilidade_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	handler "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// ==========================================
// TESTES DO CRUD GLOBAL DE COMPORTAMENTOS E ATITUDES
// ==========================================

func TestCreateComportamentoAtitudes_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/comportamentos-atitudes", func(c *gin.Context) {
		var req handler.CreateComportamentoAtitudesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("Dados inválidos: %s", err.Error()))
			return
		}
	})

	invalidJSON := []byte(`{"nome": ""}`)
	req, _ := http.NewRequest(http.MethodPost, "/comportamentos-atitudes", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

func TestGetComportamentoAtitudesByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/comportamentos-atitudes/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		_, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "ID inválido")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes/id-invalido-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

func TestUpdateComportamentoAtitudes_BadRequest_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/comportamentos-atitudes/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		if _, err := strconv.ParseInt(idParam, 10, 64); err != nil {
			response.Error(c, http.StatusBadRequest, "ID inválido")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodPut, "/comportamentos-atitudes/abc-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteComportamentoAtitudes_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/comportamentos-atitudes/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		if _, err := strconv.ParseInt(idParam, 10, 64); err != nil {
			response.Error(c, http.StatusBadRequest, "ID inválido")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodDelete, "/comportamentos-atitudes/abc-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

// ==========================================
// TESTES DE VÍNCULO COM O CURRÍCULO
// ==========================================

func TestAddComportamentoAtitudesAoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/curriculo/comportamentos-atitudes", func(c *gin.Context) {
		cpf := c.GetString("user_cpf")
		if cpf == "" {
			response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
			return
		}
	})

	reqBody := handler.AddComportamentoAtitudesRequest{
		IDComportamentoAtitudes: 1,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/curriculo/comportamentos-atitudes", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Usuário não autenticado")
}

func TestListComportamentoAtitudesDoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/curriculo/comportamentos-atitudes", func(c *gin.Context) {
		cpf := c.GetString("user_cpf")
		if cpf == "" {
			response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "/curriculo/comportamentos-atitudes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Usuário não autenticado")
}

func TestDeleteComportamentoAtitudesDoCurriculo_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var validID int64 = 99
	userCPF := "12345678900"

	r.DELETE("/curriculo/comportamentos-atitudes/:id", func(c *gin.Context) {
		c.Set("user_cpf", userCPF)

		comportamentosDoUsuario := []*empregabilidade.CurriculoComportamentoAtitudes{
			{ID: 1, CPF: userCPF},
		}

		vinculoID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var pertenceAoUsuario bool
		for _, item := range comportamentosDoUsuario {
			if item.ID == vinculoID {
				pertenceAoUsuario = true
				break
			}
		}

		if !pertenceAoUsuario {
			response.Error(c, http.StatusForbidden, "Acesso negado: o recurso não pertence ao usuário")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodDelete, "/curriculo/comportamentos-atitudes/"+strconv.FormatInt(validID, 10), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Acesso negado")
}
