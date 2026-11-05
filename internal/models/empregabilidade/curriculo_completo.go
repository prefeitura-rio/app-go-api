package empregabilidade

// CurriculoCompleto represents a complete curriculum/resume with all sections
type CurriculoCompleto struct {
	Formacoes             []*CurriculoFormacao              `json:"formacoes"`
	Idiomas               []*CurriculoIdioma                `json:"idiomas"`
	AreaAtuacaoHabilidade []*CurriculoAreaAtuacaoHabilidade `json:"area_atuacao_habilidade"`
	ComportamentoAtitudes []*CurriculoComportamentoAtitudes `json:"comportamento_atitudes"`
	CursosComplementares  []*CurriculoCursoComplementar     `json:"cursos_complementares"`
	Experiencias          []*CurriculoExperiencia           `json:"experiencias"`
	Conquistas            []*CurriculoConquista             `json:"conquistas"`
	SituacaoInteresses    *CurriculoSituacaoInteresses      `json:"situacao_interesses"`
	ResumoProfissional    string                            `json:"resumo_profissional"`
}

// CurriculoItensReplaceAll agrupa apenas os IDs para substituição massiva no currículo.
type CurriculoItensReplaceAll struct {
	AreaAtuacaoHabilidadeIDs []int64 `json:"area_atuacao_habilidade_ids"`
	ComportamentoAtitudesIDs []int64 `json:"comportamento_atitudes_ids"`
}
