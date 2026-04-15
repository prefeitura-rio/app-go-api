package models_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestModalidade_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		modal models.Modalidade
		want  bool
	}{
		{"Presencial", models.ModalidadePresencial, true},
		{"Semipresencial", models.ModalidadeSemipresencial, true},
		{"Remoto", models.ModalidadeRemoto, true},
		{"Legacy Presencial", models.ModalidadePresencialLegacy, true},
		{"Legacy Online", models.ModalidadeOnline, true},
		{"Legacy Hibrido", models.ModalidadeHibrido, true},
		{"Livre Formacao Online", models.ModalidadeLivreFormacaoOnline, true},
		{"Invalid", models.Modalidade("INVALID"), false},
		{"Empty", models.Modalidade(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.modal.IsValid(); got != tt.want {
				t.Errorf("Modalidade.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCurso_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status models.StatusCurso
		want   bool
	}{
		{"Draft", models.StatusCursoDraft, true},
		{"Opened", models.StatusCursoOpened, true},
		{"Closed", models.StatusCursoClosed, true},
		{"Canceled", models.StatusCursoCanceled, true},
		{"Legacy Criado", models.StatusCursoCriado, true},
		{"Legacy Aberto", models.StatusCursoAberto, true},
		{"Legacy Encerrado", models.StatusCursoEncerrado, true},
		{"Invalid", models.StatusCurso("INVALID"), false},
		{"Empty", models.StatusCurso(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("StatusCurso.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatoAula_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		formato models.FormatoAula
		want    bool
	}{
		{"Gravado", models.FormatoAulaGravado, true},
		{"Ao Vivo", models.FormatoAulaAoVivo, true},
		{"Empty is valid", models.FormatoAula(""), true},
		{"Invalid", models.FormatoAula("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.formato.IsValid(); got != tt.want {
				t.Errorf("FormatoAula.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourseManagementType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		mgmtType models.CourseManagementType
		want     bool
	}{
		{"OwnOrg", models.CourseManagementOwnOrg, true},
		{"External Managed By Org", models.CourseManagementExternalManagedByOrg, true},
		{"External Managed By Partner", models.CourseManagementExternalManagedByPartner, true},
		{"Empty is valid", models.CourseManagementType(""), true},
		{"Invalid", models.CourseManagementType("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mgmtType.IsValid(); got != tt.want {
				t.Errorf("CourseManagementType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCurso_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from models.StatusCurso
		to   models.StatusCurso
		want bool
	}{
		// draft transitions
		{"draft → in_review", models.StatusCursoDraft, models.StatusCursoInReview, true},
		{"draft → pending_deletion", models.StatusCursoDraft, models.StatusCursoPendingDeletion, true},
		{"draft → approved (invalid)", models.StatusCursoDraft, models.StatusCursoApproved, false},
		{"draft → published (invalid)", models.StatusCursoDraft, models.StatusCursoPublished, false},

		// in_review transitions
		{"in_review → approved", models.StatusCursoInReview, models.StatusCursoApproved, true},
		{"in_review → needs_changes", models.StatusCursoInReview, models.StatusCursoNeedsChanges, true},
		{"in_review → pending_deletion", models.StatusCursoInReview, models.StatusCursoPendingDeletion, true},
		{"in_review → published (invalid)", models.StatusCursoInReview, models.StatusCursoPublished, false},
		{"in_review → draft (invalid)", models.StatusCursoInReview, models.StatusCursoDraft, false},

		// needs_changes transitions
		{"needs_changes → in_review", models.StatusCursoNeedsChanges, models.StatusCursoInReview, true},
		{"needs_changes → pending_deletion", models.StatusCursoNeedsChanges, models.StatusCursoPendingDeletion, true},
		{"needs_changes → approved (invalid)", models.StatusCursoNeedsChanges, models.StatusCursoApproved, false},
		{"needs_changes → published (invalid)", models.StatusCursoNeedsChanges, models.StatusCursoPublished, false},

		// approved transitions
		{"approved → published", models.StatusCursoApproved, models.StatusCursoPublished, true},
		{"approved → pending_deletion", models.StatusCursoApproved, models.StatusCursoPendingDeletion, true},
		{"approved → in_review (invalid)", models.StatusCursoApproved, models.StatusCursoInReview, false},

		// published transitions
		{"published → needs_changes", models.StatusCursoPublished, models.StatusCursoNeedsChanges, true},
		{"published → pending_deletion", models.StatusCursoPublished, models.StatusCursoPendingDeletion, true},
		{"published → closed", models.StatusCursoPublished, models.StatusCursoClosed, true},
		{"published → canceled", models.StatusCursoPublished, models.StatusCursoCanceled, true},
		{"published → draft (invalid)", models.StatusCursoPublished, models.StatusCursoDraft, false},

		// closed transitions
		{"closed → pending_deletion", models.StatusCursoClosed, models.StatusCursoPendingDeletion, true},
		{"closed → published (invalid)", models.StatusCursoClosed, models.StatusCursoPublished, false},

		// canceled transitions
		{"canceled → pending_deletion", models.StatusCursoCanceled, models.StatusCursoPendingDeletion, true},
		{"canceled → published (invalid)", models.StatusCursoCanceled, models.StatusCursoPublished, false},

		// legacy status normalization
		{"CRIADO (draft) → in_review", models.StatusCursoCriado, models.StatusCursoInReview, true},
		{"ABERTO (published) → needs_changes", models.StatusCursoAberto, models.StatusCursoNeedsChanges, true},
		{"ENCERRADO (closed) → pending_deletion", models.StatusCursoEncerrado, models.StatusCursoPendingDeletion, true},
		{"ABERTO (published) → draft (invalid)", models.StatusCursoAberto, models.StatusCursoDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("StatusCurso(%q).CanTransitionTo(%q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestCurso_TableName(t *testing.T) {
	curso := &models.Curso{}
	if got := curso.TableName(); got != "cursos" {
		t.Errorf("Curso.TableName() = %v, want cursos", got)
	}
}
