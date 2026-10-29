package empregabilidade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeParticipacaoRepo struct {
	participacao *empregabilidade.ParticipacaoBanco
	atualizacao  *time.Time
	err          error

	chamadas []chamadaParticipacao
}

type chamadaParticipacao struct {
	cpf        string
	disponivel bool
	quando     time.Time
}

func (f *fakeParticipacaoRepo) GetParticipacaoBanco(_ context.Context, _ string) (*empregabilidade.ParticipacaoBanco, error) {
	return f.participacao, f.err
}

func (f *fakeParticipacaoRepo) AtualizarParticipacaoBanco(_ context.Context, cpf string, disponivel bool, quando time.Time) error {
	if f.err != nil {
		return f.err
	}
	f.chamadas = append(f.chamadas, chamadaParticipacao{cpf: cpf, disponivel: disponivel, quando: quando})
	origem := string(empregabilidade.OrigemCadastroBancoCurriculos)
	f.participacao = &empregabilidade.ParticipacaoBanco{Disponivel: disponivel, AlteradoEm: &quando, OrigemCadastro: &origem}
	return nil
}

func (f *fakeParticipacaoRepo) GetUltimaAtualizacao(_ context.Context, _ string) (*time.Time, error) {
	return f.atualizacao, nil
}

type fakeCurriculoGetter struct {
	curriculo *empregabilidade.CurriculoCompleto
	err       error
}

func (f *fakeCurriculoGetter) GetCurriculoCompleto(_ context.Context, _ string) (*empregabilidade.CurriculoCompleto, error) {
	return f.curriculo, f.err
}

const cpfCidadao = "11144477735"

var agoraParticipacao = time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)

func novoParticipacaoService(repo *fakeParticipacaoRepo, curriculo *empregabilidade.CurriculoCompleto, termosAceitos bool) *services.ParticipacaoBancoService {
	s := services.NewParticipacaoBancoService(repo, &fakeCurriculoGetter{curriculo: curriculo}, &fakeTermos{aceitos: termosAceitos})
	s.SetRelogio(func() time.Time { return agoraParticipacao })
	return s
}

func TestParticipacaoBancoService_Get_SemCurriculo(t *testing.T) {
	s := novoParticipacaoService(&fakeParticipacaoRepo{}, nil, false)

	status, err := s.Get(context.Background(), cpfCidadao)

	require.NoError(t, err)
	assert.False(t, status.TemCurriculo)
	assert.False(t, status.CurriculoCompleto)
	assert.False(t, status.Disponivel)
	assert.Equal(t, []string{"Situação atual não informada", "Termos de uso não aceitos"}, status.Pendencias)
}

func TestParticipacaoBancoService_Get_CurriculoCompletoDesligado(t *testing.T) {
	atualizado := agoraParticipacao.Add(-48 * time.Hour)
	origem := string(empregabilidade.OrigemCadastroVaga)
	repo := &fakeParticipacaoRepo{
		participacao: &empregabilidade.ParticipacaoBanco{OrigemCadastro: &origem},
		atualizacao:  &atualizado,
	}
	s := novoParticipacaoService(repo, curriculoMinimo(), true)

	status, err := s.Get(context.Background(), cpfCidadao)

	require.NoError(t, err)
	assert.True(t, status.TemCurriculo)
	assert.True(t, status.CurriculoCompleto)
	assert.Empty(t, status.Pendencias)
	assert.NotNil(t, status.Pendencias, "lista vazia, não null, para o app")
	assert.False(t, status.Disponivel)
	assert.Equal(t, &atualizado, status.CurriculoAtualizadoEm)
	assert.Equal(t, &origem, status.OrigemCadastro)
}

func TestParticipacaoBancoService_Atualizar_AtivaCurriculoCompleto(t *testing.T) {
	repo := &fakeParticipacaoRepo{participacao: &empregabilidade.ParticipacaoBanco{}}
	s := novoParticipacaoService(repo, curriculoMinimo(), true)

	status, err := s.Atualizar(context.Background(), cpfCidadao, true)

	require.NoError(t, err)
	require.Len(t, repo.chamadas, 1)
	assert.Equal(t, chamadaParticipacao{cpf: cpfCidadao, disponivel: true, quando: agoraParticipacao}, repo.chamadas[0])
	assert.True(t, status.Disponivel)
	assert.Equal(t, &agoraParticipacao, status.DisponivelAlteradoEm)
}

// O banco mostra o que a pessoa conseguiu salvar: currículo incompleto também
// pode voltar a ficar visível.
func TestParticipacaoBancoService_Atualizar_AtivaMesmoComCurriculoIncompleto(t *testing.T) {
	repo := &fakeParticipacaoRepo{participacao: &empregabilidade.ParticipacaoBanco{}}
	s := novoParticipacaoService(repo, nil, false)

	status, err := s.Atualizar(context.Background(), cpfCidadao, true)

	require.NoError(t, err)
	require.Len(t, repo.chamadas, 1)
	assert.True(t, repo.chamadas[0].disponivel)
	assert.True(t, status.Disponivel)
	assert.False(t, status.CurriculoCompleto)
}

func TestParticipacaoBancoService_Atualizar_RecusaSemCurriculo(t *testing.T) {
	for _, disponivel := range []bool{true, false} {
		repo := &fakeParticipacaoRepo{}
		s := novoParticipacaoService(repo, nil, false)

		_, err := s.Atualizar(context.Background(), cpfCidadao, disponivel)

		require.ErrorIs(t, err, services.ErrSemCurriculoParaBanco)
		assert.Empty(t, repo.chamadas)
	}
}

func TestParticipacaoBancoService_Atualizar_DesativaMesmoComCurriculoIncompleto(t *testing.T) {
	repo := &fakeParticipacaoRepo{participacao: &empregabilidade.ParticipacaoBanco{Disponivel: true}}
	s := novoParticipacaoService(repo, nil, false)

	status, err := s.Atualizar(context.Background(), cpfCidadao, false)

	require.NoError(t, err)
	require.Len(t, repo.chamadas, 1)
	assert.False(t, repo.chamadas[0].disponivel)
	assert.False(t, status.Disponivel)
}

func TestParticipacaoBancoService_Get_ErroDoRepositorio(t *testing.T) {
	s := novoParticipacaoService(&fakeParticipacaoRepo{err: errors.New("falhou")}, nil, false)

	_, err := s.Get(context.Background(), cpfCidadao)

	require.Error(t, err)
}
