package empregabilidade

// Habilidades e competências que a secretaria cadastra na vaga (DEV-524,
// história PREF-660): uma ou mais áreas de atuação, os conhecimentos de cada
// área e os comportamentos/atitudes, gravados nas tabelas por vaga da DEV-528
// (emp_vagas_areas_atuacao, emp_vagas_conhecimentos e
// emp_vagas_comportamento_atitudes). Só as rotas de gestão da vaga carregam e
// devolvem esses dados; o cidadão não os vê.

// VagaAreaAtuacao é uma área de atuação da vaga com os conhecimentos marcados
// nela. No corpo do POST/PUT valem id_area_atuacao e conhecimento_ids; nome e
// conhecimentos só saem na resposta (e são ignorados na entrada, para que o
// que a leitura devolve possa voltar inteiro no POST de uma vaga duplicada).
type VagaAreaAtuacao struct {
	IDAreaAtuacao   int64          `json:"id_area_atuacao" example:"3"`
	Nome            string         `json:"nome" example:"Tecnologia da Informação"`
	ConhecimentoIDs []int64        `json:"conhecimento_ids"`
	Conhecimentos   []CatalogoItem `json:"conhecimentos"`
}

// VagaHabilidadesCompetencias é o que está gravado para a vaga, já com os
// nomes do catálogo.
type VagaHabilidadesCompetencias struct {
	AreasAtuacao           []VagaAreaAtuacao
	ComportamentosAtitudes []CatalogoItem
}

// SetHabilidadesCompetencias põe na vaga o que está gravado, sempre com
// listas (vazias quando não há registro, como numa vaga antiga), para que a
// leitura da gestão traga os três campos.
func (v *Vaga) SetHabilidadesCompetencias(h *VagaHabilidadesCompetencias) {
	areas := []VagaAreaAtuacao{}
	comportamentoIDs := []int64{}
	comportamentos := []CatalogoItem{}
	if h != nil {
		for _, a := range h.AreasAtuacao {
			area := VagaAreaAtuacao{
				IDAreaAtuacao:   a.IDAreaAtuacao,
				Nome:            a.Nome,
				ConhecimentoIDs: make([]int64, 0, len(a.Conhecimentos)),
				Conhecimentos:   make([]CatalogoItem, 0, len(a.Conhecimentos)),
			}
			for _, c := range a.Conhecimentos {
				area.ConhecimentoIDs = append(area.ConhecimentoIDs, c.ID)
				area.Conhecimentos = append(area.Conhecimentos, c)
			}
			areas = append(areas, area)
		}
		for _, c := range h.ComportamentosAtitudes {
			comportamentoIDs = append(comportamentoIDs, c.ID)
			comportamentos = append(comportamentos, c)
		}
	}
	v.AreasAtuacao = &areas
	v.ComportamentoAtitudesIDs = &comportamentoIDs
	v.ComportamentosAtitudes = &comportamentos
}
