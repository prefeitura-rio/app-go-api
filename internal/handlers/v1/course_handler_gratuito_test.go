package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

// DEV-715: a listagem pública informa, em cada curso, se ele é gratuito. O
// app-catalogo lê o campo no sync para aplicar o filtro gratuito da busca.
func TestCourseHandler_ListPublic_ReturnsGratuito(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)
	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/public/courses", handler.ListPublic)

	cursos := []*models.Curso{
		{ID: 1, Titulo: "Gratuito", Status: models.StatusCursoPublished, Gratuito: boolPtr(true)},
		{ID: 2, Titulo: "Pago", Status: models.StatusCursoPublished, Gratuito: boolPtr(false)},
		// Sem valor carregado (não acontece com a coluna NOT NULL): sai como
		// gratuito, o padrão da coluna, nunca como null.
		{ID: 3, Titulo: "Sem valor", Status: models.StatusCursoPublished},
	}
	mockService.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(cursos, len(cursos), nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/public/courses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Data struct {
			Courses []map[string]any `json:"courses"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Data.Courses, 3)

	want := map[float64]bool{1: true, 2: false, 3: true}
	for _, c := range body.Data.Courses {
		id := c["id"].(float64)
		got, ok := c["gratuito"].(bool)
		require.True(t, ok, "curso %v: gratuito deve ser booleano, veio %#v", id, c["gratuito"])
		assert.Equal(t, want[id], got, "curso %v", id)
	}
}

// O detalhe público serializa models.Curso: o campo sai com o valor gravado.
func TestCourseHandler_GetByIDPublic_ReturnsGratuito(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)
	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/public/courses/:courseId", handler.GetByIDPublic)

	mockService.On("GetByID", mock.Anything, 2).Return(&models.Curso{
		ID: 2, Titulo: "Pago", Status: models.StatusCursoPublished, Gratuito: boolPtr(false),
	}, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil).Maybe()

	req := httptest.NewRequest(http.MethodGet, "/api/public/courses/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, false, body.Data["gratuito"])
}

// O PUT grava a linha inteira (Select("*")): sem gratuito no corpo, o valor
// gravado é preservado, para não virar NULL (a coluna é NOT NULL) nem mudar o
// curso pago para gratuito. Com gratuito no corpo, vale o enviado.
func TestCourseHandler_Update_Gratuito(t *testing.T) {
	tests := []struct {
		name     string
		existing *bool
		body     string
		want     bool
	}{
		{"ausente preserva pago", boolPtr(false), `{"title":"Novo"}`, false},
		{"ausente preserva gratuito", boolPtr(true), `{"title":"Novo"}`, true},
		{"null preserva pago", boolPtr(false), `{"title":"Novo","gratuito":null}`, false},
		{"false marca pago", boolPtr(true), `{"title":"Novo","gratuito":false}`, false},
		{"true marca gratuito", boolPtr(false), `{"title":"Novo","gratuito":true}`, true},
		{"ausente sem valor gravado vira gratuito", nil, `{"title":"Novo"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockCursoService)
			mockInscricaoService := new(MockInscricaoServiceForCourse)
			mockRepo := new(MockCursoRepositoryForCourseHandler)
			handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

			r := gin.New()
			r.PUT("/api/v1/courses/:courseId", handler.Update)

			mockService.On("GetByID", mock.Anything, 1).Return(&models.Curso{
				ID: 1, Titulo: "Antigo", Status: models.StatusCursoPublished, Gratuito: tt.existing,
			}, nil)
			var saved *models.Curso
			mockService.On("Update", mock.Anything, mock.AnythingOfType("*models.Curso")).
				Run(func(args mock.Arguments) { saved = args.Get(1).(*models.Curso) }).
				Return(nil)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.NotNil(t, saved)
			require.NotNil(t, saved.Gratuito, "gratuito nunca vai NULL para o UPDATE")
			assert.Equal(t, tt.want, *saved.Gratuito)
		})
	}
}

// Na criação (publicada ou rascunho), gratuito ausente fica nil e o banco
// aplica o padrão (true); false enviado chega ao serviço como pago.
func TestCourseHandler_Create_Gratuito(t *testing.T) {
	routes := []struct {
		name string
		path string
		reg  func(h *v1.CourseHandler) gin.HandlerFunc
	}{
		{"create", "/api/v1/courses", func(h *v1.CourseHandler) gin.HandlerFunc { return h.Create }},
		{"draft", "/api/v1/courses/draft", func(h *v1.CourseHandler) gin.HandlerFunc { return h.CreateDraft }},
	}
	bodies := []struct {
		name string
		body string
		want *bool
	}{
		{"ausente usa o padrão do banco", `{"title":"Curso"}`, nil},
		{"false marca pago", `{"title":"Curso","gratuito":false}`, boolPtr(false)},
		{"true marca gratuito", `{"title":"Curso","gratuito":true}`, boolPtr(true)},
	}

	for _, rt := range routes {
		for _, tt := range bodies {
			t.Run(rt.name+"/"+tt.name, func(t *testing.T) {
				gin.SetMode(gin.TestMode)
				mockService := new(MockCursoService)
				mockInscricaoService := new(MockInscricaoServiceForCourse)
				mockRepo := new(MockCursoRepositoryForCourseHandler)
				handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

				r := gin.New()
				r.POST(rt.path, rt.reg(handler))

				var saved *models.Curso
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Curso")).
					Run(func(args mock.Arguments) { saved = args.Get(1).(*models.Curso) }).
					Return(1, nil)

				req := httptest.NewRequest(http.MethodPost, rt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
				require.NotNil(t, saved)
				assert.Equal(t, tt.want, saved.Gratuito)
			})
		}
	}
}
