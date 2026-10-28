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
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestInscricaoHandler_Update_AdminNotes covers the mass assignment fix for
// admin_notes (pentest vuln-0003): only admins (go:admin) and casa_civil may
// set admin_notes on PUT /courses/:courseId/enrollments/:enrollmentId; for the
// citizen who owns the enrollment the field is dropped before the service call.
func TestInscricaoHandler_Update_AdminNotes(t *testing.T) {
	const ownerCPF = "12345678901"
	const otherCPF = "11111111111"

	notes := "anotação interna"
	empty := ""

	tests := []struct {
		name           string
		role           string
		roles          []string
		cpf            string
		body           map[string]interface{}
		wantAdminNotes *string
	}{
		{
			name:           "citizen owner sending admin_notes has it dropped",
			role:           "USER",
			cpf:            ownerCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": notes},
			wantAdminNotes: nil,
		},
		{
			name:           "citizen owner sending empty admin_notes cannot clear it",
			role:           "USER",
			cpf:            ownerCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": empty},
			wantAdminNotes: nil,
		},
		{
			name:           "citizen owner with unrelated roles still has it dropped",
			role:           "USER",
			roles:          []string{"go:cursos:editor"},
			cpf:            ownerCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": notes},
			wantAdminNotes: nil,
		},
		{
			name:           "admin keeps admin_notes",
			role:           "ADMIN",
			cpf:            otherCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": notes},
			wantAdminNotes: &notes,
		},
		{
			name:           "admin can clear admin_notes",
			role:           "ADMIN",
			cpf:            otherCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": empty},
			wantAdminNotes: &empty,
		},
		{
			name:           "casa_civil keeps admin_notes",
			role:           "USER",
			roles:          []string{"go:cursos:casa_civil"},
			cpf:            otherCPF,
			body:           map[string]interface{}{"name": "Novo Nome", "admin_notes": notes},
			wantAdminNotes: &notes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set(middlewares.UserRoleKey, tt.role)
				c.Set(middlewares.UserCPFKey, tt.cpf)
				if tt.roles != nil {
					c.Set(middlewares.UserRolesKey, tt.roles)
				}
				c.Next()
			})
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId", handler.Update)

			enrollmentID := uuid.New()
			existing := &models.Inscricao{ID: enrollmentID, CursoID: 1, CPF: ownerCPF}

			var received *models.InscricaoUpdateRequest
			mockService.On("GetByID", mock.Anything, enrollmentID).Return(existing, nil)
			mockService.On("UpdateInscricao", mock.Anything, enrollmentID, 1, mock.AnythingOfType("*models.InscricaoUpdateRequest")).
				Run(func(args mock.Arguments) {
					received = args.Get(3).(*models.InscricaoUpdateRequest)
				}).
				Return(nil).Once()

			body, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/courses/1/enrollments/"+enrollmentID.String(), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
			require.NotNil(t, received)
			require.NotNil(t, received.Name, "other fields must still reach the service")
			assert.Equal(t, "Novo Nome", *received.Name)
			if tt.wantAdminNotes == nil {
				assert.Nil(t, received.AdminNotes)
			} else {
				require.NotNil(t, received.AdminNotes)
				assert.Equal(t, *tt.wantAdminNotes, *received.AdminNotes)
			}
		})
	}
}
