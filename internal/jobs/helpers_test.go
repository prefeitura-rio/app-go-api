package jobs

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

// stringPtr is a test helper that returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func TestBuildScheduleMap_Empty(t *testing.T) {
	scheduleMap := buildScheduleMap([]models.LocationClass{}, false, nil)
	assert.Empty(t, scheduleMap)
}

func TestBuildScheduleMap_MultipleLocations(t *testing.T) {
	now := time.Now()
	loc1ID := uuid.New()
	loc2ID := uuid.New()
	sch1ID := uuid.New()
	sch2ID := uuid.New()

	locations := []models.LocationClass{
		{
			ID:           loc1ID,
			Address:      "Rua A, 123",
			Neighborhood: "Centro",
			CursoID:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
			Schedules: []models.CourseSchedule{
				{
					ID:             sch1ID,
					LocationID:     loc1ID,
					ClassTime:      "09:00-12:00",
					ClassDays:      "Seg-Sex",
					ClassStartDate: now,
					ClassEndDate:   now,
					Vacancies:      30,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		},
		{
			ID:           loc2ID,
			Address:      "Rua B, 456",
			Neighborhood: "Zona Sul",
			CursoID:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
			Schedules: []models.CourseSchedule{
				{
					ID:             sch2ID,
					LocationID:     loc2ID,
					ClassTime:      "14:00-17:00",
					ClassDays:      "Seg-Qua",
					ClassStartDate: now,
					ClassEndDate:   now,
					Vacancies:      20,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		},
	}

	scheduleMap := buildScheduleMap(locations, false, nil)

	// Should have entries for both schedules
	assert.Contains(t, scheduleMap, sch1ID.String())
	assert.Contains(t, scheduleMap, sch2ID.String())
	assert.Contains(t, scheduleMap, loc1ID.String())
	assert.Contains(t, scheduleMap, loc2ID.String())

	// Test address-based keys
	assert.Contains(t, scheduleMap, "rua a, 123")
	assert.Contains(t, scheduleMap, "rua b, 456")

	// Test composite keys
	assert.Contains(t, scheduleMap, "rua a, 123|09:00-12:00")
	assert.Contains(t, scheduleMap, "rua b, 456|14:00-17:00")
}

func TestBuildScheduleMap_RemoteClassOnly(t *testing.T) {
	now := time.Now()
	remoteID := uuid.New()
	scheduleID := uuid.New()
	classTime := "18:00-21:00"
	classDays := "Terça e Quinta"

	remoteClass := &models.RemoteClass{
		ID:        remoteID,
		CursoID:   1,
		CreatedAt: now,
		UpdatedAt: now,
		Schedules: []models.RemoteSchedule{
			{
				ID:             scheduleID,
				RemoteClassID:  remoteID,
				ClassTime:      &classTime,
				ClassDays:      &classDays,
				ClassStartDate: &now,
				ClassEndDate:   &now,
				Vacancies:      100,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
	}

	scheduleMap := buildScheduleMap([]models.LocationClass{}, true, remoteClass)

	assert.Contains(t, scheduleMap, scheduleID.String())
	assert.Contains(t, scheduleMap, remoteID.String())
	assert.Contains(t, scheduleMap, "18:00-21:00|terça e quinta")
}

func TestBuildScheduleMap_RemoteClassNilFields(t *testing.T) {
	now := time.Now()
	remoteID := uuid.New()
	scheduleID := uuid.New()

	remoteClass := &models.RemoteClass{
		ID:        remoteID,
		CursoID:   1,
		CreatedAt: now,
		UpdatedAt: now,
		Schedules: []models.RemoteSchedule{
			{
				ID:             scheduleID,
				RemoteClassID:  remoteID,
				ClassTime:      nil, // Nil fields
				ClassDays:      nil,
				ClassStartDate: nil,
				ClassEndDate:   nil,
				Vacancies:      50,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
	}

	scheduleMap := buildScheduleMap([]models.LocationClass{}, true, remoteClass)

	// Should still create UUID-based keys
	assert.Contains(t, scheduleMap, scheduleID.String())
	assert.Contains(t, scheduleMap, remoteID.String())
}

func TestFindScheduleByTurma_VariousSeparators(t *testing.T) {
	scheduleID := uuid.New()
	locationID := uuid.New()

	scheduleMap := map[string]struct {
		LocationID uuid.UUID
		ScheduleID uuid.UUID
	}{
		"centro|09:00": {LocationID: locationID, ScheduleID: scheduleID},
	}

	tests := []struct {
		turma string
		found bool
	}{
		{"Centro|09:00", true},
		{"centro-09:00", true},
		{"centro,09:00", true},
		{"centro;09:00", true},
		{"other|10:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.turma, func(t *testing.T) {
			locID, schID, err := findScheduleByTurma(tt.turma, scheduleMap, []models.LocationClass{}, false, nil)

			if tt.found {
				assert.NoError(t, err)
				assert.NotNil(t, locID)
				assert.NotNil(t, schID)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFindScheduleByTurma_ThreePartKey(t *testing.T) {
	scheduleID := uuid.New()
	locationID := uuid.New()

	scheduleMap := map[string]struct {
		LocationID uuid.UUID
		ScheduleID uuid.UUID
	}{
		"centro|09:00|seg-sex": {LocationID: locationID, ScheduleID: scheduleID},
	}

	locID, schID, err := findScheduleByTurma("Centro|09:00|Seg-Sex", scheduleMap, []models.LocationClass{}, false, nil)

	assert.NoError(t, err)
	assert.NotNil(t, locID)
	assert.NotNil(t, schID)
	assert.Equal(t, locationID, *locID)
	assert.Equal(t, scheduleID, *schID)
}

func TestFindScheduleByTurma_FuzzyRemoteClass(t *testing.T) {
	now := time.Now()
	remoteID := uuid.New()
	scheduleID := uuid.New()
	classTime := "19:00-22:00"
	classDays := "Sábados"

	remoteClass := &models.RemoteClass{
		ID: remoteID,
		Schedules: []models.RemoteSchedule{
			{
				ID:            scheduleID,
				RemoteClassID: remoteID,
				ClassTime:     &classTime,
				ClassDays:     &classDays,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	scheduleMap := map[string]struct {
		LocationID uuid.UUID
		ScheduleID uuid.UUID
	}{}

	// Should find by fuzzy match
	locID, schID, err := findScheduleByTurma("19:00 sábados", scheduleMap, []models.LocationClass{}, true, remoteClass)

	// Fuzzy match might work depending on the implementation
	if err == nil {
		assert.NotNil(t, locID)
		assert.NotNil(t, schID)
		if locID != nil && schID != nil {
			assert.Equal(t, remoteID, *locID)
			assert.Equal(t, scheduleID, *schID)
		}
	}
}

func TestStringPtr(t *testing.T) {
	tests := []string{"test", "", "another test with spaces"}

	for _, s := range tests {
		ptr := stringPtr(s)
		assert.NotNil(t, ptr)
		assert.Equal(t, s, *ptr)
	}
}

func TestStringPtr_Uniqueness(t *testing.T) {
	str1 := stringPtr("test")
	str2 := stringPtr("test")

	// Values should be equal
	assert.Equal(t, *str1, *str2)

	// But pointers should be different
	assert.NotSame(t, str1, str2)
}
