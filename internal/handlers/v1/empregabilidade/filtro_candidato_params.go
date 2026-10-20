package empregabilidade

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// lerFiltroCandidato lê os filtros avançados de candidatos da query string. São
// os mesmos parâmetros na listagem do banco de currículos e na lista de
// candidatos de uma vaga:
//
//	genero (repetível ou separado por vírgula), idade_min, idade_max,
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

	var err error
	if f.IdadeMinima, err = lerIdade(c.Query("idade_min")); err != nil {
		return f, err
	}
	if f.IdadeMaxima, err = lerIdade(c.Query("idade_max")); err != nil {
		return f, err
	}
	if f.IdadeMinima != nil && f.IdadeMaxima != nil && *f.IdadeMinima > *f.IdadeMaxima {
		return f, errors.New("idade mínima maior que a máxima")
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

func lerIdade(raw string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	idade, err := strconv.Atoi(raw)
	if err != nil || idade < 0 || idade > 130 {
		return nil, errors.New("idade inválida")
	}
	return &idade, nil
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
