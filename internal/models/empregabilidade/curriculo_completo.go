package empregabilidade

// CurriculoCompleto represents a complete curriculum/resume with all sections
type CurriculoCompleto struct {
	Formacoes              []*CurriculoFormacao          `json:"formacoes"`
	Idiomas                []*CurriculoIdioma            `json:"idiomas"`
	AreaAtuacao            *CatalogoItem                 `json:"area_atuacao" extensions:"x-nullable"`
	Conhecimentos          []*CatalogoItem               `json:"conhecimentos"`
	ComportamentosAtitudes []*CatalogoItem               `json:"comportamentos_atitudes"`
	CursosComplementares   []*CurriculoCursoComplementar `json:"cursos_complementares"`
	Experiencias           []*CurriculoExperiencia       `json:"experiencias"`
	Conquistas             []*CurriculoConquista         `json:"conquistas"`
	SituacaoInteresses     *CurriculoSituacaoInteresses  `json:"situacao_interesses"`
	ResumoProfissional     string                        `json:"resumo_profissional"`
}

// HabilidadesCompetencias é o estado da seção no currículo do cidadão.
type HabilidadesCompetencias struct {
	AreaAtuacao            *CatalogoItem   `json:"area_atuacao" extensions:"x-nullable"`
	Conhecimentos          []*CatalogoItem `json:"conhecimentos"`
	ComportamentosAtitudes []*CatalogoItem `json:"comportamentos_atitudes"`
}

// HabilidadesCompetenciasSave é o formulário inteiro enviado pelo Salvar.
type HabilidadesCompetenciasSave struct {
	IDAreaAtuacao            *int64  `json:"id_area_atuacao" extensions:"x-nullable"`
	HabilidadeIDs            []int64 `json:"habilidade_ids"`
	ComportamentoAtitudesIDs []int64 `json:"comportamento_atitudes_ids"`
}
