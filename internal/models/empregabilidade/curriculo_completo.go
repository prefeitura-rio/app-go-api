package empregabilidade

// CurriculoCompleto represents a complete curriculum/resume with all sections
type CurriculoCompleto struct {
	Formacoes             []*CurriculoFormacao              `json:"formacoes,omitempty"`
	Idiomas               []*CurriculoIdioma                `json:"idiomas,omitempty"`
	Habilidades           []*CurriculoHabilidade            `json:"habilidades,omitempty"`
	ComportamentoAtitudes []*CurriculoComportamentoAtitudes `json:"comportamento_atitudes,omitempty"`
	CursosComplementares  []*CurriculoCursoComplementar     `json:"cursos_complementares,omitempty"`
	Experiencias          []*CurriculoExperiencia           `json:"experiencias,omitempty"`
	Conquistas            []*CurriculoConquista             `json:"conquistas,omitempty"`
	SituacaoInteresses    *CurriculoSituacaoInteresses      `json:"situacao_interesses,omitempty"`
	ResumoProfissional    string                            `json:"resumo_profissional,omitempty"`
}

// CurriculoItensReplaceAll agrupa apenas os IDs para substituição massiva no currículo.
type CurriculoItensReplaceAll struct {
	HabilidadesIDs           []int64 `json:"habilidades_ids"`
	ComportamentoAtitudesIDs []int64 `json:"comportamento_atitudes_ids"`
}
