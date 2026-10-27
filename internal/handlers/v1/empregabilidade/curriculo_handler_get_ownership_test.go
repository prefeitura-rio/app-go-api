package empregabilidade_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

const (
	getOwnershipOwnerCPF = "11122233344"
	getOwnershipOtherCPF = "55566677788"
	getOwnershipMarker   = "dado-sensivel-do-dono"
)

// getByIDOwnershipCases builds one repo per curriculo sub-resource holding an item
// owned by getOwnershipOwnerCPF, plus the GET path that reads it by ID.
func getByIDOwnershipCases() []struct {
	name string
	path string
	repo func() *mockCurriculoRepoH
} {
	id := uuid.MustParse(validUUID)
	return []struct {
		name string
		path string
		repo func() *mockCurriculoRepoH
	}{
		{
			name: "formacao",
			path: "/curriculo/formacoes/" + validUUID,
			repo: func() *mockCurriculoRepoH {
				return &mockCurriculoRepoH{formacao: &empmodels.CurriculoFormacao{ID: id, CPF: getOwnershipOwnerCPF, NomeInstituicao: getOwnershipMarker}}
			},
		},
		{
			name: "idioma",
			path: "/curriculo/idiomas/" + validUUID,
			repo: func() *mockCurriculoRepoH {
				return &mockCurriculoRepoH{idioma: &empmodels.CurriculoIdioma{ID: id, CPF: getOwnershipOwnerCPF}}
			},
		},
		{
			name: "curso complementar",
			path: "/curriculo/cursos/" + validUUID,
			repo: func() *mockCurriculoRepoH {
				return &mockCurriculoRepoH{curso: &empmodels.CurriculoCursoComplementar{ID: id, CPF: getOwnershipOwnerCPF, NomeCurso: getOwnershipMarker}}
			},
		},
		{
			name: "experiencia",
			path: "/curriculo/experiencias/" + validUUID,
			repo: func() *mockCurriculoRepoH {
				return &mockCurriculoRepoH{experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: getOwnershipOwnerCPF, Cargo: getOwnershipMarker}}
			},
		},
		{
			name: "conquista",
			path: "/curriculo/conquistas/" + validUUID,
			repo: func() *mockCurriculoRepoH {
				return &mockCurriculoRepoH{conquista: &empmodels.CurriculoConquista{ID: id, CPF: getOwnershipOwnerCPF, Titulo: getOwnershipMarker}}
			},
		},
	}
}

func TestCurriculoHandler_GetByID_OtherCitizen_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range getByIDOwnershipCases() {
		t.Run(tc.name, func(t *testing.T) {
			r := setupCurriculoRouter(tc.repo(), getOwnershipOtherCPF, false)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NotContains(t, w.Body.String(), getOwnershipOwnerCPF)
			assert.NotContains(t, w.Body.String(), getOwnershipMarker)
		})
	}
}

func TestCurriculoHandler_GetByID_Owner_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range getByIDOwnershipCases() {
		t.Run(tc.name, func(t *testing.T) {
			r := setupCurriculoRouter(tc.repo(), getOwnershipOwnerCPF, false)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), validUUID)
		})
	}
}

func TestCurriculoHandler_GetByID_Admin_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range getByIDOwnershipCases() {
		t.Run(tc.name, func(t *testing.T) {
			r := setupCurriculoRouter(tc.repo(), getOwnershipOtherCPF, true)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), validUUID)
		})
	}
}

// Without any user context (no CPF, no role) the handler itself must refuse
// with 401 instead of returning the item.
func TestCurriculoHandler_GetByID_NoUserContext_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := map[string]func(h *handlers.CurriculoHandler) gin.HandlerFunc{
		"formacao":           func(h *handlers.CurriculoHandler) gin.HandlerFunc { return h.GetFormacaoByID },
		"idioma":             func(h *handlers.CurriculoHandler) gin.HandlerFunc { return h.GetIdiomaByID },
		"curso complementar": func(h *handlers.CurriculoHandler) gin.HandlerFunc { return h.GetCursoComplementarByID },
		"experiencia":        func(h *handlers.CurriculoHandler) gin.HandlerFunc { return h.GetExperienciaByID },
		"conquista":          func(h *handlers.CurriculoHandler) gin.HandlerFunc { return h.GetConquistaByID },
	}
	for _, tc := range getByIDOwnershipCases() {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewCurriculoHandler(services.NewCurriculoServiceWithInterface(tc.repo()))
			r := gin.New()
			r.GET("/item/:id", routes[tc.name](h))

			req := httptest.NewRequest(http.MethodGet, "/item/"+validUUID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.NotContains(t, w.Body.String(), getOwnershipOwnerCPF)
			assert.NotContains(t, w.Body.String(), getOwnershipMarker)
		})
	}
}
