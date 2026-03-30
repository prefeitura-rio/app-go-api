package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrolledUnit_Value(t *testing.T) {
	tests := []struct {
		name        string
		unit        EnrolledUnit
		expectNil   bool
		expectError bool
	}{
		{
			name: "valid_enrolled_unit",
			unit: EnrolledUnit{
				ID:               "unit-123",
				CursoID:          1,
				Address:          "Rua das Flores, 123",
				Neighborhood:     "Centro",
				NeighborhoodZone: "Zona Sul",
				Schedules: []EnrolledUnitSchedule{
					{
						ID:                 "schedule-1",
						LocationID:         "loc-1",
						Vacancies:          30,
						ClassStartDate:     "2024-01-01",
						ClassEndDate:       "2024-12-31",
						ClassTime:          "14:00-18:00",
						ClassDays:          "Segunda, Quarta",
						RemainingVacancies: 10,
					},
				},
				CreatedAt: "2024-01-01T10:00:00Z",
				UpdatedAt: "2024-01-01T10:00:00Z",
			},
			expectNil:   false,
			expectError: false,
		},
		{
			name:        "empty_unit_returns_nil",
			unit:        EnrolledUnit{},
			expectNil:   true,
			expectError: false,
		},
		{
			name: "unit_with_only_id",
			unit: EnrolledUnit{
				ID: "unit-456",
			},
			expectNil:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.unit.Value()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, value)
			} else {
				assert.NotNil(t, value)

				// Verify it's valid JSON
				var result map[string]interface{}
				err := json.Unmarshal(value.([]byte), &result)
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnrolledUnit_Scan(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		validate    func(t *testing.T, unit *EnrolledUnit)
	}{
		{
			name:        "nil_value",
			input:       nil,
			expectError: false,
			validate: func(t *testing.T, unit *EnrolledUnit) {
				assert.Equal(t, "", unit.ID)
			},
		},
		{
			name:        "empty_bytes",
			input:       []byte{},
			expectError: false,
			validate: func(t *testing.T, unit *EnrolledUnit) {
				assert.Equal(t, "", unit.ID)
			},
		},
		{
			name: "valid_json",
			input: []byte(`{
				"id": "unit-789",
				"curso_id": 5,
				"address": "Av. Brasil, 1000",
				"neighborhood": "Penha",
				"schedules": []
			}`),
			expectError: false,
			validate: func(t *testing.T, unit *EnrolledUnit) {
				assert.Equal(t, "unit-789", unit.ID)
				assert.Equal(t, 5, unit.CursoID)
				assert.Equal(t, "Av. Brasil, 1000", unit.Address)
				assert.Equal(t, "Penha", unit.Neighborhood)
			},
		},
		{
			name:        "invalid_json",
			input:       []byte(`{invalid json`),
			expectError: true,
			validate:    nil,
		},
		{
			name:        "non_byte_slice",
			input:       "string value",
			expectError: true,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var unit EnrolledUnit
			err := unit.Scan(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, &unit)
				}
			}
		})
	}
}

func TestCustomField_TableName(t *testing.T) {
	field := &CustomField{}
	assert.Equal(t, "custom_fields", field.TableName())
}

func TestCourseSchedule_TableName(t *testing.T) {
	schedule := &CourseSchedule{}
	assert.Equal(t, "course_schedules", schedule.TableName())
}

func TestRemoteSchedule_TableName(t *testing.T) {
	schedule := &RemoteSchedule{}
	assert.Equal(t, "remote_schedules", schedule.TableName())
}

func TestLocationClass_TableName(t *testing.T) {
	location := &LocationClass{}
	assert.Equal(t, "location_classes", location.TableName())
}

func TestRemoteClass_TableName(t *testing.T) {
	remote := &RemoteClass{}
	assert.Equal(t, "remote_classes", remote.TableName())
}

func TestEnrollmentSummary_Structure(t *testing.T) {
	t.Run("all_fields_initialized", func(t *testing.T) {
		summary := EnrollmentSummary{
			Total:     100,
			Pending:   20,
			Approved:  50,
			Rejected:  10,
			Cancelled: 5,
			Concluded: 15,
		}

		assert.Equal(t, 100, summary.Total)
		assert.Equal(t, 20, summary.Pending)
		assert.Equal(t, 50, summary.Approved)
		assert.Equal(t, 10, summary.Rejected)
		assert.Equal(t, 5, summary.Cancelled)
		assert.Equal(t, 15, summary.Concluded)
	})

	t.Run("zero_values", func(t *testing.T) {
		summary := EnrollmentSummary{}

		assert.Equal(t, 0, summary.Total)
		assert.Equal(t, 0, summary.Pending)
		assert.Equal(t, 0, summary.Approved)
		assert.Equal(t, 0, summary.Rejected)
		assert.Equal(t, 0, summary.Cancelled)
		assert.Equal(t, 0, summary.Concluded)
	})
}

func TestEnrollmentStatusUpdateRequest_Validation(t *testing.T) {
	t.Run("valid_request", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()

		req := EnrollmentStatusUpdateRequest{
			EnrollmentIDs: []uuid.UUID{id1, id2},
			Status:        StatusInscricaoApproved,
			Reason:        "Approved by admin",
			AdminNotes:    "All requirements met",
		}

		assert.Len(t, req.EnrollmentIDs, 2)
		assert.Equal(t, StatusInscricaoApproved, req.Status)
		assert.Equal(t, "Approved by admin", req.Reason)
		assert.Equal(t, "All requirements met", req.AdminNotes)
	})

	t.Run("multiple_statuses", func(t *testing.T) {
		statuses := []StatusInscricao{
			StatusInscricaoPending,
			StatusInscricaoApproved,
			StatusInscricaoRejected,
			StatusInscricaoCancelled,
			StatusInscricaoConcluded,
		}

		for _, status := range statuses {
			req := EnrollmentStatusUpdateRequest{
				EnrollmentIDs: []uuid.UUID{uuid.New()},
				Status:        status,
			}

			assert.Equal(t, status, req.Status)
		}
	})
}

func TestInscricaoUpdateRequest_Structure(t *testing.T) {
	t.Run("all_fields_present", func(t *testing.T) {
		name := "João Silva"
		email := "joao@example.com"
		phone := "21987654321"
		age := 30
		address := "Rua A, 123"
		neighborhood := "Centro"
		adminNotes := "Updated by admin"

		enrolledUnit := &EnrolledUnit{
			ID:      "unit-1",
			CursoID: 1,
		}

		req := InscricaoUpdateRequest{
			Name:         &name,
			Email:        &email,
			Phone:        &phone,
			Age:          &age,
			Address:      &address,
			Neighborhood: &neighborhood,
			AdminNotes:   &adminNotes,
			EnrolledUnit: enrolledUnit,
		}

		assert.Equal(t, name, *req.Name)
		assert.Equal(t, email, *req.Email)
		assert.Equal(t, phone, *req.Phone)
		assert.Equal(t, age, *req.Age)
		assert.Equal(t, address, *req.Address)
		assert.Equal(t, neighborhood, *req.Neighborhood)
		assert.Equal(t, adminNotes, *req.AdminNotes)
		assert.NotNil(t, req.EnrolledUnit)
	})

	t.Run("partial_update", func(t *testing.T) {
		name := "Maria Santos"

		req := InscricaoUpdateRequest{
			Name: &name,
		}

		assert.Equal(t, name, *req.Name)
		assert.Nil(t, req.Email)
		assert.Nil(t, req.Phone)
		assert.Nil(t, req.Age)
	})
}

func TestScheduleChangeRequest_Structure(t *testing.T) {
	t.Run("valid_schedule_change", func(t *testing.T) {
		scheduleID := uuid.New()
		enrolledUnit := &EnrolledUnit{
			ID:      "unit-2",
			CursoID: 3,
		}

		req := ScheduleChangeRequest{
			ScheduleID:   &scheduleID,
			EnrolledUnit: enrolledUnit,
		}

		assert.NotNil(t, req.ScheduleID)
		assert.Equal(t, scheduleID, *req.ScheduleID)
		require.NotNil(t, req.EnrolledUnit)
		assert.Equal(t, "unit-2", req.EnrolledUnit.ID)
		assert.Equal(t, 3, req.EnrolledUnit.CursoID)
	})

	t.Run("nil_schedule_id", func(t *testing.T) {
		enrolledUnit := &EnrolledUnit{
			ID: "unit-3",
		}

		req := ScheduleChangeRequest{
			ScheduleID:   nil,
			EnrolledUnit: enrolledUnit,
		}

		assert.Nil(t, req.ScheduleID)
		assert.NotNil(t, req.EnrolledUnit)
	})
}

func TestEnrolledUnitSchedule_Structure(t *testing.T) {
	t.Run("complete_schedule", func(t *testing.T) {
		schedule := EnrolledUnitSchedule{
			ID:                 "sched-1",
			LocationID:         "loc-1",
			Vacancies:          50,
			ClassStartDate:     "2024-02-01",
			ClassEndDate:       "2024-06-30",
			ClassTime:          "09:00-12:00",
			ClassDays:          "Terça, Quinta",
			CreatedAt:          "2024-01-15T10:00:00Z",
			UpdatedAt:          "2024-01-15T10:00:00Z",
			RemainingVacancies: 25,
		}

		assert.Equal(t, "sched-1", schedule.ID)
		assert.Equal(t, "loc-1", schedule.LocationID)
		assert.Equal(t, 50, schedule.Vacancies)
		assert.Equal(t, 25, schedule.RemainingVacancies)
		assert.Equal(t, "2024-02-01", schedule.ClassStartDate)
		assert.Equal(t, "2024-06-30", schedule.ClassEndDate)
		assert.Equal(t, "09:00-12:00", schedule.ClassTime)
		assert.Equal(t, "Terça, Quinta", schedule.ClassDays)
	})
}
