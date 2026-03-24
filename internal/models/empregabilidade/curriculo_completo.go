package empregabilidade

// CurriculoCompleto represents a complete curriculum/resume with all sections
type CurriculoCompleto struct {
	Formacoes            []*CurriculoFormacao          `json:"formacoes"`
	Idiomas              []*CurriculoIdioma            `json:"idiomas"`
	CursosComplementares []*CurriculoCursoComplementar `json:"cursos_complementares"`
	Experiencias         []*CurriculoExperiencia       `json:"experiencias"`
	Conquistas           []*CurriculoConquista         `json:"conquistas"`
	SituacaoInteresses   *CurriculoSituacaoInteresses  `json:"situacao_interesses"`
}
