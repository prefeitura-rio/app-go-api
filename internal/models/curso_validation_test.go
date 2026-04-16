package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModalidade_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		modalidade Modalidade
		expected   bool
	}{
		{"valid_presencial", ModalidadePresencial, true},
		{"valid_remoto", ModalidadeRemoto, true},
		{"valid_hibrido", ModalidadeHibrido, true},
		{"valid_online", ModalidadeOnline, true},
		{"valid_livre_formacao_online", ModalidadeLivreFormacaoOnline, true},
		{"invalid_empty", "", false},
		{"invalid_value", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.modalidade.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModalidade_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Modalidade
		expected Modalidade
	}{
		{"presencial_uppercase_maps", "PRESENCIAL", ModalidadePresencial},
		{"online_maps_to_remoto", "ONLINE", ModalidadeRemoto},
		{"hibrido_maps_to_semipresencial", "HIBRIDO", ModalidadeSemipresencial},
		{"already_normalized", ModalidadePresencial, ModalidadePresencial},
		{"unknown_stays_same", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Normalize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusCurso_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   StatusCurso
		expected bool
	}{
		{"valid_draft", StatusCursoDraft, true},
		{"valid_opened", StatusCursoOpened, true},
		{"valid_closed", StatusCursoClosed, true},
		{"valid_canceled", StatusCursoCanceled, true},
		{"invalid_empty", "", false},
		{"invalid_value", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusCurso_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    StatusCurso
		expected StatusCurso
	}{
		{"criado_maps_to_draft", "CRIADO", StatusCursoDraft},
		{"aberto_maps_to_published", "ABERTO", StatusCursoPublished},
		{"encerrado_maps_to_closed", "ENCERRADO", StatusCursoClosed},
		{"opened_maps_to_published", StatusCursoOpened, StatusCursoPublished},
		{"unknown_stays_same", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Normalize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: FormatoAula and Turno don't have IsValid/Normalize methods or constants defined
// Skipping tests for these types until they are implemented

func TestCourseManagementType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cmt      CourseManagementType
		expected bool
	}{
		{"valid_own_org", CourseManagementOwnOrg, true},
		{"valid_external_managed_by_org", CourseManagementExternalManagedByOrg, true},
		{"valid_external_managed_by_partner", CourseManagementExternalManagedByPartner, true},
		{"empty_should_be_valid", "", true},
		{"invalid_value", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmt.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatusCurso_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     StatusCurso
		to       StatusCurso
		expected bool
	}{
		// draft transitions
		{StatusCursoDraft, StatusCursoInReview, true},
		{StatusCursoDraft, StatusCursoPendingDeletion, true},
		{StatusCursoDraft, StatusCursoApproved, false},
		{StatusCursoDraft, StatusCursoPublished, false},
		// needs_changes transitions
		{StatusCursoNeedsChanges, StatusCursoInReview, true},
		{StatusCursoNeedsChanges, StatusCursoPendingDeletion, true},
		{StatusCursoNeedsChanges, StatusCursoApproved, false},
		// in_review transitions
		{StatusCursoInReview, StatusCursoApproved, true},
		{StatusCursoInReview, StatusCursoPendingDeletion, true},
		{StatusCursoInReview, StatusCursoPublished, false},
		{StatusCursoInReview, StatusCursoDraft, false},
		// approved transitions
		{StatusCursoApproved, StatusCursoPublished, true},
		{StatusCursoApproved, StatusCursoPendingDeletion, true},
		{StatusCursoApproved, StatusCursoInReview, false},
		// published transitions
		{StatusCursoPublished, StatusCursoNeedsChanges, true},
		{StatusCursoPublished, StatusCursoPendingDeletion, true},
		{StatusCursoPublished, StatusCursoClosed, true},
		{StatusCursoPublished, StatusCursoCanceled, true},
		{StatusCursoPublished, StatusCursoDraft, false},
		// closed/canceled → pending_deletion
		{StatusCursoClosed, StatusCursoPendingDeletion, true},
		{StatusCursoClosed, StatusCursoPublished, false},
		{StatusCursoCanceled, StatusCursoPendingDeletion, true},
		{StatusCursoCanceled, StatusCursoPublished, false},
		// legacy values normalize correctly
		{StatusCursoOpened, StatusCursoNeedsChanges, true},    // opened → published → needs_changes
		{StatusCursoOpened, StatusCursoPendingDeletion, true},  // opened → published → pending_deletion
		{StatusCursoAberto, StatusCursoNeedsChanges, true},    // ABERTO → published → needs_changes
		{StatusCursoCriado, StatusCursoInReview, true},        // CRIADO → draft → in_review
	}

	for _, tt := range tests {
		name := string(tt.from) + "_to_" + string(tt.to)
		t.Run(name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}
