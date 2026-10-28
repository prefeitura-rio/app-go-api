package empregabilidade

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// ErrSemCurriculoParaBanco indica que o cidadão ainda não salvou nenhuma parte do
// currículo, então não há o que mostrar ou ocultar no banco de currículos.
var ErrSemCurriculoParaBanco = errors.New("cidadão ainda não tem currículo")

// ParticipacaoBancoRepositoryInterface guarda a escolha do cidadão no registro
// raiz do currículo.
type ParticipacaoBancoRepositoryInterface interface {
	GetParticipacaoBanco(ctx context.Context, cpf string) (*empregabilidade.ParticipacaoBanco, error)
	AtualizarParticipacaoBanco(ctx context.Context, cpf string, disponivel bool, quando time.Time) error
	GetUltimaAtualizacao(ctx context.Context, cpf string) (*time.Time, error)
}

// ParticipacaoBancoService atende o card do banco de currículos no app: quem já
// salvou alguma parte do currículo escolhe se os recrutadores podem consultá-lo.
// O banco mostra o que a pessoa conseguiu salvar, então a escolha não depende do
// currículo mínimo; as pendências servem só para o app sugerir que termine.
type ParticipacaoBancoService struct {
	repo       ParticipacaoBancoRepositoryInterface
	curriculos CurriculoCompletoGetterInterface
	termos     AceiteTermosCheckerInterface
	agora      func() time.Time
}

func NewParticipacaoBancoService(
	repo ParticipacaoBancoRepositoryInterface,
	curriculos CurriculoCompletoGetterInterface,
	termos AceiteTermosCheckerInterface,
) *ParticipacaoBancoService {
	return &ParticipacaoBancoService{repo: repo, curriculos: curriculos, termos: termos, agora: time.Now}
}

// SetRelogio troca a hora usada para registrar a mudança (testes).
func (s *ParticipacaoBancoService) SetRelogio(agora func() time.Time) {
	s.agora = agora
}

// Get diz se o currículo está no banco e o que falta para o currículo mínimo.
func (s *ParticipacaoBancoService) Get(ctx context.Context, cpf string) (*empregabilidade.ParticipacaoBancoStatus, error) {
	participacao, err := s.repo.GetParticipacaoBanco(ctx, cpf)
	if err != nil {
		return nil, err
	}
	pendencias, err := s.pendencias(ctx, cpf)
	if err != nil {
		return nil, err
	}
	atualizacao, err := s.repo.GetUltimaAtualizacao(ctx, cpf)
	if err != nil {
		return nil, err
	}

	status := &empregabilidade.ParticipacaoBancoStatus{
		TemCurriculo:          participacao != nil,
		CurriculoCompleto:     len(pendencias) == 0,
		Pendencias:            pendencias,
		CurriculoAtualizadoEm: atualizacao,
	}
	if participacao != nil {
		status.Disponivel = participacao.Disponivel
		status.DisponivelAlteradoEm = participacao.AlteradoEm
		status.OrigemCadastro = participacao.OrigemCadastro
	}
	return status, nil
}

// Atualizar liga ou desliga o currículo no banco.
func (s *ParticipacaoBancoService) Atualizar(ctx context.Context, cpf string, disponivel bool) (*empregabilidade.ParticipacaoBancoStatus, error) {
	participacao, err := s.repo.GetParticipacaoBanco(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if participacao == nil {
		return nil, ErrSemCurriculoParaBanco
	}
	if err := s.repo.AtualizarParticipacaoBanco(ctx, cpf, disponivel, s.agora()); err != nil {
		return nil, err
	}
	return s.Get(ctx, cpf)
}

func (s *ParticipacaoBancoService) pendencias(ctx context.Context, cpf string) ([]string, error) {
	curriculo, err := s.curriculos.GetCurriculoCompleto(ctx, cpf)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar currículo: %w", err)
	}
	aceitos, err := s.termos.HasAcceptedTerms(ctx, cpf)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar termos de uso: %w", err)
	}
	pendencias := pendenciasCurriculoMinimo(curriculo, aceitos)
	if pendencias == nil {
		pendencias = []string{}
	}
	return pendencias, nil
}
