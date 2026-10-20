package empregabilidade

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// referenciasElegibilidade são as tabelas de apoio que a régua de elegibilidade
// consulta: a ordem das escolaridades e dos níveis de idioma, e os nomes que
// entram nas mensagens.
type referenciasElegibilidade struct {
	escolaridades []*empregabilidade.Escolaridade
	idiomas       []*empregabilidade.Idioma
	niveisIdioma  []*empregabilidade.NivelIdioma
}

// candidatoElegibilidade é o que a régua precisa saber do candidato.
type candidatoElegibilidade struct {
	dataNascimento *time.Time
	// escolaridadeCadastro é a escolaridade autodeclarada no cadastro (RMI),
	// usada quando nenhuma formação do currículo atende.
	escolaridadeCadastro string
	curriculo            *empregabilidade.CurriculoCompleto
}

// requisitosNaoAtendidos aplica os critérios de elegibilidade da vaga — idade
// mínima, escolaridade mínima e idiomas — e devolve o que o candidato não
// atende, com o mesmo texto que o app mostra. É a mesma régua do app do
// cidadão (superapp, src/lib/eligibility-utils.ts), repetida aqui porque a
// associação pelo backoffice não passa pelo app.
func requisitosNaoAtendidos(vaga *empregabilidade.Vaga, candidato candidatoElegibilidade, refs referenciasElegibilidade, agora time.Time) []string {
	var faltas []string

	if vaga.IdadeMinima != nil {
		idade := idadeEm(candidato.dataNascimento, agora)
		if idade == nil || *idade < *vaga.IdadeMinima {
			faltas = append(faltas, fmt.Sprintf("Ter pelo menos %d anos", *vaga.IdadeMinima))
		}
	}

	if vaga.IDEscolaridadeMinima != nil && !atendeEscolaridade(*vaga.IDEscolaridadeMinima, candidato, refs) {
		nome := "Escolaridade mínima exigida"
		if e := escolaridadePorID(refs.escolaridades, *vaga.IDEscolaridadeMinima); e != nil {
			nome = e.Descricao
		}
		faltas = append(faltas, "Formação mínima — "+nome)
	}

	for _, requisito := range vaga.IdiomasRequisito {
		if !atendeIdioma(requisito, candidato.curriculo, refs) {
			faltas = append(faltas, fmt.Sprintf("Idioma %s — nível %s ou superior",
				nomeIdioma(refs.idiomas, requisito.IDIdioma), nomeNivel(refs.niveisIdioma, requisito.IDNivelMinimo)))
		}
	}

	return faltas
}

// atendeEscolaridade vale pela formação do currículo ou, sem nenhuma que
// atenda, pela escolaridade do cadastro. Exigência "incompleto" aceita a mesma
// escolaridade com qualquer status; as demais pedem formação concluída de
// ordem igual ou maior.
func atendeEscolaridade(idMinima uuid.UUID, candidato candidatoElegibilidade, refs referenciasElegibilidade) bool {
	minima := escolaridadePorID(refs.escolaridades, idMinima)
	ordemMinima := ordemEscolaridade(minima)
	if ordemMinima < 0 {
		return false
	}
	exigeIncompleto := strings.Contains(strings.ToLower(minima.Descricao), "incompleto")

	if candidato.curriculo != nil {
		for _, formacao := range candidato.curriculo.Formacoes {
			if formacao == nil || formacao.Status == "" || formacao.IDEscolaridade == nil {
				continue
			}
			if exigeIncompleto && *formacao.IDEscolaridade == idMinima {
				return true
			}
			if formacao.Status == empregabilidade.StatusFormacaoCompleto &&
				ordemEscolaridade(escolaridadePorID(refs.escolaridades, *formacao.IDEscolaridade)) >= ordemMinima {
				return true
			}
		}
	}

	cadastro := escolaridadePorDescricao(refs.escolaridades, candidato.escolaridadeCadastro)
	return cadastro != nil && ordemEscolaridade(cadastro) >= ordemMinima
}

func atendeIdioma(requisito empregabilidade.VagaIdiomaRequisito, curriculo *empregabilidade.CurriculoCompleto, refs referenciasElegibilidade) bool {
	if curriculo == nil {
		return false
	}
	ordemMinima := ordemNivel(refs.niveisIdioma, requisito.IDNivelMinimo)
	for _, idioma := range curriculo.Idiomas {
		if idioma != nil && idioma.IDIdioma == requisito.IDIdioma &&
			ordemNivel(refs.niveisIdioma, idioma.IDNivel) >= ordemMinima {
			return true
		}
	}
	return false
}

func escolaridadePorID(escolaridades []*empregabilidade.Escolaridade, id uuid.UUID) *empregabilidade.Escolaridade {
	for _, e := range escolaridades {
		if e != nil && e.ID == id {
			return e
		}
	}
	return nil
}

func escolaridadePorDescricao(escolaridades []*empregabilidade.Escolaridade, descricao string) *empregabilidade.Escolaridade {
	alvo := strings.ToLower(strings.TrimSpace(descricao))
	if alvo == "" {
		return nil
	}
	for _, e := range escolaridades {
		if e != nil && strings.ToLower(strings.TrimSpace(e.Descricao)) == alvo {
			return e
		}
	}
	return nil
}

// ordemEscolaridade devolve -1 quando a escolaridade não existe ou não tem
// ordem, o que nunca atende uma exigência.
func ordemEscolaridade(e *empregabilidade.Escolaridade) int {
	if e == nil || e.Ordem == nil {
		return -1
	}
	return *e.Ordem
}

func ordemNivel(niveis []*empregabilidade.NivelIdioma, id uuid.UUID) int {
	for _, n := range niveis {
		if n != nil && n.ID == id && n.Ordem != nil {
			return *n.Ordem
		}
	}
	return -1
}

func nomeIdioma(idiomas []*empregabilidade.Idioma, id uuid.UUID) string {
	for _, i := range idiomas {
		if i != nil && i.ID == id {
			return i.Descricao
		}
	}
	return "exigido"
}

func nomeNivel(niveis []*empregabilidade.NivelIdioma, id uuid.UUID) string {
	for _, n := range niveis {
		if n != nil && n.ID == id {
			return n.Descricao
		}
	}
	return "mínimo exigido"
}
