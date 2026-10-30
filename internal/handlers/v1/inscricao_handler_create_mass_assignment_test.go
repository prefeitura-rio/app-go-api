package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestInscricaoHandler_Create_InternalFields covers the mass assignment fix on
// POST /courses/:courseId/enrollments (DEV-134): a citizen creating their own
// enrollment cannot set admin_notes, certificate_url or concluded_at. Those are
// dropped before the service call; admins and casa_civil keep them.
func TestInscricaoHandler_Create_InternalFields(t *testing.T) {
	const citizenCPF = "12345678901"
	scheduleID := uuid.New()

	body := map[string]interface{}{
		"name":            "Cidadão Teste",
		"cpf":             citizenCPF,
		"reason":          "Quero aprender",
		"schedule_id":     scheduleID.String(),
		"admin_notes":     "aprovado por telefone",
		"certificate_url": "https://exemplo.invalid/cert.pdf",
		"concluded_at":    "2026-01-01T00:00:00Z",
	}

	tests := []struct {
		name          string
		role          string
		roles         []string
		serviceMethod string
		wantInternal  bool
	}{
		{name: "citizen has internal fields dropped", role: "USER", serviceMethod: "Create", wantInternal: false},
		{name: "admin keeps internal fields", role: "ADMIN", serviceMethod: "CreateByAdmin", wantInternal: true},
		{name: "casa_civil keeps internal fields", role: "USER", roles: []string{"go:cursos:casa_civil"}, serviceMethod: "CreateByAdmin", wantInternal: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set("user_role", tt.role)
				if tt.roles != nil {
					c.Set("user_roles", tt.roles)
				}
				c.Set("user_cpf", citizenCPF)
				c.Next()
			})
			r.POST("/api/v1/courses/:courseId/enrollments", handler.Create)

			var captured *models.Inscricao
			mockService.On(tt.serviceMethod, mock.Anything, mock.AnythingOfType("*models.Inscricao")).
				Run(func(args mock.Arguments) {
					captured = args.Get(1).(*models.Inscricao)
					captured.Status = models.StatusInscricaoPending
				}).
				Return(nil)

			raw, err := json.Marshal(body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/1/enrollments", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			mockService.AssertExpectations(t)
			require.NotNil(t, captured)

			// Citizen-provided fields are kept in both flows.
			assert.Equal(t, "Quero aprender", captured.Reason)
			require.NotNil(t, captured.ScheduleID)
			assert.Equal(t, scheduleID, *captured.ScheduleID)

			if tt.wantInternal {
				assert.Equal(t, "aprovado por telefone", captured.AdminNotes)
				assert.Equal(t, "https://exemplo.invalid/cert.pdf", captured.CertificateURL)
				require.NotNil(t, captured.ConcludedAt)
				assert.True(t, captured.ConcludedAt.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
			} else {
				assert.Empty(t, captured.AdminNotes)
				assert.Empty(t, captured.CertificateURL)
				assert.Nil(t, captured.ConcludedAt)
			}
		})
	}
}
