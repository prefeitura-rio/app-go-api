package empregabilidade

import "github.com/google/uuid"

// FaixaExperiencia agrupa o tempo total de experiência profissional, somado
// entre todas as experiências do currículo.
type FaixaExperiencia string

const (
	FaixaSemExperiencia FaixaExperiencia = "sem_experiencia"
	FaixaAte2Anos       FaixaExperiencia = "ate_2_anos"
	Faixa2a4Anos        FaixaExperiencia = "2_a_4_anos"
	Faixa4a6Anos        FaixaExperiencia = "4_a_6_anos"
	Faixa6a8Anos        FaixaExperiencia = "6_a_8_anos"
	Faixa8a10Anos       FaixaExperiencia = "8_a_10_anos"
	FaixaMaisDe10Anos   FaixaExperiencia = "mais_de_10_anos"
)

// LimitesMeses devolve o intervalo (mínimo exclusivo, máximo inclusivo) da
// faixa em meses; máximo nil é sem teto. "Até 2 anos" vai de 1 a 24 meses,
// "2 a 4 anos" de 25 a 48, e assim por diante.
func (f FaixaExperiencia) LimitesMeses() (minimoExclusivo int, maximo *int, ok bool) {
	teto := func(anos int) *int { m := anos * 12; return &m }
	switch f {
	case FaixaAte2Anos:
		return 0, teto(2), true
	case Faixa2a4Anos:
		return 24, teto(4), true
	case Faixa4a6Anos:
		return 48, teto(6), true
	case Faixa6a8Anos:
		return 72, teto(8), true
	case Faixa8a10Anos:
		return 96, teto(10), true
	case FaixaMaisDe10Anos:
		return 120, nil, true
	}
	return 0, nil, false
}

func (f FaixaExperiencia) IsValid() bool {
	if f == FaixaSemExperiencia {
		return true
	}
	_, _, ok := f.LimitesMeses()
	return ok
}

// ExigenciaCarteira filtra pela experiência comprovada em carteira.
type ExigenciaCarteira string

const (
	// CarteiraExigida: ao menos uma experiência comprovada em carteira.
	CarteiraExigida ExigenciaCarteira = "sim"
	// CarteiraSeisMeses: ao menos 6 meses somados de experiência em carteira.
	CarteiraSeisMeses ExigenciaCarteira = "6_meses"
)

func (e ExigenciaCarteira) IsValid() bool {
	return e == CarteiraExigida || e == CarteiraSeisMeses
}

// FaixaEtaria é o filtro de idade, com as mesmas opções do critério de idade
// da vaga (PREF-634): Maiores de 18 ou Jovem Aprendiz, sem idade livre, para
// não abrir espaço a etarismo. A idade vem da data de nascimento do cadastro.
type FaixaEtaria string

const (
	FaixaEtariaMaioresDe18   FaixaEtaria = "maiores_de_18"
	FaixaEtariaJovemAprendiz FaixaEtaria = "jovem_aprendiz"
)

// IdadeMinimaMaioresDe18 é a idade da opção Maiores de 18.
const IdadeMinimaMaioresDe18 = 18

// LimitesIdade devolve a idade mínima e a máxima da faixa, as duas inclusivas;
// máximo nil é sem teto.
func (f FaixaEtaria) LimitesIdade() (minimo int, maximo *int, ok bool) {
	switch f {
	case FaixaEtariaMaioresDe18:
		return IdadeMinimaMaioresDe18, nil, true
	case FaixaEtariaJovemAprendiz:
		teto := IdadeMaximaJovemAprendiz
		return IdadeMinimaJovemAprendiz, &teto, true
	}
	return 0, nil, false
}

func (f FaixaEtaria) IsValid() bool {
	_, _, ok := f.LimitesIdade()
	return ok
}

// FiltroCandidato são os filtros avançados de candidatos, os mesmos no banco
// de currículos e na lista de candidatos de uma vaga. Os dados pessoais vêm do
// cadastro do cidadão; experiência e idiomas, do currículo. Campo vazio não
// filtra, e quem não informou o dado filtrado fica de fora.
type FiltroCandidato struct {
	// Generos são identidades de gênero autodeclaradas, como o cidadão marca no app.
	Generos []string
	// FaixaEtaria deixa de fora quem não tem data de nascimento no cadastro.
	FaixaEtaria FaixaEtaria
	// TempoExperiencia vale pela soma das experiências do currículo.
	TempoExperiencia FaixaExperiencia
	Carteira         ExigenciaCarteira
	// Escolaridade é comparada exatamente com a do cadastro, sem diferenciar
	// maiúsculas.
	Escolaridade string
	IDIdioma     *uuid.UUID
	// IDNivelIdiomaMinimo, com IDIdioma, exige esse nível ou acima.
	IDNivelIdiomaMinimo *uuid.UUID
	Bairro              string
	// PCD true traz quem declarou deficiência; false, quem declarou não ter.
	PCD *bool
}

// Vazio diz se nenhum filtro foi pedido.
func (f FiltroCandidato) Vazio() bool {
	return len(f.Generos) == 0 && f.FaixaEtaria == "" &&
		f.TempoExperiencia == "" && f.Carteira == "" && f.Escolaridade == "" &&
		f.IDIdioma == nil && f.Bairro == "" && f.PCD == nil
}
