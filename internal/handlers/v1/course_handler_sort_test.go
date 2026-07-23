package v1_test

import (
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

// Availability sorting must place courses a citizen can still enroll in ahead of the
// rest (finished/closed/canceled, past deadline, or sold out — including
// accepting_enrollments with zero vacancies), globally, before paginating.
func TestCourseHandler_List_SortByAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/public/courses", handler.ListPublic)

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	classStart := now.Add(-1 * time.Hour)
	classEnd := now.Add(48 * time.Hour)

	schedAvailable := uuid.New() // has vacancies
	schedSoldOut := uuid.New()   // accepting_enrollments but zero vacancies

	openCourse := func(id int, schedID uuid.UUID, vacancies int) *models.Curso {
		return &models.Curso{
			ID:                  id,
			Titulo:              "Course",
			Status:              models.StatusCursoPublished,
			Modalidade:          models.ModalidadePresencial,
			EnrollmentStartDate: &past,
			EnrollmentEndDate:   &future,
			LocationClasses: []models.LocationClass{
				{Schedules: []models.CourseSchedule{
					{ID: schedID, Vacancies: vacancies, ClassStartDate: classStart, ClassEndDate: classEnd},
				}},
			},
		}
	}

	// Repository returns rows in id-desc order: 3 (sold out), 2 (closed), 1 (available).
	cursos := []*models.Curso{
		openCourse(3, schedSoldOut, 5),
		{ID: 2, Titulo: "Closed", Status: models.StatusCursoClosed, Modalidade: models.ModalidadePresencial},
		openCourse(1, schedAvailable, 10),
	}

	// limit <= 0 signals "return every match" for the sorted path.
	mockRepo.On("List", mock.Anything, mock.Anything, -1, 0).Return(cursos, 3, nil)
	// schedSoldOut is fully booked (5/5); schedAvailable has no enrollments.
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(map[uuid.UUID]int64{schedSoldOut: 5}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/public/courses?sort=availability&page=1&limit=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data struct {
			Courses []struct {
				ID int `json:"id"`
			} `json:"courses"`
			Pagination struct {
				Total      int `json:"total"`
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Page 1 (limit 2): available course (id 1) first, then id 3 preserving id-desc
	// within the "unavailable" group (the sold-out accepting course sinks below).
	require.Len(t, response.Data.Courses, 2)
	assert.Equal(t, 1, response.Data.Courses[0].ID, "available course must come first")
	assert.Equal(t, 3, response.Data.Courses[1].ID)
	assert.Equal(t, 3, response.Data.Pagination.Total)
	assert.Equal(t, 2, response.Data.Pagination.TotalPages)

	// The sorted path must not fall back to the paginated service.List.
	mockService.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

// Without the sort param, the handler keeps the existing behavior: it delegates to
// the paginated service.List and does not fetch the full result set.
func TestCourseHandler_List_NoSort_UsesService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCursoService)
	mockInscricaoService := new(MockInscricaoServiceForCourse)
	mockRepo := new(MockCursoRepositoryForCourseHandler)

	handler := v1.NewCourseHandler(mockService, mockInscricaoService, mockRepo)

	r := gin.New()
	r.GET("/api/public/courses", handler.ListPublic)

	cursos := []*models.Curso{{ID: 1, Titulo: "Course 1"}}
	mockService.On("List", mock.Anything, mock.Anything, 1, 10).Return(cursos, 1, nil)
	mockRepo.On("CountEnrollmentsByScheduleIDs", mock.Anything, mock.Anything).
		Return(make(map[uuid.UUID]int64), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/public/courses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	// The unsorted path must not fetch the full result set via repo.List.
	mockRepo.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
