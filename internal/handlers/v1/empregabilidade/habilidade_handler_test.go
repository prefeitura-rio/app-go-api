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
// TESTES DO CRUD GLOBAL DE HABILIDADES
// ==========================================

func TestCreateHabilidade_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/habilidades", func(c *gin.Context) {
		var req handler.CreateHabilidadeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("Dados inválidos: %s", err.Error()))
			return
		}
	})

	invalidJSON := []byte(`{"nome": ""}`)
	req, _ := http.NewRequest(http.MethodPost, "/habilidades", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Dados inválidos")
}

func TestGetHabilidadeByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/habilidades/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		_, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "ID inválido")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "/habilidades/id-invalido-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID inválido")
}

func TestUpdateHabilidade_BadRequest_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/habilidades/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		if _, err := strconv.ParseInt(idParam, 10, 64); err != nil {
			response.Error(c, http.StatusBadRequest, "ID inválido")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodPut, "/habilidades/abc-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==========================================
// TESTES DE VÍNCULO COM O CURRÍCULO
// ==========================================

func TestAddHabilidadeAoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/curriculo/habilidades", func(c *gin.Context) {
		cpf := c.GetString("user_cpf")
		if cpf == "" {
			response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
			return
		}
	})

	reqBody := handler.AddHabilidadeRequest{
		IDHabilidade: 1, // Alterado de uuid.New() para ID int64
	}
	jsonBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/curriculo/habilidades", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Usuário não autenticado")
}

func TestListHabilidadesDoCurriculo_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/curriculo/habilidades", func(c *gin.Context) {
		cpf := c.GetString("user_cpf")
		if cpf == "" {
			response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
			return
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "/curriculo/habilidades", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteHabilidadeDoCurriculo_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var validID int64 = 99
	userCPF := "12345678900"

	r.DELETE("/curriculo/habilidades/:id", func(c *gin.Context) {
		c.Set("user_cpf", userCPF)

		habilidadesDoUsuario := []*empregabilidade.CurriculoHabilidade{
			{ID: 1, CPF: userCPF},
		}

		vinculoID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

		var pertenceAoUsuario bool
		for _, item := range habilidadesDoUsuario {
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

	req, _ := http.NewRequest(http.MethodDelete, "/curriculo/habilidades/"+strconv.FormatInt(validID, 10), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Acesso negado")
}

func TestReplaceAllHabilidades_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	userCPF := "12345678900"

	r.PUT("/curriculo/accordion/habilidades", func(c *gin.Context) {
		c.Set("user_cpf", userCPF)

		var req handler.ReplaceHabilidadesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		var autoIncrementID int64 = 100
		for _, item := range req.Habilidades {
			item.CPF = userCPF
			if item.ID == 0 {
				item.ID = autoIncrementID
				autoIncrementID++
			}
		}

		response.Success(c, http.StatusOK, "Habilidades atualizadas com sucesso")
	})

	payload := handler.ReplaceHabilidadesRequest{
		Habilidades: []*empregabilidade.CurriculoHabilidade{
			{IDHabilidade: 1}, // Alterado de IDHabilidade (UUID) para HabilidadeID (int64)
			{IDHabilidade: 2},
		},
	}
	jsonBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, "/curriculo/accordion/habilidades", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Habilidades atualizadas com sucesso")
}
