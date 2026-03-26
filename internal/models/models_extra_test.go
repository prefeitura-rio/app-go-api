package models

import (
	"testing"

	"github.com/google/uuid"
)

// ──────────────────────────────────────────────────────────────────────────────
// Emprego model tests
// ──────────────────────────────────────────────────────────────────────────────

func TestEmprego_TableName(t *testing.T) {
	e := Emprego{}
	if e.TableName() != "empregos" {
		t.Errorf("expected 'empregos', got %q", e.TableName())
	}
}

func TestTipoContratacao_IsValid(t *testing.T) {
	valid := []TipoContratacao{
		TipoContratacaoCLT,
		TipoContratacaoEstagio,
		TipoContratacaoJovemAprendiz,
		TipoContratacaoMEI,
		TipoContratacaoPJ,
	}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("expected %q to be valid", v)
		}
	}
}

func TestTipoContratacao_IsInvalid(t *testing.T) {
	invalid := []TipoContratacao{"", "CONTRATO", "OUTRO"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestStatusEmprego_IsValid(t *testing.T) {
	valid := []StatusEmprego{
		StatusEmpregoCriado,
		StatusEmpregoAberto,
		StatusEmpregoEmAnalise,
		StatusEmpregoEncerrado,
	}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("expected %q to be valid", v)
		}
	}
}

func TestStatusEmprego_IsInvalid(t *testing.T) {
	invalid := []StatusEmprego{"", "ATIVO", "INATIVO"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestJornadaTrabalho_IsValid(t *testing.T) {
	valid := []JornadaTrabalho{
		JornadaIntegral,
		JornadaMeioPeriodo,
		JornadaFlexivel,
		JornadaEstagio,
		JornadaHomeOffice,
	}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("expected %q to be valid", v)
		}
	}
}

func TestJornadaTrabalho_IsInvalid(t *testing.T) {
	invalid := []JornadaTrabalho{"", "PARCIAL", "FULL_TIME"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestEmprego_Validate_Success(t *testing.T) {
	e := &Emprego{
		Titulo:          "Desenvolvedor",
		TipoContratacao: TipoContratacaoCLT,
		Status:          StatusEmpregoAberto,
	}
	if err := e.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestEmprego_Validate_MissingTitulo(t *testing.T) {
	e := &Emprego{
		TipoContratacao: TipoContratacaoCLT,
		Status:          StatusEmpregoAberto,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing titulo")
	}
}

func TestEmprego_Validate_MissingTipoContratacao(t *testing.T) {
	e := &Emprego{
		Titulo: "Dev",
		Status: StatusEmpregoAberto,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing tipo contratacao")
	}
}

func TestEmprego_Validate_InvalidTipoContratacao(t *testing.T) {
	e := &Emprego{
		Titulo:          "Dev",
		TipoContratacao: "INVALIDO",
		Status:          StatusEmpregoAberto,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for invalid tipo contratacao")
	}
}

func TestEmprego_Validate_MissingStatus(t *testing.T) {
	e := &Emprego{
		Titulo:          "Dev",
		TipoContratacao: TipoContratacaoCLT,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing status")
	}
}

func TestEmprego_Validate_InvalidStatus(t *testing.T) {
	e := &Emprego{
		Titulo:          "Dev",
		TipoContratacao: TipoContratacaoCLT,
		Status:          "INVALIDO",
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestEmprego_Validate_InvalidJornada(t *testing.T) {
	e := &Emprego{
		Titulo:          "Dev",
		TipoContratacao: TipoContratacaoCLT,
		Status:          StatusEmpregoAberto,
		JornadaTrabalho: "INVALIDA",
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for invalid jornada")
	}
}

func TestEmprego_Validate_ValidJornada(t *testing.T) {
	e := &Emprego{
		Titulo:          "Dev",
		TipoContratacao: TipoContratacaoCLT,
		Status:          StatusEmpregoAberto,
		JornadaTrabalho: JornadaIntegral,
	}
	if err := e.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// PropostaMEI model tests
// ──────────────────────────────────────────────────────────────────────────────

func TestPropostaMEI_TableName(t *testing.T) {
	p := PropostaMEI{}
	if p.TableName() != "propostas_mei" {
		t.Error("expected 'propostas_mei'")
	}
}

func TestStatusPropostaAdmin_IsValid(t *testing.T) {
	valid := []StatusPropostaAdmin{
		StatusPropostaAdminDraft,
		StatusPropostaAdminActive,
		StatusPropostaAdminExpired,
	}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("expected %q to be valid", v)
		}
	}
}

func TestStatusPropostaAdmin_IsInvalid(t *testing.T) {
	invalid := []StatusPropostaAdmin{"", "pending", "ACTIVE"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestStatusPropostaCidadao_IsValid(t *testing.T) {
	valid := []StatusPropostaCidadao{
		StatusPropostaCidadaoSubmitted,
		StatusPropostaCidadaoApproved,
		StatusPropostaCidadaoRejected,
	}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("expected %q to be valid", v)
		}
	}
}

func TestStatusPropostaCidadao_IsInvalid(t *testing.T) {
	invalid := []StatusPropostaCidadao{"", "pending", "APPROVED"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestPropostaMEI_Validate_Success(t *testing.T) {
	valor := 1500.0
	p := &PropostaMEI{
		OportunidadeMEIID: 1,
		MEIEmpresaID:      "12.345.678/0001-90",
		ValorProposta:     &valor,
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestPropostaMEI_Validate_MissingOportunidade(t *testing.T) {
	p := &PropostaMEI{
		MEIEmpresaID: "12.345.678/0001-90",
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing oportunidade")
	}
}

func TestPropostaMEI_Validate_MissingMEIEmpresa(t *testing.T) {
	p := &PropostaMEI{
		OportunidadeMEIID: 1,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for missing MEI empresa")
	}
}

func TestPropostaMEI_Validate_NegativeValor(t *testing.T) {
	neg := -100.0
	p := &PropostaMEI{
		OportunidadeMEIID: 1,
		MEIEmpresaID:      "12.345.678/0001-90",
		ValorProposta:     &neg,
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative valor")
	}
}

func TestPropostaMEI_BeforeCreate_GeneratesID(t *testing.T) {
	p := &PropostaMEI{}
	if err := p.BeforeCreate(nil); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if p.ID == uuid.Nil {
		t.Error("expected non-nil UUID after BeforeCreate")
	}
}

func TestPropostaMEI_BeforeCreate_PreservesExistingID(t *testing.T) {
	existingID := uuid.New()
	p := &PropostaMEI{ID: existingID}
	if err := p.BeforeCreate(nil); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if p.ID != existingID {
		t.Errorf("expected ID to remain %v, got %v", existingID, p.ID)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Acessibilidade, Categoria, Empresa, Escolaridade, Instituicao, Job TableName tests
// ──────────────────────────────────────────────────────────────────────────────

func TestModelTableNames(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Empresa", Empresa{}.TableName(), "empresas"},
		{"Escolaridade", Escolaridade{}.TableName(), "escolaridades"},
		{"Instituicao", InstituicaoEnsino{}.TableName(), "instituicoes_ensino"},
		{"Job", Job{}.TableName(), "jobs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("TableName() = %q, want %q", tt.got, tt.expected)
			}
		})
	}
}
