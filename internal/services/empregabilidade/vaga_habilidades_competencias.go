package empregabilidade

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// Campos das habilidades e competências da vaga (DEV-524), como aparecem no
// JSON e nas chaves de "fields" dos erros de validação.
const (
	CampoAreasAtuacao             = "areas_atuacao"
	CampoConhecimentosDaArea      = "areas_atuacao.conhecimento_ids"
	CampoComportamentoAtitudesIDs = "comportamento_atitudes_ids"
)

// Tamanho máximo das listas aceitas no corpo da vaga. Folgados para o
// formulário, eles mantêm a validação (um IN por lista) bem abaixo do limite
// de parâmetros do Postgres.
const (
	MaxAreasAtuacaoPorVaga   = 50
	MaxConhecimentosPorArea  = 300
	MaxComportamentosPorVaga = 100
)

// Mensagens dos campos exigidos para publicar ou enviar para aprovação.
const (
	MsgAreaAtuacaoObrigatoria   = "Informe ao menos uma área de atuação."
	MsgComportamentoObrigatorio = "Informe ao menos um comportamento ou atitude."
)

func erroCampo(campo, mensagem string) error {
	return apperrors.ValidationFields(mensagem, map[string]string{campo: mensagem})
}

func juntarIDs(ids []int64) string {
	partes := make([]string, len(ids))
	for i, id := range ids {
		partes[i] = fmt.Sprint(id)
	}
	return strings.Join(partes, ", ")
}

// validarHabilidadesCompetencias confere as listas enviadas no corpo do
// POST/PUT e as deixa no formato gravado: tira conhecimento repetido na mesma
// área e comportamento repetido. Recusa id inválido, área repetida na vaga,
// área ou comportamento fora do catálogo e conhecimento que não pertence à
// área em que foi marcado. Lista ausente (nil) não é exigida: o rascunho
// salva sem elas e o PUT mantém o gravado.
func (s *VagaService) validarHabilidadesCompetencias(ctx context.Context, entity *empregabilidade.Vaga) error {
	if entity.AreasAtuacao != nil {
		areas, err := s.normalizarAreasAtuacao(ctx, *entity.AreasAtuacao)
		if err != nil {
			return err
		}
		entity.AreasAtuacao = &areas
	}
	if entity.ComportamentoAtitudesIDs != nil {
		comportamentos, err := s.normalizarComportamentos(ctx, *entity.ComportamentoAtitudesIDs)
		if err != nil {
			return err
		}
		entity.ComportamentoAtitudesIDs = &comportamentos
	}
	// Nome e conhecimentos com nome só saem na resposta.
	entity.ComportamentosAtitudes = nil
	return nil
}

func (s *VagaService) normalizarAreasAtuacao(ctx context.Context, entrada []empregabilidade.VagaAreaAtuacao) ([]empregabilidade.VagaAreaAtuacao, error) {
	if len(entrada) > MaxAreasAtuacaoPorVaga {
		return nil, erroCampo(CampoAreasAtuacao, fmt.Sprintf("Informe no máximo %d áreas de atuação por vaga.", MaxAreasAtuacaoPorVaga))
	}
	areas := make([]empregabilidade.VagaAreaAtuacao, 0, len(entrada))
	vistas := make(map[int64]bool, len(entrada))
	areaIDs := make([]int64, 0, len(entrada))
	var todosConhecimentos []int64
	for _, a := range entrada {
		if a.IDAreaAtuacao <= 0 {
			return nil, erroCampo(CampoAreasAtuacao, fmt.Sprintf("id_area_atuacao inválido: %d.", a.IDAreaAtuacao))
		}
		if vistas[a.IDAreaAtuacao] {
			return nil, erroCampo(CampoAreasAtuacao, fmt.Sprintf("Área de atuação repetida na vaga (id_area_atuacao %d): informe cada área uma vez, com todos os conhecimentos dela.", a.IDAreaAtuacao))
		}
		vistas[a.IDAreaAtuacao] = true
		areaIDs = append(areaIDs, a.IDAreaAtuacao)
		if len(a.ConhecimentoIDs) > MaxConhecimentosPorArea {
			return nil, erroCampo(CampoConhecimentosDaArea, fmt.Sprintf("Informe no máximo %d conhecimentos por área de atuação (área %d).", MaxConhecimentosPorArea, a.IDAreaAtuacao))
		}

		conhecimentos := make([]int64, 0, len(a.ConhecimentoIDs))
		repetido := make(map[int64]bool, len(a.ConhecimentoIDs))
		for _, c := range a.ConhecimentoIDs {
			if c <= 0 {
				return nil, erroCampo(CampoConhecimentosDaArea, fmt.Sprintf("Conhecimento inválido na área de atuação %d: %d.", a.IDAreaAtuacao, c))
			}
			if repetido[c] {
				continue
			}
			repetido[c] = true
			conhecimentos = append(conhecimentos, c)
			todosConhecimentos = append(todosConhecimentos, c)
		}
		areas = append(areas, empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: a.IDAreaAtuacao, ConhecimentoIDs: conhecimentos})
	}
	if len(areas) == 0 {
		return areas, nil
	}

	existentes, err := s.repo.AreasAtuacaoExistentes(ctx, areaIDs)
	if err != nil {
		return nil, err
	}
	existe := make(map[int64]bool, len(existentes))
	for _, id := range existentes {
		existe[id] = true
	}
	var inexistentes []int64
	for _, id := range areaIDs {
		if !existe[id] {
			inexistentes = append(inexistentes, id)
		}
	}
	if len(inexistentes) > 0 {
		return nil, erroCampo(CampoAreasAtuacao, fmt.Sprintf("Área de atuação não encontrada: %s.", juntarIDs(inexistentes)))
	}

	if len(todosConhecimentos) == 0 {
		return areas, nil
	}
	pares, err := s.repo.ConhecimentosDasAreas(ctx, areaIDs, todosConhecimentos)
	if err != nil {
		return nil, err
	}
	var problemas []string
	for _, a := range areas {
		var fora []int64
		for _, c := range a.ConhecimentoIDs {
			if !pares[a.IDAreaAtuacao][c] {
				fora = append(fora, c)
			}
		}
		if len(fora) > 0 {
			problemas = append(problemas, fmt.Sprintf("conhecimento %s não pertence à área de atuação %d", juntarIDs(fora), a.IDAreaAtuacao))
		}
	}
	if len(problemas) > 0 {
		msg := "Cada conhecimento precisa ser da área em que foi marcado: " + strings.Join(problemas, "; ") + "."
		return nil, erroCampo(CampoConhecimentosDaArea, msg)
	}
	return areas, nil
}

func (s *VagaService) normalizarComportamentos(ctx context.Context, entrada []int64) ([]int64, error) {
	if len(entrada) > MaxComportamentosPorVaga {
		return nil, erroCampo(CampoComportamentoAtitudesIDs, fmt.Sprintf("Informe no máximo %d comportamentos ou atitudes por vaga.", MaxComportamentosPorVaga))
	}
	out := make([]int64, 0, len(entrada))
	visto := make(map[int64]bool, len(entrada))
	for _, id := range entrada {
		if id <= 0 {
			return nil, erroCampo(CampoComportamentoAtitudesIDs, fmt.Sprintf("Comportamento ou atitude inválido: %d.", id))
		}
		if visto[id] {
			continue
		}
		visto[id] = true
		out = append(out, id)
	}
	if len(out) == 0 {
		return out, nil
	}
	existentes, err := s.repo.ComportamentosExistentes(ctx, out)
	if err != nil {
		return nil, err
	}
	existe := make(map[int64]bool, len(existentes))
	for _, id := range existentes {
		existe[id] = true
	}
	var inexistentes []int64
	for _, id := range out {
		if !existe[id] {
			inexistentes = append(inexistentes, id)
		}
	}
	if len(inexistentes) > 0 {
		return nil, erroCampo(CampoComportamentoAtitudesIDs, fmt.Sprintf("Comportamento ou atitude não encontrado: %s.", juntarIDs(inexistentes)))
	}
	return out, nil
}

// exigirHabilidadesCompetencias recusa publicar ou enviar para aprovação a
// vaga sem área de atuação, com área sem conhecimento ou sem comportamento,
// com o erro de cada campo que falta.
func (s *VagaService) exigirHabilidadesCompetencias(ctx context.Context, vagaID uuid.UUID, acao string) error {
	h, err := s.repo.GetHabilidadesCompetencias(ctx, vagaID)
	if err != nil {
		return err
	}
	if h == nil {
		h = &empregabilidade.VagaHabilidadesCompetencias{}
	}
	fields := map[string]string{}
	if len(h.AreasAtuacao) == 0 {
		fields[CampoAreasAtuacao] = MsgAreaAtuacaoObrigatoria
	}
	var semConhecimento []string
	for _, a := range h.AreasAtuacao {
		if len(a.Conhecimentos) == 0 {
			nome := a.Nome
			if nome == "" {
				nome = fmt.Sprint(a.IDAreaAtuacao)
			}
			semConhecimento = append(semConhecimento, nome)
		}
	}
	if len(semConhecimento) > 0 {
		fields[CampoConhecimentosDaArea] = "Informe ao menos um conhecimento em cada área de atuação (falta em: " + strings.Join(semConhecimento, ", ") + ")."
	}
	if len(h.ComportamentosAtitudes) == 0 {
		fields[CampoComportamentoAtitudesIDs] = MsgComportamentoObrigatorio
	}
	if len(fields) > 0 {
		return apperrors.ValidationFields("Preencha as habilidades e competências da vaga para "+acao+".", fields)
	}
	return nil
}

// habilidadesDoCorpo monta, sem os nomes, as listas que vieram no corpo.
func habilidadesDoCorpo(entity *empregabilidade.Vaga) *empregabilidade.VagaHabilidadesCompetencias {
	h := &empregabilidade.VagaHabilidadesCompetencias{}
	if entity.AreasAtuacao != nil {
		for _, a := range *entity.AreasAtuacao {
			area := empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: a.IDAreaAtuacao}
			for _, c := range a.ConhecimentoIDs {
				area.Conhecimentos = append(area.Conhecimentos, empregabilidade.CatalogoItem{ID: c})
			}
			h.AreasAtuacao = append(h.AreasAtuacao, area)
		}
	}
	if entity.ComportamentoAtitudesIDs != nil {
		for _, c := range *entity.ComportamentoAtitudesIDs {
			h.ComportamentosAtitudes = append(h.ComportamentosAtitudes, empregabilidade.CatalogoItem{ID: c})
		}
	}
	return h
}
