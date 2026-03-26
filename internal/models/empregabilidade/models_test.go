package empregabilidade

import (
	"testing"
	"time"
)

// StatusVaga tests

func TestStatusVaga_IsValid(t *testing.T) {
	valid := []StatusVaga{
		StatusVagaEmEdicao,
		StatusVagaEmAprovacao,
		StatusVagaPublicadoAtivo,
		StatusVagaPublicadoExpirado,
		StatusVagaCongelada,
		StatusVagaDescontinuada,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestStatusVaga_IsInvalid(t *testing.T) {
	invalid := []StatusVaga{"", "unknown", "ativo", "PUBLICADO"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

// AcessibilidadePCD tests

func TestAcessibilidadePCD_IsValid(t *testing.T) {
	valid := []AcessibilidadePCD{
		AcessibilidadeParaPCD,
		AcessibilidadePreferencialPCD,
		AcessibilidadeExclusivoPCD,
	}
	for _, a := range valid {
		if !a.IsValid() {
			t.Errorf("expected %q to be valid", a)
		}
	}
}

func TestAcessibilidadePCD_IsInvalid(t *testing.T) {
	invalid := []AcessibilidadePCD{"", "pcd", "invalid"}
	for _, a := range invalid {
		if a.IsValid() {
			t.Errorf("expected %q to be invalid", a)
		}
	}
}

// Vaga.UpdateStatusBasedOnExpiration tests

func TestVaga_UpdateStatusBasedOnExpiration_NilDataLimite(t *testing.T) {
	v := &Vaga{
		Status:     StatusVagaPublicadoAtivo,
		DataLimite: nil,
	}
	v.UpdateStatusBasedOnExpiration()
	if v.Status != StatusVagaPublicadoAtivo {
		t.Errorf("expected status to remain %q, got %q", StatusVagaPublicadoAtivo, v.Status)
	}
}

func TestVaga_UpdateStatusBasedOnExpiration_Expired_ActiveToExpired(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	v := &Vaga{
		Status:     StatusVagaPublicadoAtivo,
		DataLimite: &past,
	}
	v.UpdateStatusBasedOnExpiration()
	if v.Status != StatusVagaPublicadoExpirado {
		t.Errorf("expected status %q, got %q", StatusVagaPublicadoExpirado, v.Status)
	}
}

func TestVaga_UpdateStatusBasedOnExpiration_Expired_NonActiveUnchanged(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	v := &Vaga{
		Status:     StatusVagaEmEdicao,
		DataLimite: &past,
	}
	v.UpdateStatusBasedOnExpiration()
	if v.Status != StatusVagaEmEdicao {
		t.Errorf("expected status to remain %q, got %q", StatusVagaEmEdicao, v.Status)
	}
}

func TestVaga_UpdateStatusBasedOnExpiration_NotExpired_ExpiredToActive(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	v := &Vaga{
		Status:     StatusVagaPublicadoExpirado,
		DataLimite: &future,
	}
	v.UpdateStatusBasedOnExpiration()
	if v.Status != StatusVagaPublicadoAtivo {
		t.Errorf("expected status %q, got %q", StatusVagaPublicadoAtivo, v.Status)
	}
}

func TestVaga_UpdateStatusBasedOnExpiration_NotExpired_NonExpiredUnchanged(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	v := &Vaga{
		Status:     StatusVagaEmEdicao,
		DataLimite: &future,
	}
	v.UpdateStatusBasedOnExpiration()
	if v.Status != StatusVagaEmEdicao {
		t.Errorf("expected status to remain %q, got %q", StatusVagaEmEdicao, v.Status)
	}
}

// StatusCandidatura tests

func TestStatusCandidatura_IsValid(t *testing.T) {
	valid := []StatusCandidatura{
		StatusCandidaturaEnviada,
		StatusCandidaturaAprovada,
		StatusCandidaturaReprovada,
		StatusCandidaturaVagaCongelada,
		StatusCandidaturaDescontinuada,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestStatusCandidatura_IsInvalid(t *testing.T) {
	invalid := []StatusCandidatura{"", "unknown", "pendente"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

// StatusFormacao tests

func TestStatusFormacao_IsValid(t *testing.T) {
	valid := []StatusFormacao{
		StatusFormacaoCompleto,
		StatusFormacaoEmAndamento,
		StatusFormacaoIncompleto,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestStatusFormacao_IsInvalid(t *testing.T) {
	invalid := []StatusFormacao{"", "completo", "COMPLETO", "invalid"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

// CurriculoFormacao.Validate tests

func TestCurriculoFormacao_Validate_ValidStatus(t *testing.T) {
	f := &CurriculoFormacao{Status: StatusFormacaoCompleto}
	if err := f.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCurriculoFormacao_Validate_EmptyStatus(t *testing.T) {
	f := &CurriculoFormacao{Status: ""}
	if err := f.Validate(); err != nil {
		t.Errorf("expected no error for empty status, got %v", err)
	}
}

func TestCurriculoFormacao_Validate_InvalidStatus(t *testing.T) {
	f := &CurriculoFormacao{Status: "invalido"}
	if err := f.Validate(); err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

// TableName tests (ensures they compile and return correct values)

func TestTableNames(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Vaga", Vaga{}.TableName(), "emp_vagas"},
		{"Candidatura", Candidatura{}.TableName(), "emp_candidaturas"},
		{"CurriculoFormacao", CurriculoFormacao{}.TableName(), "emp_curriculo_formacoes"},
		{"CurriculoIdioma", CurriculoIdioma{}.TableName(), "emp_curriculo_idiomas"},
		{"CurriculoCursoComplementar", CurriculoCursoComplementar{}.TableName(), "emp_curriculo_cursos_complementares"},
		{"CurriculoExperiencia", CurriculoExperiencia{}.TableName(), "emp_curriculo_experiencias"},
		{"CurriculoConquista", CurriculoConquista{}.TableName(), "emp_curriculo_conquistas"},
		{"CurriculoSituacaoInteresses", CurriculoSituacaoInteresses{}.TableName(), "emp_curriculo_situacao_interesses"},
		{"Disponibilidade", Disponibilidade{}.TableName(), "emp_disponibilidades"},
		{"SituacaoAtual", SituacaoAtual{}.TableName(), "emp_situacoes_atual"},
		{"NivelIdioma", NivelIdioma{}.TableName(), "emp_niveis_idioma"},
		{"TipoPCD", TipoPCD{}.TableName(), "emp_tipos_pcd"},
		{"RegimeContratacao", RegimeContratacao{}.TableName(), "emp_regimes_contratacao"},
		{"ModeloTrabalho", ModeloTrabalho{}.TableName(), "emp_modelos_trabalho"},
		{"Idioma", Idioma{}.TableName(), "emp_idiomas"},
		{"Escolaridade", Escolaridade{}.TableName(), "emp_escolaridades"},
		{"TermosUso", TermosUso{}.TableName(), "emp_termos_uso"},
		{"TipoConquista", TipoConquista{}.TableName(), "emp_tipos_conquista"},
		{"Onboarding", Onboarding{}.TableName(), "emp_onboarding"},
		{"Empresa", Empresa{}.TableName(), "emp_empresas"},
		{"Etapa", Etapa{}.TableName(), "emp_etapas"},
		{"InformacaoComplementar", InformacaoComplementar{}.TableName(), "emp_informacoes_complementares"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("TableName() = %q, want %q", tt.got, tt.expected)
			}
		})
	}
}
