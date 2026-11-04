package services

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// DEV-238: a candidatura carregada sem a Vaga (ex.: aprovação/reprovação em lote)
// derrubava o processo com nil pointer dereference ao montar o e-mail.

func candidaturaSemVaga() *empregabilidade.Candidatura {
	nome := "Pessoa Sintética"
	email := "pessoa.sintetica@example.com"
	return &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: &email,
		CPF:   "00000000191",
	}
}

func TestSendCandidaturaReprovadaEmail_NilVaga_SkipsWithoutPanic(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(mockClient, nil, nil, nil, true, "oportunidades.rio")

	var err error
	require.NotPanics(t, func() {
		err = service.SendCandidaturaReprovadaEmail(context.Background(), candidaturaSemVaga())
	})
	assert.NoError(t, err)
	assert.Empty(t, mockClient.sentEmails, "sem vaga não há como montar o e-mail; deve ser pulado")
}

func TestSendCandidaturaAprovadaEmail_NilVaga_SkipsWithoutPanic(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(mockClient, nil, nil, nil, true, "oportunidades.rio")

	var err error
	require.NotPanics(t, func() {
		err = service.SendCandidaturaAprovadaEmail(context.Background(), candidaturaSemVaga())
	})
	assert.NoError(t, err)
	assert.Empty(t, mockClient.sentEmails, "sem vaga não há como montar o e-mail; deve ser pulado")
}

func TestSendCandidaturaAprovadaEmail_NilContratante_SendsWithoutPanic(t *testing.T) {
	mockClient := &MockDataRelayClient{}
	service := NewEmailNotificationService(mockClient, nil, nil, nil, true, "oportunidades.rio")

	candidatura := candidaturaSemVaga()
	candidatura.Vaga = &empregabilidade.Vaga{ID: uuid.New(), Titulo: "Vaga Sintética"}

	var err error
	require.NotPanics(t, func() {
		err = service.SendCandidaturaAprovadaEmail(context.Background(), candidatura)
	})
	require.NoError(t, err)
	require.Len(t, mockClient.sentEmails, 1)
	assert.True(t, strings.Contains(mockClient.sentEmails[0].Body, "Vaga Sintética"))
}

func TestGetCandidaturaReprovadaEmailTemplate_NilVaga_NoPanic(t *testing.T) {
	var tpl EmailTemplate
	require.NotPanics(t, func() {
		tpl = GetCandidaturaReprovadaEmailTemplate(candidaturaSemVaga(), nil, "oportunidades.rio")
	})
	assert.NotEmpty(t, tpl.Subject)
	assert.Contains(t, tpl.Body, "Pessoa Sintética")
}

func TestGetCandidaturaReprovadaEmailTemplate_NilCandidatura_NoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		GetCandidaturaReprovadaEmailTemplate(nil, &empregabilidade.Vaga{Titulo: "Vaga"}, "oportunidades.rio")
	})
}

func TestGetCandidaturaAprovadaEmailTemplate_NilVagaEmpresa_NoPanic(t *testing.T) {
	var tpl EmailTemplate
	require.NotPanics(t, func() {
		tpl = GetCandidaturaAprovadaEmailTemplate(candidaturaSemVaga(), nil, nil)
	})
	assert.NotEmpty(t, tpl.Subject)
	assert.Contains(t, tpl.Body, "Pessoa Sintética")
}
