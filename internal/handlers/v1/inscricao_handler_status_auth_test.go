package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
)

// statusAuthCaller describes the authenticated user injected into the Gin
// context, mirroring what middlewares.ExtractUserContext sets. A zero value is
// an anonymous request (no token).
type statusAuthCaller struct {
	role  string
	roles []string
	cpf   string
}

func (u statusAuthCaller) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if u.role != "" {
			c.Set(middlewares.UserRoleKey, u.role)
		}
		if u.cpf != "" {
			c.Set(middlewares.UserCPFKey, u.cpf)
		}
		if u.roles != nil {
			c.Set(middlewares.UserRolesKey, u.roles)
		}
		c.Next()
	}
}

const (
	statusAuthOwnerCPF = "12345678901"
	statusAuthOtherCPF = "11111111111"
)

var (
	statusAuthCitizenOwner = statusAuthCaller{role: "USER", cpf: statusAuthOwnerCPF}
	statusAuthCitizenOther = statusAuthCaller{role: "USER", cpf: statusAuthOtherCPF}
	statusAuthAnonymous    = statusAuthCaller{}
	statusAuthAdmin        = statusAuthCaller{role: "ADMIN", roles: []string{"go:admin"}, cpf: statusAuthOtherCPF}
	statusAuthCasaCivil    = statusAuthCaller{role: "USER", roles: []string{"go:cursos:casa_civil"}, cpf: statusAuthOtherCPF}
	statusAuthEditor       = statusAuthCaller{role: "USER", roles: []string{"go:cursos:editor"}, cpf: statusAuthOtherCPF}
	statusAuthEditorOwner  = statusAuthCaller{role: "USER", roles: []string{"go:cursos:editor"}, cpf: statusAuthOwnerCPF}
	statusAuthAdminOwner   = statusAuthCaller{role: "ADMIN", roles: []string{"go:admin"}, cpf: statusAuthOwnerCPF}
)

// TestInscricaoHandler_UpdateIndividualStatus_Authorization covers DEV-14: on
// PUT /courses/:courseId/enrollments/:enrollmentId/status only staff
// (go:admin, go:cursos:casa_civil, go:cursos:editor) may set any status and
// admin_notes. The citizen who owns the enrollment may only cancel it, and
// admin_notes is dropped. The check lives in the handler so it also holds with
// RBAC_ENABLED=false, when enrollmentSingleAccess is a no-op.
func TestInscricaoHandler_UpdateIndividualStatus_Authorization(t *testing.T) {
	tests := []struct {
		name       string
		caller     statusAuthCaller
		body       map[string]interface{}
		wantCode   int
		wantCall   bool
		wantStatus models.StatusInscricao
		wantReason string
		wantNotes  string
	}{
		{
			name:     "citizen owner cannot approve own enrollment",
			caller:   statusAuthCitizenOwner,
			body:     map[string]interface{}{"status": "approved", "admin_notes": "x"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "citizen owner cannot conclude own enrollment",
			caller:   statusAuthCitizenOwner,
			body:     map[string]interface{}{"status": "concluded"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "citizen owner cannot move own enrollment back to pending",
			caller:   statusAuthCitizenOwner,
			body:     map[string]interface{}{"status": "pending"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "citizen owner cannot reject own enrollment",
			caller:   statusAuthCitizenOwner,
			body:     map[string]interface{}{"status": "rejected"},
			wantCode: http.StatusForbidden,
		},
		{
			name:       "citizen owner cancels own enrollment and admin_notes is dropped",
			caller:     statusAuthCitizenOwner,
			body:       map[string]interface{}{"status": "cancelled", "reason": "desisti", "admin_notes": "x"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoCancelled,
			wantReason: "desisti",
			wantNotes:  "",
		},
		{
			name:     "another citizen cannot cancel someone else's enrollment",
			caller:   statusAuthCitizenOther,
			body:     map[string]interface{}{"status": "cancelled"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "anonymous request is refused",
			caller:   statusAuthAnonymous,
			body:     map[string]interface{}{"status": "cancelled"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "citizen owner with an unrelated role is still only allowed to cancel",
			caller:   statusAuthCaller{role: "USER", roles: []string{"go:vagas:editor"}, cpf: statusAuthOwnerCPF},
			body:     map[string]interface{}{"status": "approved"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "editor who owns the enrollment cannot approve it",
			caller:   statusAuthEditorOwner,
			body:     map[string]interface{}{"status": "approved", "admin_notes": "x"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "editor who owns the enrollment cannot conclude it",
			caller:   statusAuthEditorOwner,
			body:     map[string]interface{}{"status": "concluded"},
			wantCode: http.StatusForbidden,
		},
		{
			name:       "editor who owns the enrollment may cancel it and admin_notes is dropped",
			caller:     statusAuthEditorOwner,
			body:       map[string]interface{}{"status": "cancelled", "admin_notes": "x"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoCancelled,
			wantNotes:  "",
		},
		{
			name:       "admin who owns the enrollment keeps full access",
			caller:     statusAuthAdminOwner,
			body:       map[string]interface{}{"status": "approved", "admin_notes": "revisado"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoApproved,
			wantNotes:  "revisado",
		},
		{
			name:       "admin approves with admin_notes",
			caller:     statusAuthAdmin,
			body:       map[string]interface{}{"status": "approved", "reason": "ok", "admin_notes": "revisado"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoApproved,
			wantReason: "ok",
			wantNotes:  "revisado",
		},
		{
			name:       "casa_civil concludes with admin_notes",
			caller:     statusAuthCasaCivil,
			body:       map[string]interface{}{"status": "concluded", "admin_notes": "concluiu"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoConcluded,
			wantNotes:  "concluiu",
		},
		{
			name:       "editor rejects with admin_notes",
			caller:     statusAuthEditor,
			body:       map[string]interface{}{"status": "rejected", "reason": "sem vaga", "admin_notes": "lotado"},
			wantCode:   http.StatusOK,
			wantCall:   true,
			wantStatus: models.StatusInscricaoRejected,
			wantReason: "sem vaga",
			wantNotes:  "lotado",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			r := gin.New()
			r.Use(tt.caller.middleware())
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", handler.UpdateIndividualStatus)

			enrollmentID := uuid.New()
			mockService.On("GetByID", mock.Anything, enrollmentID).Return(&models.Inscricao{
				ID:      enrollmentID,
				CursoID: 1,
				CPF:     statusAuthOwnerCPF,
				Status:  models.StatusInscricaoPending,
			}, nil).Maybe()
			if tt.wantCall {
				mockService.On("UpdateStatus", mock.Anything, enrollmentID, tt.wantStatus, tt.wantReason, tt.wantNotes).
					Return(nil).Once()
			}

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut,
				"/api/v1/courses/1/enrollments/"+enrollmentID.String()+"/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantCall {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "UpdateStatus",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

// The status handler now loads the enrollment before updating it: an
// enrollment of another course is refused (as GET/PUT/DELETE of the enrollment
// already do), so the courseId in the path cannot be used to reach it.
func TestInscricaoHandler_UpdateIndividualStatus_EnrollmentLookup(t *testing.T) {
	tests := []struct {
		name      string
		caller    statusAuthCaller
		found     *models.Inscricao
		lookupErr error
		wantCode  int
	}{
		{
			name:     "enrollment of another course is refused for staff",
			caller:   statusAuthEditor,
			found:    &models.Inscricao{CursoID: 2, CPF: statusAuthOwnerCPF},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "enrollment of another course is refused for the owner",
			caller:   statusAuthCitizenOwner,
			found:    &models.Inscricao{CursoID: 2, CPF: statusAuthOwnerCPF},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing enrollment is 404",
			caller:   statusAuthAdmin,
			found:    nil,
			wantCode: http.StatusNotFound,
		},
		{
			name:      "lookup failure is 500",
			caller:    statusAuthAdmin,
			lookupErr: errors.New("db down"),
			wantCode:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			r := gin.New()
			r.Use(tt.caller.middleware())
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", handler.UpdateIndividualStatus)

			enrollmentID := uuid.New()
			if tt.found != nil {
				tt.found.ID = enrollmentID
				mockService.On("GetByID", mock.Anything, enrollmentID).Return(tt.found, nil)
			} else {
				mockService.On("GetByID", mock.Anything, enrollmentID).Return(nil, tt.lookupErr)
			}

			body, _ := json.Marshal(map[string]interface{}{"status": "cancelled"})
			req := httptest.NewRequest(http.MethodPut,
				"/api/v1/courses/1/enrollments/"+enrollmentID.String()+"/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			mockService.AssertNotCalled(t, "UpdateStatus",
				mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// TestInscricaoHandler_UpdateCertificate_Authorization covers DEV-14: only
// staff may set the certificate URL of an enrollment; the citizen (owner or
// not) is refused, with or without RBAC_ENABLED.
func TestInscricaoHandler_UpdateCertificate_Authorization(t *testing.T) {
	tests := []struct {
		name     string
		caller   statusAuthCaller
		wantCode int
		wantCall bool
	}{
		{name: "citizen owner cannot set own certificate", caller: statusAuthCitizenOwner, wantCode: http.StatusForbidden},
		{name: "another citizen cannot set certificate", caller: statusAuthCitizenOther, wantCode: http.StatusForbidden},
		{name: "anonymous request is refused", caller: statusAuthAnonymous, wantCode: http.StatusForbidden},
		{name: "admin sets certificate", caller: statusAuthAdmin, wantCode: http.StatusOK, wantCall: true},
		{name: "casa_civil sets certificate", caller: statusAuthCasaCivil, wantCode: http.StatusOK, wantCall: true},
		{name: "editor sets certificate", caller: statusAuthEditor, wantCode: http.StatusOK, wantCall: true},
		{name: "editor who owns the enrollment cannot set own certificate", caller: statusAuthEditorOwner, wantCode: http.StatusForbidden},
		{name: "admin who owns the enrollment sets certificate", caller: statusAuthAdminOwner, wantCode: http.StatusOK, wantCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			r := gin.New()
			r.Use(tt.caller.middleware())
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate", handler.UpdateCertificate)

			enrollmentID := uuid.New()
			certificateURL := "https://example.com/certificates/cert.pdf"
			mockService.On("GetByID", mock.Anything, enrollmentID).Return(&models.Inscricao{
				ID:      enrollmentID,
				CursoID: 1,
				CPF:     statusAuthOwnerCPF,
			}, nil).Maybe()
			if tt.wantCall {
				mockService.On("UpdateCertificate", mock.Anything, 1, enrollmentID, certificateURL).Return(nil).Once()
			}

			body, _ := json.Marshal(map[string]interface{}{"certificate_url": certificateURL})
			req := httptest.NewRequest(http.MethodPut,
				"/api/v1/courses/1/enrollments/"+enrollmentID.String()+"/certificate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantCall {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "UpdateCertificate",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

// TestInscricaoHandler_EnrollmentRoutes_EditorOwnerWithRBAC wires the real
// CourseEnrollmentSingleAccess middleware (RBAC_ENABLED=true) in front of the
// handlers. The middleware lets the enrollment owner through before the orgao
// check, so an editor of another orgao who owns the enrollment must still be
// held to the owner rules by the handler.
func TestInscricaoHandler_EnrollmentRoutes_EditorOwnerWithRBAC(t *testing.T) {
	const courseOrgao = "orgao-a"

	editorOtherOrgaoOwner := statusAuthCaller{role: "USER", roles: []string{"go:cursos:editor"}, cpf: statusAuthOwnerCPF}
	editorOtherOrgao := statusAuthCaller{role: "USER", roles: []string{"go:cursos:editor"}, cpf: statusAuthOtherCPF}
	editorSameOrgao := statusAuthCaller{role: "USER", roles: []string{"go:cursos:editor"}, cpf: statusAuthOtherCPF}

	tests := []struct {
		name       string
		caller     statusAuthCaller
		groups     []string
		path       string
		body       map[string]interface{}
		wantCode   int
		wantUpdate string
	}{
		{
			name:     "editor of another orgao who owns the enrollment cannot approve it",
			caller:   editorOtherOrgaoOwner,
			groups:   []string{"go:orgao:orgao-b"},
			path:     "/status",
			body:     map[string]interface{}{"status": "approved", "admin_notes": "x"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "editor of another orgao who owns the enrollment cannot set the certificate",
			caller:   editorOtherOrgaoOwner,
			groups:   []string{"go:orgao:orgao-b"},
			path:     "/certificate",
			body:     map[string]interface{}{"certificate_url": "https://example.com/cert.pdf"},
			wantCode: http.StatusForbidden,
		},
		{
			name:       "editor of another orgao who owns the enrollment may cancel it",
			caller:     editorOtherOrgaoOwner,
			groups:     []string{"go:orgao:orgao-b"},
			path:       "/status",
			body:       map[string]interface{}{"status": "cancelled", "admin_notes": "x"},
			wantCode:   http.StatusOK,
			wantUpdate: "status",
		},
		{
			name:     "editor of another orgao is refused by the middleware",
			caller:   editorOtherOrgao,
			groups:   []string{"go:orgao:orgao-b"},
			path:     "/status",
			body:     map[string]interface{}{"status": "approved"},
			wantCode: http.StatusForbidden,
		},
		{
			name:       "editor of the course orgao approves",
			caller:     editorSameOrgao,
			groups:     []string{"go:orgao:" + courseOrgao},
			path:       "/status",
			body:       map[string]interface{}{"status": "approved", "admin_notes": "ok"},
			wantCode:   http.StatusOK,
			wantUpdate: "status",
		},
		{
			name:       "editor of the course orgao sets the certificate",
			caller:     editorSameOrgao,
			groups:     []string{"go:orgao:" + courseOrgao},
			path:       "/certificate",
			body:       map[string]interface{}{"certificate_url": "https://example.com/cert.pdf"},
			wantCode:   http.StatusOK,
			wantUpdate: "certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			mockService := new(MockInscricaoService)
			handler := v1.NewInscricaoHandler(mockService, new(MockJobService), new(MockCursoRepository))

			courseLoader := func(_ context.Context, _ int) (string, bool, error) {
				return courseOrgao, true, nil
			}
			cpfLoader := func(_ context.Context, _ uuid.UUID) (string, bool, error) {
				return statusAuthOwnerCPF, true, nil
			}
			access := middlewares.CourseEnrollmentSingleAccess(courseLoader, cpfLoader)

			r := gin.New()
			r.Use(tt.caller.middleware(), func(c *gin.Context) {
				c.Set(middlewares.UserGroupsKey, tt.groups)
				c.Next()
			})
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/status", access, handler.UpdateIndividualStatus)
			r.PUT("/api/v1/courses/:courseId/enrollments/:enrollmentId/certificate", access, handler.UpdateCertificate)

			enrollmentID := uuid.New()
			mockService.On("GetByID", mock.Anything, enrollmentID).Return(&models.Inscricao{
				ID:      enrollmentID,
				CursoID: 1,
				CPF:     statusAuthOwnerCPF,
			}, nil).Maybe()
			switch tt.wantUpdate {
			case "status":
				mockService.On("UpdateStatus", mock.Anything, enrollmentID, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Once()
			case "certificate":
				mockService.On("UpdateCertificate", mock.Anything, 1, enrollmentID, mock.Anything).Return(nil).Once()
			}

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut,
				"/api/v1/courses/1/enrollments/"+enrollmentID.String()+tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			switch tt.wantUpdate {
			case "status":
				mockService.AssertExpectations(t)
				if tt.caller.cpf == statusAuthOwnerCPF {
					mockService.AssertCalled(t, "UpdateStatus", mock.Anything, enrollmentID,
						models.StatusInscricaoCancelled, "", "")
				}
			case "certificate":
				mockService.AssertExpectations(t)
			default:
				mockService.AssertNotCalled(t, "UpdateStatus",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockService.AssertNotCalled(t, "UpdateCertificate",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
