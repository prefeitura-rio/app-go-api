package empregabilidade

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// lerFiltroCandidato lê os filtros avançados de candidatos da query string. São
// os mesmos parâmetros na listagem do banco de currículos e na lista de
// candidatos de uma vaga:
//
//	genero (repetível ou separado por vírgula), faixa_etaria (maiores_de_18|jovem_aprendiz),
//	tempo_experiencia, experiencia_carteira, escolaridade, id_idioma,
//	id_nivel_idioma, bairro, pcd (sim|nao)
func lerFiltroCandidato(c *gin.Context) (empregabilidade.FiltroCandidato, error) {
	var f empregabilidade.FiltroCandidato

	for _, valor := range c.QueryArray("genero") {
		for _, genero := range strings.Split(valor, ",") {
			if genero = strings.TrimSpace(genero); genero != "" {
				f.Generos = append(f.Generos, genero)
			}
		}
	}

	if raw := strings.TrimSpace(c.Query("faixa_etaria")); raw != "" {
		f.FaixaEtaria = empregabilidade.FaixaEtaria(raw)
		if !f.FaixaEtaria.IsValid() {
			return f, errors.New("faixa etária deve ser maiores_de_18 ou jovem_aprendiz")
		}
	}

	if raw := strings.TrimSpace(c.Query("tempo_experiencia")); raw != "" {
		f.TempoExperiencia = empregabilidade.FaixaExperiencia(raw)
		if !f.TempoExperiencia.IsValid() {
			return f, errors.New("tempo de experiência inválido")
		}
	}
	if raw := strings.TrimSpace(c.Query("experiencia_carteira")); raw != "" {
		f.Carteira = empregabilidade.ExigenciaCarteira(raw)
		if !f.Carteira.IsValid() {
			return f, errors.New("experiência em carteira inválida")
		}
	}

	var err error
	f.Escolaridade = strings.TrimSpace(c.Query("escolaridade"))
	f.Bairro = strings.TrimSpace(c.Query("bairro"))

	if f.IDIdioma, err = lerUUID(c.Query("id_idioma"), "idioma inválido"); err != nil {
		return f, err
	}
	if f.IDNivelIdiomaMinimo, err = lerUUID(c.Query("id_nivel_idioma"), "nível de idioma inválido"); err != nil {
		return f, err
	}
	if f.IDNivelIdiomaMinimo != nil && f.IDIdioma == nil {
		return f, errors.New("informe o idioma junto com o nível")
	}

	switch strings.TrimSpace(c.Query("pcd")) {
	case "":
	case "sim":
		sim := true
		f.PCD = &sim
	case "nao":
		nao := false
		f.PCD = &nao
	default:
		return f, errors.New("pcd deve ser sim ou nao")
	}

	return f, nil
}

func lerUUID(raw, mensagem string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, errors.New(mensagem)
	}
	return &id, nil
}
