package services

import (
	"strings"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

func TestGetEnrollmentApprovedEmailTemplate_WithScheduleInfo(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "João da Silva",
		Email: "joao@teste.com",
	}

	curso := &models.Curso{
		Titulo:     "Curso de Programação",
		Modalidade: models.ModalidadePresencial,
	}

	orgaoName := "Secretaria Municipal de Educação"

	scheduleInfo := &ScheduleInfo{
		ClassTime:      "14:00",
		ClassStartDate: "20/01/2026",
		ClassDays:      "Segunda, Quarta, Sexta",
		Address:        "Rua das Flores, 123 - Centro",
	}

	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, orgaoName, scheduleInfo, "oportunidades.rio")

	// Verificar se o nome do usuário está presente
	if !strings.Contains(template.Body, "João da Silva") {
		t.Error("Nome do usuário não encontrado no template")
	}

	// Verificar se o nome do órgão está presente (2x no template)
	if strings.Count(template.Body, orgaoName) < 2 {
		t.Errorf("Nome do órgão deveria aparecer 2 vezes, aparece %d", strings.Count(template.Body, orgaoName))
	}

	// Verificar se o horário do schedule está presente
	if !strings.Contains(template.Body, "20/01/2026 às 14:00") {
		t.Error("Horário de início do schedule não encontrado no template")
	}

	// Verificar se os dias da semana estão presentes
	if !strings.Contains(template.Body, "Segunda, Quarta, Sexta") {
		t.Error("Dias da semana não encontrados no template")
	}

	// Verificar se o endereço está presente
	if !strings.Contains(template.Body, "Rua das Flores, 123 - Centro") {
		t.Error("Endereço não encontrado no template")
	}

	// Verificar que NÃO está usando placeholder
	if strings.Contains(template.Body, "[puxar") {
		t.Error("Template ainda contém placeholder [puxar...]")
	}
}

func TestGetEnrollmentApprovedEmailTemplate_WithoutScheduleInfo_FallbackToCurso(t *testing.T) {
	dataInicio := time.Date(2026, 1, 25, 9, 30, 0, 0, time.UTC)

	inscricao := &models.Inscricao{
		Name:  "Maria Santos",
		Email: "maria@teste.com",
	}

	curso := &models.Curso{
		Titulo:          "Curso de Excel",
		Modalidade:      models.ModalidadeRemoto,
		DataInicio:      &dataInicio,
		LocalRealizacao: "Online - Teams",
	}

	orgaoName := "Secretaria de Trabalho"

	// Sem schedule info - deve usar fallback do curso
	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, orgaoName, nil, "oportunidades.rio")

	// Verificar se usou o DataInicio do curso como fallback
	if !strings.Contains(template.Body, "25/01/2026 09:30") {
		t.Error("Fallback para DataInicio do curso não funcionou")
	}

	// Verificar se o nome do órgão está correto
	if !strings.Contains(template.Body, "Secretaria de Trabalho") {
		t.Error("Nome do órgão não encontrado no template")
	}
}

func TestGetEnrollmentApprovedEmailTemplate_OnlineCourse(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Pedro Souza",
		Email: "pedro@teste.com",
	}

	curso := &models.Curso{
		Titulo:     "Curso Online de Marketing",
		Modalidade: models.ModalidadeRemoto,
	}

	scheduleInfo := &ScheduleInfo{
		ClassTime:      "19:00",
		ClassStartDate: "15/02/2026",
		Address:        "online",
	}

	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, "SMTE", scheduleInfo, "oportunidades.rio")

	// Verificar endereço online
	if !strings.Contains(template.Body, "Endereço: online") {
		t.Error("Endereço online não encontrado")
	}

	// Verificar horário
	if !strings.Contains(template.Body, "15/02/2026 às 19:00") {
		t.Error("Horário não encontrado")
	}
}

func TestGetEnrollmentPendingEmailTemplate_WithOrgaoName(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Ana Paula",
		Email: "ana@teste.com",
	}

	curso := &models.Curso{
		Titulo: "Curso de Empreendedorismo",
	}

	orgaoName := "SEBRAE Rio"

	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, orgaoName, "oportunidades.rio")

	// Verificar se o nome do órgão está presente
	if !strings.Contains(template.Body, "SEBRAE Rio") {
		t.Error("Nome do órgão não encontrado no template pending")
	}

	// Verificar subject
	expectedSubject := "Inscrição recebida! - Curso de Empreendedorismo"
	if template.Subject != expectedSubject {
		t.Errorf("Subject incorreto. Esperado: %s, Recebido: %s", expectedSubject, template.Subject)
	}
}

func TestGetEnrollmentRejectedEmailTemplate(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:   "Carlos Oliveira",
		Email:  "carlos@teste.com",
		Reason: "Documentação incompleta",
	}

	curso := &models.Curso{
		Titulo: "Curso de Gestão",
	}

	template := GetEnrollmentRejectedEmailTemplate(inscricao, curso, "oportunidades.rio")

	// Verificar nome do usuário
	if !strings.Contains(template.Body, "Carlos Oliveira") {
		t.Error("Nome do usuário não encontrado no template de rejeição")
	}

	// Verificar título do curso
	if !strings.Contains(template.Body, "Curso de Gestão") {
		t.Error("Título do curso não encontrado no template de rejeição")
	}

	// Verificar mensagem de não aprovação
	if !strings.Contains(template.Body, "sua inscrição não foi aprovada") {
		t.Error("Mensagem de não aprovação não encontrada no template")
	}

	// Verificar subject
	expectedSubject := "Informações sobre sua inscrição - Curso de Gestão"
	if template.Subject != expectedSubject {
		t.Errorf("Subject incorreto. Esperado: %s, Recebido: %s", expectedSubject, template.Subject)
	}
}

func TestGetEnrollmentPendingEmailTemplate_WithOrgaoName2(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Fernanda Lima",
		Email: "fernanda@teste.com",
	}

	curso := &models.Curso{
		Titulo: "Curso de Informática",
	}

	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, "SMTI", "oportunidades.rio")

	// Verificar nome do usuário
	if !strings.Contains(template.Body, "Fernanda Lima") {
		t.Error("Nome do usuário não encontrado no template")
	}

	// Verificar título do curso
	if !strings.Contains(template.Body, "Curso de Informática") {
		t.Error("Título do curso não encontrado no template")
	}

	// Verificar orgão
	if !strings.Contains(template.Body, "SMTI") {
		t.Error("Nome do órgão não encontrado no template")
	}

	// Verificar subject
	expectedSubject := "Inscrição recebida! - Curso de Informática"
	if template.Subject != expectedSubject {
		t.Errorf("Subject incorreto. Esperado: %s, Recebido: %s", expectedSubject, template.Subject)
	}
}

func TestGetEnrollmentPendingEmailTemplate_WithOrgaoName3(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Roberto Costa",
		Email: "roberto@teste.com",
	}

	curso := &models.Curso{
		Titulo: "Curso de Segurança",
	}

	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, "SEOP", "oportunidades.rio")

	// Verificar nome do usuário
	if !strings.Contains(template.Body, "Roberto Costa") {
		t.Error("Nome do usuário não encontrado no template")
	}

	// Verificar orgão
	if !strings.Contains(template.Body, "SEOP") {
		t.Error("Nome do órgão não encontrado no template")
	}
}

func TestEmailTemplate_EmptyFields(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "",
		Email: "test@example.com",
	}

	curso := &models.Curso{
		Titulo: "",
	}

	// Should not panic with empty fields
	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, "", "oportunidades.rio")
	if template.Subject == "" {
		t.Error("Subject should not be empty")
	}
	if template.Body == "" {
		t.Error("Body should not be empty")
	}
}

func TestEmailTemplate_SpecialCharacters(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "José & María <test>",
		Email: "jose@example.com",
	}

	curso := &models.Curso{
		Titulo: "Curso \"Especial\" & Avançado",
	}

	// Should handle special characters without breaking
	template := GetEnrollmentPendingEmailTemplate(inscricao, curso, "Test & Co.", "oportunidades.rio")
	if template.Body == "" {
		t.Error("Body should not be empty with special characters")
	}

	// Verify special characters are preserved
	if !strings.Contains(template.Body, "José & María") {
		t.Error("Special characters in name should be preserved")
	}
}

func TestScheduleInfo_NilValues(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Test User",
		Email: "test@example.com",
	}

	curso := &models.Curso{
		Titulo:     "Test Course",
		Modalidade: models.ModalidadePresencial,
	}

	// Nil schedule info should use fallback
	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, "Test Org", nil, "oportunidades.rio")

	if template.Body == "" {
		t.Error("Template body should not be empty with nil schedule")
	}

	// Should still contain user name and course title
	if !strings.Contains(template.Body, "Test User") {
		t.Error("Should contain user name even without schedule info")
	}
	if !strings.Contains(template.Body, "Test Course") {
		t.Error("Should contain course title even without schedule info")
	}
}

func TestScheduleInfo_EmptyStrings(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Test User",
		Email: "test@example.com",
	}

	curso := &models.Curso{
		Titulo:     "Test Course",
		Modalidade: models.ModalidadePresencial,
	}

	scheduleInfo := &ScheduleInfo{
		ClassTime:      "",
		ClassStartDate: "",
		ClassDays:      "",
		Address:        "",
	}

	// Empty schedule info fields should not break template
	template := GetEnrollmentApprovedEmailTemplate(inscricao, curso, "Test Org", scheduleInfo, "oportunidades.rio")

	if template.Body == "" {
		t.Error("Template body should not be empty with empty schedule fields")
	}
}

func TestGetScheduleChangedEmailTemplate(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "Usuário Teste",
		Email: "teste@example.com",
	}

	dataInicio := time.Date(2026, 5, 25, 9, 30, 0, 0, time.UTC)

	curso := &models.Curso{
		Titulo:          "Curso Teste",
		Modalidade:      models.ModalidadeRemoto,
		DataInicio:      &dataInicio,
		LocalRealizacao: "Online - Teams",
	}

	scheduleInfo := &ScheduleInfo{
		ClassTime:      "14:00",
		ClassStartDate: "20/01/2026",
		ClassDays:      "Segunda, Quarta, Sexta",
		Address:        "Rua das Flores, 123 - Centro",
	}

	template := GetScheduleChangedEmailTemplate(inscricao, curso, scheduleInfo, "Org Teste", "oportunidades.rio")

	if !strings.Contains(template.Body, "Usuário Teste") {
		t.Error("User name not found on template")
	}

	if !strings.Contains(template.Body, "Curso Teste") {
		t.Error("Course name not found on template")
	}

	if !strings.Contains(template.Body, "Org Teste") {
		t.Error("Organ name not found on template")
	}
}

func TestGetScheduleChangedEmailTemplate_EmptyStrings(t *testing.T) {
	inscricao := &models.Inscricao{
		Name:  "",
		Email: "",
	}

	dataInicio := time.Date(2026, 5, 25, 9, 30, 0, 0, time.UTC)

	curso := &models.Curso{
		Titulo:          "",
		Modalidade:      models.ModalidadeRemoto,
		DataInicio:      &dataInicio,
		LocalRealizacao: "",
	}

	scheduleInfo := &ScheduleInfo{
		ClassTime:      "",
		ClassStartDate: "",
		ClassDays:      "",
		Address:        "",
	}

	template := GetScheduleChangedEmailTemplate(inscricao, curso, scheduleInfo, "Org Teste", "oportunidades.rio")

	if template.Body == "" {
		t.Error("Template body should not be empty with empty strings")
	}
}

func TestGetCandidaturaEnviadaEmailTemplate(t *testing.T) {
	nome := "Usuário Teste"
	email := "test@example.com"

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "Vaga Teste",
	}

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	template := GetCandidaturaEnviadaEmailTemplate(candidatura, vaga, empresa, "Org Teste", "oportunidades.rio")

	if !strings.Contains(template.Body, "Usuário Teste") {
		t.Error("User name not found on template")
	}

	if !strings.Contains(template.Body, "Vaga Teste") {
		t.Error("Position name not found on template")
	}

	if !strings.Contains(template.Body, "Empresa Teste") {
		t.Error("Company name not found on template")
	}

	if !strings.Contains(template.Body, "Org Teste") {
		t.Error("Organ name not found on template")
	}
}

func TestGetCandidaturaEnviadaEmailTemplate_EmptyStrings(t *testing.T) {
	nome := ""
	email := ""

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "",
	}

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "",
	}

	template := GetCandidaturaEnviadaEmailTemplate(candidatura, vaga, empresa, "Org Teste", "oportunidades.rio")

	if template.Body == "" {
		t.Error("Template body should not be empty with empty strings")
	}
}

func TestGetCandidaturaAprovadaEmailTemplate(t *testing.T) {
	nome := "Usuário Teste"
	email := "test@example.com"

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "Vaga Teste",
	}

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "Empresa Teste",
	}

	template := GetCandidaturaAprovadaEmailTemplate(candidatura, vaga, empresa)

	if !strings.Contains(template.Body, "Usuário Teste") {
		t.Error("User name not found on template")
	}

	if !strings.Contains(template.Body, "Vaga Teste") {
		t.Error("Position name not found on template")
	}

	if !strings.Contains(template.Body, "Empresa Teste") {
		t.Error("Company name not found on template")
	}
}

func TestGetCandidaturaAprovadaEmailTemplate_EmptyStrings(t *testing.T) {
	nome := ""
	email := ""

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "",
	}

	empresa := &empregabilidade.Empresa{
		NomeFantasia: "",
	}

	template := GetCandidaturaAprovadaEmailTemplate(candidatura, vaga, empresa)

	if template.Body == "" {
		t.Error("Template body should not be empty with empty strings")
	}
}

func TestGetCandidaturaReprovadaEmailTemplate(t *testing.T) {
	nome := "Usuário Teste"
	email := "test@example.com"

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "Vaga Teste",
	}
	template := GetCandidaturaReprovadaEmailTemplate(candidatura, vaga, "oportunidades.rio")

	if !strings.Contains(template.Body, "Usuário Teste") {
		t.Error("User name not found on template")
	}

	if !strings.Contains(template.Body, "Vaga Teste") {
		t.Error("Position name not found on template")
	}
}

func TestGetCandidaturaReprovadaEmailTemplate_EmptyStrings(t *testing.T) {
	nome := ""
	email := ""

	candidatura := &empregabilidade.Candidatura{
		Nome:  &nome,
		Email: &email,
	}

	vaga := &empregabilidade.Vaga{
		Titulo: "",
	}

	template := GetCandidaturaReprovadaEmailTemplate(candidatura, vaga, "oportunidades.rio")

	if template.Body == "" {
		t.Error("Template body should not be empty with empty strings")
	}
}
