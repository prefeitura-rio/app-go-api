package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-49: os serviços devolvem erros de domínio tipados, com as mesmas
// mensagens de antes, para o handler escolher o status HTTP sem comparar texto.

func assertDomainErr(t *testing.T, err error, kind apperrors.Kind, message string) {
	t.Helper()
	appErr, ok := apperrors.As(err)
	if !assert.True(t, ok, "esperava erro de domínio, veio %T: %v", err, err) {
		return
	}
	assert.Equal(t, kind, appErr.Kind)
	assert.Equal(t, message, appErr.Error())
}

func openCursoRepo() *MockCursoRepository {
	return &MockCursoRepository{
		ValidateForEnrollmentFunc: func(context.Context, int) (string, *time.Time, *time.Time, bool, error) {
			return string(models.StatusCursoOpened), nil, nil, false, nil
		},
	}
}

func TestInscricaoService_DomainErrors(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AppConfig{}
	alreadyEnrolled := &MockInscricaoRepository{
		ExistsByCPFAndCursoFunc: func(context.Context, string, int) (bool, error) { return true, nil },
	}
	missing := &MockInscricaoRepository{
		GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) { return nil, nil },
	}
	otherCourse := &MockInscricaoRepository{
		GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) {
			return &models.Inscricao{CursoID: 2, Status: models.StatusInscricaoApproved}, nil
		},
	}
	pending := &MockInscricaoRepository{
		GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) {
			return &models.Inscricao{CursoID: 1, Status: models.StatusInscricaoPending}, nil
		},
	}
	newInscricao := func() *models.Inscricao {
		return &models.Inscricao{CPF: "12345678900", CursoID: 1, Name: "Teste", Email: "teste@example.com"}
	}

	t.Run("Create duplicado é conflito", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(alreadyEnrolled, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.Create(ctx, newInscricao()), apperrors.KindConflict, "CPF já inscrito neste curso")
	})
	t.Run("CreateByAdmin duplicado é conflito", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(alreadyEnrolled, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.CreateByAdmin(ctx, newInscricao()), apperrors.KindConflict, "CPF já inscrito neste curso")
	})
	t.Run("CreateManual duplicado é conflito", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(alreadyEnrolled, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.CreateManual(ctx, newInscricao()), apperrors.KindConflict, "CPF já inscrito neste curso")
	})
	t.Run("UpdateStatus de inscrição inexistente", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(missing, openCursoRepo(), nil, nil, nil, cfg)
		err := svc.UpdateStatus(ctx, uuid.New(), models.StatusInscricaoApproved, "", "")
		assertDomainErr(t, err, apperrors.KindNotFound, "inscrição não encontrada")
	})
	t.Run("UpdateCertificate de inscrição inexistente", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(missing, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.UpdateCertificate(ctx, 1, uuid.New(), "https://x"), apperrors.KindNotFound, "inscrição não encontrada")
	})
	t.Run("UpdateCertificate de outro curso", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(otherCourse, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.UpdateCertificate(ctx, 1, uuid.New(), "https://x"), apperrors.KindInvalid, "inscrição não pertence ao curso especificado")
	})
	t.Run("UpdateCertificate de inscrição pendente", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(pending, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.UpdateCertificate(ctx, 1, uuid.New(), "https://x"), apperrors.KindInvalid, "certificado só pode ser atribuído a inscrições aprovadas ou concluídas")
	})
	t.Run("UpdateInscricao de inscrição inexistente", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(missing, openCursoRepo(), nil, nil, nil, cfg)
		err := svc.UpdateInscricao(ctx, uuid.New(), 1, &models.InscricaoUpdateRequest{})
		assertDomainErr(t, err, apperrors.KindNotFound, "inscrição não encontrada")
	})
	t.Run("UpdateInscricao de outro curso", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(otherCourse, openCursoRepo(), nil, nil, nil, cfg)
		err := svc.UpdateInscricao(ctx, uuid.New(), 1, &models.InscricaoUpdateRequest{})
		assertDomainErr(t, err, apperrors.KindInvalid, "inscrição não pertence ao curso especificado")
	})
	t.Run("Delete de inscrição inexistente", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(missing, openCursoRepo(), nil, nil, nil, cfg)
		assertDomainErr(t, svc.Delete(ctx, uuid.New()), apperrors.KindNotFound, "inscrição não encontrada")
	})
	t.Run("UpdateMultipleStatus sem seleção é Invalid", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(missing, openCursoRepo(), nil, nil, nil, cfg)
		_, err := svc.UpdateMultipleStatus(ctx, nil, models.StatusInscricaoApproved, "", "")
		assertDomainErr(t, err, apperrors.KindInvalid, "nenhuma inscrição selecionada")
	})
	t.Run("falha de banco segue sem tipo", func(t *testing.T) {
		dbFail := &MockInscricaoRepository{
			GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) { return nil, errors.New("conn reset") },
		}
		svc := services.NewInscricaoServiceWithInterface(dbFail, openCursoRepo(), nil, nil, nil, cfg)
		_, ok := apperrors.As(svc.Delete(ctx, uuid.New()))
		assert.False(t, ok)
	})
}

func TestOportunidadeMEIService_DomainErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("Create inválida é Invalid com a mensagem da validação", func(t *testing.T) {
		svc := services.NewOportunidadeMEIServiceWithInterface(&mockOportunidadeMEIRepo{})
		invalid := &models.OportunidadeMEI{}
		_, err := svc.Create(ctx, invalid, false)
		validationErr := invalid.Validate()
		assertDomainErr(t, err, apperrors.KindInvalid, validationErr.Error())
	})
	t.Run("Update de oportunidade inexistente", func(t *testing.T) {
		svc := services.NewOportunidadeMEIServiceWithInterface(&mockOportunidadeMEIRepo{})
		assertDomainErr(t, svc.Update(ctx, validOportunidade()), apperrors.KindNotFound, "oportunidade não encontrada")
	})
	t.Run("Publish de oportunidade inexistente", func(t *testing.T) {
		svc := services.NewOportunidadeMEIServiceWithInterface(&mockOportunidadeMEIRepo{})
		assertDomainErr(t, svc.Publish(ctx, 1), apperrors.KindNotFound, "oportunidade não encontrada")
	})
	t.Run("Publish incompleta é Invalid", func(t *testing.T) {
		incomplete := &models.OportunidadeMEI{ID: 1, Status: models.StatusOportunidadeDraft}
		svc := services.NewOportunidadeMEIServiceWithInterface(&mockOportunidadeMEIRepo{entity: incomplete})
		err := svc.Publish(ctx, 1)
		assert.True(t, apperrors.IsKind(err, apperrors.KindInvalid), "got %v", err)
		assert.Contains(t, err.Error(), "não é possível publicar: ")
	})
	t.Run("falha de banco segue sem tipo", func(t *testing.T) {
		svc := services.NewOportunidadeMEIServiceWithInterface(&mockOportunidadeMEIRepo{createErr: errors.New("db down")})
		_, err := svc.Create(ctx, validOportunidade(), false)
		_, ok := apperrors.As(err)
		assert.False(t, ok)
	})
}

func TestPropostaMEIService_DomainErrors(t *testing.T) {
	ctx := context.Background()
	newSvc := func(propostas *MockPropostaRepo, oportunidades *MockOportunidadeRepo, cnae *MockCNAEValidation) *services.PropostaMEIService {
		return services.NewPropostaMEIService(propostas, oportunidades, cnae, nil)
	}
	activeOportunidades := func(status models.StatusOportunidadeMEI) *MockOportunidadeRepo {
		repo := NewMockOportunidadeRepo()
		repo.oportunidades[1] = &models.OportunidadeMEI{ID: 1, Status: status, CNAEIDs: []string{"4110700"}}
		return repo
	}
	proposta := func() *models.PropostaMEI {
		return &models.PropostaMEI{OportunidadeMEIID: 1, MEIEmpresaID: "12345678000190"}
	}

	t.Run("Create inválida é Invalid", func(t *testing.T) {
		empty := &models.PropostaMEI{}
		validationErr := empty.Validate()
		_, err := newSvc(NewMockPropostaRepo(), NewMockOportunidadeRepo(), &MockCNAEValidation{}).Create(ctx, empty, "Bearer t")
		assertDomainErr(t, err, apperrors.KindInvalid, validationErr.Error())
	})
	t.Run("Create em oportunidade inexistente", func(t *testing.T) {
		_, err := newSvc(NewMockPropostaRepo(), NewMockOportunidadeRepo(), &MockCNAEValidation{}).Create(ctx, proposta(), "Bearer t")
		assertDomainErr(t, err, apperrors.KindNotFound, "oportunidade não encontrada")
	})
	t.Run("Create em oportunidade inativa", func(t *testing.T) {
		_, err := newSvc(NewMockPropostaRepo(), activeOportunidades(models.StatusOportunidadeDraft), &MockCNAEValidation{}).Create(ctx, proposta(), "Bearer t")
		assertDomainErr(t, err, apperrors.KindInvalid, "oportunidade não está ativa")
	})
	t.Run("Create duplicada é conflito", func(t *testing.T) {
		propostas := NewMockPropostaRepo()
		propostas.existingCheck = true
		_, err := newSvc(propostas, activeOportunidades(models.StatusOportunidadeActive), &MockCNAEValidation{}).Create(ctx, proposta(), "Bearer t")
		assertDomainErr(t, err, apperrors.KindConflict, "já existe uma proposta desta empresa para esta oportunidade")
	})
	t.Run("UpdateProposta inexistente", func(t *testing.T) {
		err := newSvc(NewMockPropostaRepo(), NewMockOportunidadeRepo(), &MockCNAEValidation{}).UpdateProposta(ctx, uuid.New(), 1, nil, nil, nil)
		assertDomainErr(t, err, apperrors.KindNotFound, "proposta não encontrada")
	})
	t.Run("UpdateProposta de outra oportunidade e valor negativo", func(t *testing.T) {
		propostas := NewMockPropostaRepo()
		id := uuid.New()
		propostas.propostas[id] = &models.PropostaMEI{ID: id, OportunidadeMEIID: 1}
		svc := newSvc(propostas, NewMockOportunidadeRepo(), &MockCNAEValidation{})

		assertDomainErr(t, svc.UpdateProposta(ctx, id, 2, nil, nil, nil), apperrors.KindInvalid, "proposta não pertence à oportunidade especificada")
		negative := -1.0
		assertDomainErr(t, svc.UpdateProposta(ctx, id, 1, &negative, nil, nil), apperrors.KindInvalid, "valor_proposta deve ser positivo")
	})
	t.Run("UpdateStatusCidadao inexistente", func(t *testing.T) {
		err := newSvc(NewMockPropostaRepo(), NewMockOportunidadeRepo(), &MockCNAEValidation{}).UpdateStatusCidadao(ctx, uuid.New(), models.StatusPropostaCidadaoApproved)
		assertDomainErr(t, err, apperrors.KindNotFound, "proposta não encontrada")
	})
	t.Run("UpdateMultipleStatus sem seleção ou com status inválido", func(t *testing.T) {
		svc := newSvc(NewMockPropostaRepo(), NewMockOportunidadeRepo(), &MockCNAEValidation{})
		_, err := svc.UpdateMultipleStatus(ctx, nil, models.StatusPropostaCidadaoApproved)
		assertDomainErr(t, err, apperrors.KindInvalid, "nenhuma proposta selecionada")
		_, err = svc.UpdateMultipleStatus(ctx, []uuid.UUID{uuid.New()}, models.StatusPropostaCidadao("xpto"))
		assertDomainErr(t, err, apperrors.KindInvalid, "status inválido")
	})
}

func TestCNAEValidationService_DomainErrors(t *testing.T) {
	ctx := context.Background()
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"
	company := []models.LegalEntity{{CNPJ: "12345678000190", CNAEFiscal: "4110700"}}

	tests := []struct {
		name    string
		rmi     *MockRMIClient
		token   string
		cnaes   []string
		kind    apperrors.Kind
		message string
	}{
		{"token sem CPF", &MockRMIClient{}, "invalid-token", []string{"4110700"}, apperrors.KindInvalid, "não foi possível extrair CPF do token de autenticação"},
		{"RMI fora", &MockRMIClient{err: errors.New("dial tcp: connection refused")}, token, []string{"4110700"}, apperrors.KindUnavailable, "não foi possível validar seus CNPJs no momento. Tente novamente em instantes"},
		{"sem CNPJ", &MockRMIClient{entities: []models.LegalEntity{}}, token, []string{"4110700"}, apperrors.KindForbidden, "nenhum CNPJ encontrado vinculado ao seu CPF"},
		{"CNPJ de outro CPF", &MockRMIClient{entities: []models.LegalEntity{{CNPJ: "98765432000199", CNAEFiscal: "4110700"}}}, token, []string{"4110700"}, apperrors.KindForbidden, "o CNPJ 12345678000190 não pertence ao seu CPF"},
		{"CNAE incompatível", &MockRMIClient{entities: company}, token, []string{"6201500"}, apperrors.KindForbidden, "o CNPJ 12345678000190 não possui CNAE compatível com esta oportunidade. CNAEs aceitos: 6201500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewCNAEValidationService(tt.rmi, NewMockCache())
			err := svc.ValidatePropostaForCNAE(ctx, tt.token, "12345678000190", tt.cnaes)
			assertDomainErr(t, err, tt.kind, tt.message)
		})
	}
}
