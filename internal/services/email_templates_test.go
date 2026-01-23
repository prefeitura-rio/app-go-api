package services

import (
	"strings"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
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
