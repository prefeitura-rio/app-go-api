package empregabilidade

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
)

func ordem(v int) *int { return &v }

var (
	idFundamentalIncompleto = uuid.New()
	idMedioIncompleto       = uuid.New()
	idMedioCompleto         = uuid.New()
	idSuperiorCompleto      = uuid.New()
	idSemOrdem              = uuid.New()
	idIngles                = uuid.New()
	idEspanhol              = uuid.New()
	idBasico                = uuid.New()
	idIntermediario         = uuid.New()
	idAvancado              = uuid.New()
)

var refsTeste = referenciasElegibilidade{
	escolaridades: []*empregabilidade.Escolaridade{
		{ID: idFundamentalIncompleto, Descricao: "Fundamental incompleto", Ordem: ordem(1)},
		{ID: idMedioIncompleto, Descricao: "Médio incompleto", Ordem: ordem(3)},
		{ID: idMedioCompleto, Descricao: "Médio completo", Ordem: ordem(4)},
		{ID: idSuperiorCompleto, Descricao: "Superior completo", Ordem: ordem(7)},
		{ID: idSemOrdem, Descricao: "Sem ordem"},
	},
	idiomas: []*empregabilidade.Idioma{
		{ID: idIngles, Descricao: "Inglês"},
		{ID: idEspanhol, Descricao: "Espanhol"},
	},
	niveisIdioma: []*empregabilidade.NivelIdioma{
		{ID: idBasico, Descricao: "Básico", Ordem: ordem(1)},
		{ID: idIntermediario, Descricao: "Intermediário", Ordem: ordem(2)},
		{ID: idAvancado, Descricao: "Avançado", Ordem: ordem(3)},
	},
}

var hojeTeste = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

func formacao(id uuid.UUID, status empregabilidade.StatusFormacao) *empregabilidade.CurriculoFormacao {
	return &empregabilidade.CurriculoFormacao{IDEscolaridade: &id, Status: status}
}

func TestRequisitosNaoAtendidos_SemCriterios(t *testing.T) {
	faltas := requisitosNaoAtendidos(&empregabilidade.Vaga{}, candidatoElegibilidade{}, refsTeste, hojeTeste)
	assert.Empty(t, faltas)
}

func TestRequisitosNaoAtendidos_IdadeMinima(t *testing.T) {
	vaga := &empregabilidade.Vaga{IdadeMinima: ordem(18)}
	fez18Hoje := time.Date(2008, 10, 13, 0, 0, 0, 0, time.UTC)
	faz18Amanha := time.Date(2008, 10, 14, 0, 0, 0, 0, time.UTC)

	assert.Empty(t, requisitosNaoAtendidos(vaga, candidatoElegibilidade{dataNascimento: &fez18Hoje}, refsTeste, hojeTeste))
	assert.Equal(t, []string{"Ter pelo menos 18 anos"},
		requisitosNaoAtendidos(vaga, candidatoElegibilidade{dataNascimento: &faz18Amanha}, refsTeste, hojeTeste))
	assert.Equal(t, []string{"Ter pelo menos 18 anos"},
		requisitosNaoAtendidos(vaga, candidatoElegibilidade{}, refsTeste, hojeTeste), "sem data de nascimento não atende")
}

func TestRequisitosNaoAtendidos_Escolaridade(t *testing.T) {
	exigeMedioCompleto := &empregabilidade.Vaga{IDEscolaridadeMinima: &idMedioCompleto}
	exigeMedioIncompleto := &empregabilidade.Vaga{IDEscolaridadeMinima: &idMedioIncompleto}
	falta := []string{"Formação mínima — Médio completo"}

	cases := []struct {
		name      string
		vaga      *empregabilidade.Vaga
		candidato candidatoElegibilidade
		faltas    []string
	}{
		{
			name: "formação concluída de ordem maior atende",
			vaga: exigeMedioCompleto,
			candidato: candidatoElegibilidade{curriculo: &empregabilidade.CurriculoCompleto{
				Formacoes: []*empregabilidade.CurriculoFormacao{formacao(idSuperiorCompleto, empregabilidade.StatusFormacaoCompleto)},
			}},
		},
		{
			name: "formação de ordem maior em andamento não atende",
			vaga: exigeMedioCompleto,
			candidato: candidatoElegibilidade{curriculo: &empregabilidade.CurriculoCompleto{
				Formacoes: []*empregabilidade.CurriculoFormacao{formacao(idSuperiorCompleto, empregabilidade.StatusFormacaoEmAndamento)},
			}},
			faltas: falta,
		},
		{
			name: "exigência incompleto aceita a mesma escolaridade com qualquer status",
			vaga: exigeMedioIncompleto,
			candidato: candidatoElegibilidade{curriculo: &empregabilidade.CurriculoCompleto{
				Formacoes: []*empregabilidade.CurriculoFormacao{formacao(idMedioIncompleto, empregabilidade.StatusFormacaoEmAndamento)},
			}},
		},
		{
			name: "formação sem status é ignorada",
			vaga: exigeMedioIncompleto,
			candidato: candidatoElegibilidade{curriculo: &empregabilidade.CurriculoCompleto{
				Formacoes: []*empregabilidade.CurriculoFormacao{formacao(idMedioIncompleto, "")},
			}},
			faltas: []string{"Formação mínima — Médio incompleto"},
		},
		{
			name:      "sem formação, vale a escolaridade do cadastro, sem diferenciar maiúsculas",
			vaga:      exigeMedioCompleto,
			candidato: candidatoElegibilidade{escolaridadeCadastro: " superior COMPLETO "},
		},
		{
			name:      "escolaridade do cadastro abaixo da exigida",
			vaga:      exigeMedioCompleto,
			candidato: candidatoElegibilidade{escolaridadeCadastro: "Fundamental incompleto"},
			faltas:    falta,
		},
		{
			name:      "escolaridade do cadastro desconhecida",
			vaga:      exigeMedioCompleto,
			candidato: candidatoElegibilidade{escolaridadeCadastro: "Ensino médio"},
			faltas:    falta,
		},
		{
			name:      "exigência sem ordem nunca é atendida",
			vaga:      &empregabilidade.Vaga{IDEscolaridadeMinima: &idSemOrdem},
			candidato: candidatoElegibilidade{escolaridadeCadastro: "Superior completo"},
			faltas:    []string{"Formação mínima — Sem ordem"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.faltas, requisitosNaoAtendidos(tc.vaga, tc.candidato, refsTeste, hojeTeste))
		})
	}
}

func TestRequisitosNaoAtendidos_Idiomas(t *testing.T) {
	vaga := &empregabilidade.Vaga{IdiomasRequisito: []empregabilidade.VagaIdiomaRequisito{
		{IDIdioma: idIngles, IDNivelMinimo: idIntermediario},
		{IDIdioma: idEspanhol, IDNivelMinimo: idBasico},
	}}
	curriculo := &empregabilidade.CurriculoCompleto{Idiomas: []*empregabilidade.CurriculoIdioma{
		{IDIdioma: idIngles, IDNivel: idAvancado},
	}}

	faltas := requisitosNaoAtendidos(vaga, candidatoElegibilidade{curriculo: curriculo}, refsTeste, hojeTeste)
	assert.Equal(t, []string{"Idioma Espanhol — nível Básico ou superior"}, faltas)

	curriculo.Idiomas = []*empregabilidade.CurriculoIdioma{
		{IDIdioma: idIngles, IDNivel: idBasico},
		{IDIdioma: idEspanhol, IDNivel: idBasico},
	}
	faltas = requisitosNaoAtendidos(vaga, candidatoElegibilidade{curriculo: curriculo}, refsTeste, hojeTeste)
	assert.Equal(t, []string{"Idioma Inglês — nível Intermediário ou superior"}, faltas)
}

func TestRequisitosNaoAtendidos_JuntaTodasAsFaltas(t *testing.T) {
	vaga := &empregabilidade.Vaga{
		IdadeMinima:          ordem(18),
		IDEscolaridadeMinima: &idMedioCompleto,
		IdiomasRequisito:     []empregabilidade.VagaIdiomaRequisito{{IDIdioma: idIngles, IDNivelMinimo: idBasico}},
	}

	faltas := requisitosNaoAtendidos(vaga, candidatoElegibilidade{}, refsTeste, hojeTeste)

	assert.Equal(t, []string{
		"Ter pelo menos 18 anos",
		"Formação mínima — Médio completo",
		"Idioma Inglês — nível Básico ou superior",
	}, faltas)
}
