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

func TestCurso_TableName(t *testing.T) {
	curso := &models.Curso{}
	if got := curso.TableName(); got != "cursos" {
		t.Errorf("Curso.TableName() = %v, want cursos", got)
	}
}
