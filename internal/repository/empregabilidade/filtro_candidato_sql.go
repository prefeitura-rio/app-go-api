package empregabilidade

import (
	"strings"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// cpfDigitosCandidatura é o CPF da candidatura só com dígitos: a candidatura
// criada pelo backoffice pode ter gravado o CPF com máscara.
const cpfDigitosCandidatura = `regexp_replace(emp_candidaturas.cpf, '[^0-9]', '', 'g')`

// semDeficiencia são as respostas de quem declarou não ter deficiência: a
// opção do app e valores antigos do RMI. Comparadas em minúsculas.
var semDeficiencia = []string{
	"não sou pessoa com deficiência",
	"nenhuma",
	"não possui",
	"não",
	"nao",
}

// condicaoFiltroCandidato monta a condição SQL dos filtros avançados para o
// CPF dado por cpfExpr (uma expressão SQL que vale o CPF só com dígitos). Cada
// filtro vira uma subconsulta pelo CPF, sem JOIN, para não deixar colunas
// ambíguas na consulta de quem chama. Devolve "" quando não há filtro.
func condicaoFiltroCandidato(f empregabilidade.FiltroCandidato, cpfExpr string) (string, []interface{}) {
	var cadastro, curriculo []string
	var argsCadastro, argsCurriculo []interface{}

	if len(f.Generos) > 0 {
		generos := make([]string, 0, len(f.Generos))
		for _, g := range f.Generos {
			generos = append(generos, strings.ToLower(strings.TrimSpace(g)))
		}
		cadastro = append(cadastro, "lower(trim(fcs.genero)) IN ?")
		argsCadastro = append(argsCadastro, generos)
	}
	if f.IdadeMinima != nil {
		cadastro = append(cadastro, "date_part('year', age(CURRENT_DATE, fcs.data_nascimento)) >= ?")
		argsCadastro = append(argsCadastro, *f.IdadeMinima)
	}
	if f.IdadeMaxima != nil {
		cadastro = append(cadastro, "date_part('year', age(CURRENT_DATE, fcs.data_nascimento)) <= ?")
		argsCadastro = append(argsCadastro, *f.IdadeMaxima)
	}
	if f.Escolaridade != "" {
		cadastro = append(cadastro, "lower(trim(fcs.escolaridade)) = lower(trim(?))")
		argsCadastro = append(argsCadastro, f.Escolaridade)
	}
	if f.Bairro != "" {
		cadastro = append(cadastro, "unaccent(lower(trim(fcs.endereco->>'bairro'))) = unaccent(lower(trim(?)))")
		argsCadastro = append(argsCadastro, f.Bairro)
	}
	if f.PCD != nil {
		if *f.PCD {
			cadastro = append(cadastro, "NULLIF(trim(fcs.deficiencia), '') IS NOT NULL AND lower(trim(fcs.deficiencia)) NOT IN ?")
		} else {
			cadastro = append(cadastro, "lower(trim(fcs.deficiencia)) IN ?")
		}
		argsCadastro = append(argsCadastro, semDeficiencia)
	}

	somaExperiencia := "COALESCE((SELECT SUM(fe.tempo_experiencia_meses) FROM emp_curriculo_experiencias fe WHERE fe.cpf = " + cpfExpr + "), 0)"
	switch f.TempoExperiencia {
	case "":
	case empregabilidade.FaixaSemExperiencia:
		curriculo = append(curriculo, "NOT EXISTS (SELECT 1 FROM emp_curriculo_experiencias fe WHERE fe.cpf = "+cpfExpr+")")
	default:
		if minimo, maximo, ok := f.TempoExperiencia.LimitesMeses(); ok {
			curriculo = append(curriculo, somaExperiencia+" > ?")
			argsCurriculo = append(argsCurriculo, minimo)
			if maximo != nil {
				curriculo = append(curriculo, somaExperiencia+" <= ?")
				argsCurriculo = append(argsCurriculo, *maximo)
			}
		}
	}

	switch f.Carteira {
	case empregabilidade.CarteiraExigida:
		curriculo = append(curriculo, "EXISTS (SELECT 1 FROM emp_curriculo_experiencias fe WHERE fe.cpf = "+cpfExpr+" AND fe.experiencia_comprovada_ct)")
	case empregabilidade.CarteiraSeisMeses:
		curriculo = append(curriculo, "COALESCE((SELECT SUM(fe.tempo_experiencia_meses) FROM emp_curriculo_experiencias fe WHERE fe.cpf = "+cpfExpr+" AND fe.experiencia_comprovada_ct), 0) >= 6")
	}

	if f.IDIdioma != nil {
		idioma := "EXISTS (SELECT 1 FROM emp_curriculo_idiomas fi JOIN emp_niveis_idioma fn ON fn.id = fi.id_nivel WHERE fi.cpf = " + cpfExpr + " AND fi.id_idioma = ?"
		argsCurriculo = append(argsCurriculo, *f.IDIdioma)
		if f.IDNivelIdiomaMinimo != nil {
			idioma += " AND fn.ordem >= (SELECT ordem FROM emp_niveis_idioma WHERE id = ?)"
			argsCurriculo = append(argsCurriculo, *f.IDNivelIdiomaMinimo)
		}
		curriculo = append(curriculo, idioma+")")
	}

	var partes []string
	var args []interface{}
	if len(cadastro) > 0 {
		partes = append(partes, "EXISTS (SELECT 1 FROM citizen_snapshots fcs WHERE fcs.cpf = "+cpfExpr+" AND "+strings.Join(cadastro, " AND ")+")")
		args = append(args, argsCadastro...)
	}
	if len(curriculo) > 0 {
		partes = append(partes, curriculo...)
		args = append(args, argsCurriculo...)
	}
	if len(partes) == 0 {
		return "", nil
	}
	return "(" + strings.Join(partes, " AND ") + ")", args
}
