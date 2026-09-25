package empregabilidade

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// generosAutodeclarados são as opções de identidade de gênero do app do cidadão
// (as mesmas de ValidGenderOptions no app-rmi), indexadas em minúsculas.
// Quando a pessoa não declarou gênero, o cadastro guarda no mesmo campo o sexo da
// base cadastral ("M"/"F"), que não é identidade de gênero e fica de fora.
var generosAutodeclarados = indexarPorMinusculas(
	"Homem cisgênero",
	"Homem transgênero",
	"Mulher cisgênero",
	"Mulher transgênero",
	"Não binário",
	"Prefiro não informar",
	"Outro",
)

func indexarPorMinusculas(opcoes ...string) map[string]string {
	m := make(map[string]string, len(opcoes))
	for _, o := range opcoes {
		m[strings.ToLower(o)] = o
	}
	return m
}

// BancoCurriculoService serve o banco de currículos do backoffice: a listagem de
// todos os currículos da base e a ficha de cada um.
type BancoCurriculoService struct {
	repo                BancoCurriculoRepositoryInterface
	curriculos          CurriculoCompletoGetterInterface
	citizenSnapshotRepo CitizenSnapshotRepoForCandidaturaInterface
	citizenDataFetcher  CitizenDataFetcherForCandidaturaInterface // pode ser nil
}

func NewBancoCurriculoService(
	repo BancoCurriculoRepositoryInterface,
	curriculos CurriculoCompletoGetterInterface,
	citizenSnapshotRepo CitizenSnapshotRepoForCandidaturaInterface,
) *BancoCurriculoService {
	return &BancoCurriculoService{
		repo:                repo,
		curriculos:          curriculos,
		citizenSnapshotRepo: citizenSnapshotRepo,
	}
}

// SetCitizenDataFetcher liga o sync do cadastro depois da construção, quando o
// worker é iniciado no router.
func (s *BancoCurriculoService) SetCitizenDataFetcher(fetcher CitizenDataFetcherForCandidaturaInterface) {
	s.citizenDataFetcher = fetcher
}

func (s *BancoCurriculoService) List(ctx context.Context, filter empregabilidade.BancoCurriculoFilter, page, pageSize int) ([]*empregabilidade.BancoCurriculoItem, int64, error) {
	return s.repo.ListBancoCurriculos(ctx, filter, page, pageSize)
}

// GetDetalhe retorna a ficha do currículo, ou nil se o CPF não tem currículo.
func (s *BancoCurriculoService) GetDetalhe(ctx context.Context, cpf string) (*empregabilidade.BancoCurriculoDetalhe, error) {
	raiz, err := s.repo.GetCurriculoByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if raiz == nil {
		return nil, nil
	}

	curriculo, err := s.curriculos.GetCurriculoCompleto(ctx, cpf)
	if err != nil {
		return nil, err
	}

	detalhe := &empregabilidade.BancoCurriculoDetalhe{
		CPF:          raiz.CPF,
		DataInclusao: raiz.CreatedAt,
		Curriculo:    curriculo,
	}
	if curriculo != nil {
		detalhe.Profissao = profissaoDasExperiencias(curriculo.Experiencias)
	}

	atualizacao, err := s.repo.GetUltimaAtualizacao(ctx, cpf)
	if err != nil {
		return nil, err
	}
	detalhe.DataAtualizacao = atualizacao

	if cadastro := s.carregarCadastro(ctx, cpf); cadastro != nil {
		detalhe.Nome = naoVazio(cadastro.Nome)
		detalhe.NomeSocial = naoVazio(cadastro.NomeSocial)
		detalhe.Escolaridade = naoVazio(cadastro.Escolaridade)
		detalhe.Celular = naoVazio(cadastro.Celular)
		detalhe.Email = naoVazio(cadastro.Email)
		detalhe.Raca = naoVazio(cadastro.Raca)
		detalhe.Deficiencia = naoVazio(cadastro.Deficiencia)
		detalhe.Genero = generoAutodeclarado(cadastro.Genero)
		detalhe.Idade = idadeEm(cadastro.DataNascimento, time.Now())
		if cadastro.Endereco != nil {
			detalhe.Bairro = naoVazio(cadastro.Endereco.Bairro)
		}
	}

	return detalhe, nil
}

// carregarCadastro busca os dados pessoais do cidadão e os atualiza no RMI quando
// estão ausentes ou desatualizados. Falhas não impedem a ficha: sem cadastro, os
// campos pessoais saem como não informados.
func (s *BancoCurriculoService) carregarCadastro(ctx context.Context, cpf string) *models.CitizenSnapshot {
	cadastro, err := s.citizenSnapshotRepo.GetByCPF(ctx, cpf)
	if err != nil {
		log.Printf("[BancoCurriculoService] Failed to get citizen snapshot: %v", err)
		cadastro = nil
	}

	if s.citizenDataFetcher == nil {
		return cadastro
	}
	if cadastro != nil && time.Since(cadastro.LastSyncedAt) < s.citizenDataFetcher.StaleThreshold() {
		return cadastro
	}

	atualizado, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, cpf)
	if err != nil {
		log.Printf("[BancoCurriculoService] On-demand citizen sync failed: %v", err)
		return cadastro
	}
	if atualizado != nil {
		return atualizado
	}
	return cadastro
}

// profissaoDasExperiencias aplica a mesma regra da listagem (bancoCurriculosSelect,
// no repositório): o emprego atual; sem ele, o cargo com mais tempo de experiência.
func profissaoDasExperiencias(experiencias []*empregabilidade.CurriculoExperiencia) *string {
	var escolhida *empregabilidade.CurriculoExperiencia
	for _, e := range experiencias {
		if e != nil && (escolhida == nil || precedeComoProfissao(e, escolhida)) {
			escolhida = e
		}
	}
	if escolhida == nil {
		return nil
	}
	return naoVazio(escolhida.Cargo)
}

func precedeComoProfissao(a, b *empregabilidade.CurriculoExperiencia) bool {
	if a.EhTrabalhoAtual != b.EhTrabalhoAtual {
		return a.EhTrabalhoAtual
	}
	ma, mb := mesesDeExperiencia(a), mesesDeExperiencia(b)
	if ma != mb {
		return ma > mb
	}
	return a.Cargo < b.Cargo
}

// mesesDeExperiencia trata tempo não informado como menor que qualquer valor,
// como o NULLS LAST da listagem.
func mesesDeExperiencia(e *empregabilidade.CurriculoExperiencia) int {
	if e.TempoExperienciaMeses == nil {
		return -1
	}
	return *e.TempoExperienciaMeses
}

func generoAutodeclarado(valor string) *string {
	if opcao, ok := generosAutodeclarados[strings.ToLower(strings.TrimSpace(valor))]; ok {
		return &opcao
	}
	return nil
}

func idadeEm(nascimento *time.Time, hoje time.Time) *int {
	if nascimento == nil || nascimento.IsZero() {
		return nil
	}
	anos := hoje.Year() - nascimento.Year()
	if hoje.Month() < nascimento.Month() || (hoje.Month() == nascimento.Month() && hoje.Day() < nascimento.Day()) {
		anos--
	}
	if anos < 0 {
		return nil
	}
	return &anos
}

func naoVazio(valor string) *string {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return nil
	}
	return &valor
}
