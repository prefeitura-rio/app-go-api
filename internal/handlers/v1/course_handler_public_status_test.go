package v1_test

import (
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

// curationStatuses are the stored statuses the public course detail answers 404
// for (draft and its legacy CRIADO, in_review, needs_changes, approved,
// pending_deletion). The public listing must never return them either.
var curationStatuses = []string{
	string(models.StatusCursoDraft),
	string(models.StatusCursoCriado),
	string(models.StatusCursoInReview),
	string(models.StatusCursoNeedsChanges),
	string(models.StatusCursoApproved),
	string(models.StatusCursoPendingDeletion),
}

// listFilterFor runs one request against the given route and returns the filter
// the handler passed to CursoService.List.
func listFilterFor(t *testing.T, public bool, target string) map[string]interface{} {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)
	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	if public {
		r.GET("/api/public/courses", handler.ListPublic)
	} else {
		r.GET("/api/v1/courses", handler.List)
	}

	var captured map[string]interface{}
	mockService.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			captured = args.Get(1).(map[string]interface{})
		}).
		Return([]*models.Curso{}, 0, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, captured, "CursoService.List must be called")
	return captured
}

// The public listing must hide every course still in the curation pipeline, with
// or without ?status=, so it shows exactly what the public detail shows.
func TestCourseHandler_ListPublic_HidesCurationStatuses(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		wantStatusIn []string // nil = no "status IN" filter
	}{
		{"no status param", "/api/public/courses", nil},
		{"status=draft is ignored", "/api/public/courses?status=draft", nil},
		{
			"every curation status is ignored",
			"/api/public/courses?status=draft,CRIADO,in_review,needs_changes,approved,pending_deletion",
			nil,
		},
		{
			"curation statuses are dropped, public ones kept",
			"/api/public/courses?status=draft,published,in_review,accepting_enrollments",
			[]string{"published", "accepting_enrollments"},
		},
		{"invalid status falls back to the public listing", "/api/public/courses?status=foo", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := listFilterFor(t, true, tt.target)

			notIn, ok := filter["status NOT IN"].([]string)
			require.True(t, ok, "public listing must exclude curation statuses, filter=%v", filter)
			assert.ElementsMatch(t, curationStatuses, notIn)

			if tt.wantStatusIn == nil {
				assert.NotContains(t, filter, "status IN")
			} else {
				assert.Equal(t, tt.wantStatusIn, filter["status IN"])
			}

			_, visible := filter["is_visible NOT_FALSE"]
			assert.True(t, visible, "public listing keeps hiding is_visible=false courses")
		})
	}
}

// The admin listing keeps its behavior: no status param hides only drafts, and
// ?status= may ask for any status, curation ones included.
func TestCourseHandler_List_Admin_KeepsCurationStatuses(t *testing.T) {
	t.Run("no status param excludes only drafts", func(t *testing.T) {
		filter := listFilterFor(t, false, "/api/v1/courses")

		assert.Equal(t, models.StatusCursoDraft, filter["status NOT"])
		assert.NotContains(t, filter, "status NOT IN")
		assert.NotContains(t, filter, "status IN")
	})

	t.Run("status=draft,in_review returns them", func(t *testing.T) {
		filter := listFilterFor(t, false, "/api/v1/courses?status=draft,in_review")

		assert.Equal(t, []string{"draft", "in_review"}, filter["status IN"])
		assert.NotContains(t, filter, "status NOT IN")
		assert.NotContains(t, filter, "status NOT")
	})
}
