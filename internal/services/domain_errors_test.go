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
	t.Run("falha de banco segue sem tipo", func(t *testing.T) {
		dbFail := &MockInscricaoRepository{
			GetByIDFunc: func(context.Context, uuid.UUID) (*models.Inscricao, error) { return nil, errors.New("conn reset") },
		}
		svc := services.NewInscricaoServiceWithInterface(dbFail, openCursoRepo(), nil, nil, nil, cfg)
		_, ok := apperrors.As(svc.Delete(ctx, uuid.New()))
		assert.False(t, ok)
	})
}
