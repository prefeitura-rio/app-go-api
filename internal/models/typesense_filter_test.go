package models

import (
	"errors"
	"strings"
	"testing"
)

// hostilFilterValues são valores que, interpolados crus no filter_by, mudam a
// expressão: "||" anula o action:=read fixo, "&&"/parênteses/colchetes/vírgula
// abrem filtros ou listas novos e a crase fecharia o literal.
var hostilFilterValues = []string{
	"x || action:=write",
	"x||action:!=read",
	"x && record.status:=rascunho",
	"(x) || (action:=delete)",
	"[a,b]",
	"a, b",
	"x` || action:=write || record.orgao_id:=`y",
}

func TestCursoSearchParameters_ToSearchParameters_HostileTextFilters_StayLiteral(t *testing.T) {
	fields := []struct {
		name  string
		field string
		set   func(p *CursoSearchParameters, v string)
	}{
		{"orgao_id", "record.orgao_id", func(p *CursoSearchParameters, v string) { p.OrgaoID = v }},
		{"status", "record.status", func(p *CursoSearchParameters, v string) { p.Status = v }},
		{"modalidade", "record.modalidade", func(p *CursoSearchParameters, v string) { p.Modalidade = v }},
		{"turno", "record.turno", func(p *CursoSearchParameters, v string) { p.Turno = v }},
		{"formato_aula", "record.formato_aula", func(p *CursoSearchParameters, v string) { p.FormatoAula = v }},
	}

	for _, f := range fields {
		for _, v := range hostilFilterValues {
			t.Run(f.name+"/"+v, func(t *testing.T) {
				p := &CursoSearchParameters{Q: "curso"}
				f.set(p, v)

				filterBy := p.ToSearchParameters().FilterBy

				assertSingleLiteralFilter(t, filterBy, f.field, v)
			})
		}
	}
}

func TestEmpregoSearchParameters_ToSearchParameters_HostileTextFilters_StayLiteral(t *testing.T) {
	fields := []struct {
		name  string
		field string
		set   func(p *EmpregoSearchParameters, v string)
	}{
		{"orgao_id", "record.orgao_id", func(p *EmpregoSearchParameters, v string) { p.OrgaoID = v }},
		{"status", "record.status", func(p *EmpregoSearchParameters, v string) { p.Status = v }},
		{"tipo_contratacao", "record.tipo_contratacao", func(p *EmpregoSearchParameters, v string) { p.TipoContratacao = v }},
		{"jornada_trabalho", "record.jornada_trabalho", func(p *EmpregoSearchParameters, v string) { p.JornadaTrabalho = v }},
		{"turno", "record.turno", func(p *EmpregoSearchParameters, v string) { p.Turno = v }},
	}

	for _, f := range fields {
		for _, v := range hostilFilterValues {
			t.Run(f.name+"/"+v, func(t *testing.T) {
				p := &EmpregoSearchParameters{Q: "vaga"}
				f.set(p, v)

				filterBy := p.ToSearchParameters().FilterBy

				assertSingleLiteralFilter(t, filterBy, f.field, v)
			})
		}
	}
}

// assertSingleLiteralFilter confere que o filter_by tem exatamente o
// action:=read fixo e um filtro do campo, com o valor inteiro dentro de um
// único literal entre crases (sem crase interna), de modo que nada do valor
// seja lido como operador pelo Typesense.
func assertSingleLiteralFilter(t *testing.T, filterBy, field, raw string) {
	t.Helper()

	prefix := "action:=read && " + field + ":=`"
	if !strings.HasPrefix(filterBy, prefix) {
		t.Fatalf("filter_by deveria começar com %q, veio %q", prefix, filterBy)
	}
	if !strings.HasSuffix(filterBy, "`") {
		t.Fatalf("filter_by deveria terminar fechando a crase, veio %q", filterBy)
	}

	literal := strings.TrimSuffix(strings.TrimPrefix(filterBy, prefix), "`")
	if strings.Contains(literal, "`") {
		t.Fatalf("literal não pode conter crase (fecharia o valor): %q", filterBy)
	}
	if want := strings.ReplaceAll(raw, "`", ""); literal != want {
		t.Fatalf("literal = %q, esperado %q", literal, want)
	}
	if strings.Count(filterBy, "`") != 2 {
		t.Fatalf("filter_by deveria ter exatamente um literal entre crases: %q", filterBy)
	}
}

func TestCursoSearchParameters_ToSearchParameters_NumericAndPlainFilters(t *testing.T) {
	p := &CursoSearchParameters{
		Q:             "curso",
		OrgaoID:       "SMAS",
		InstituicaoID: 7,
		Turno:         "MANHA",
	}

	got := p.ToSearchParameters().FilterBy

	want := "action:=read && record.orgao_id:=`SMAS` && record.instituicao_id:=7 && record.turno:=`MANHA`"
	if got != want {
		t.Fatalf("filter_by = %q, esperado %q", got, want)
	}
}

func TestEmpregoSearchParameters_ToSearchParameters_NumericAndPlainFilters(t *testing.T) {
	p := &EmpregoSearchParameters{
		Q:               "vaga",
		EmpresaID:       3,
		EscolaridadeID:  2,
		TipoContratacao: "CLT",
		SalarioMin:      1000,
		SalarioMax:      5000,
	}

	got := p.ToSearchParameters().FilterBy

	want := "action:=read && record.empresa_id:=3 && record.escolaridade_id:=2 && " +
		"record.tipo_contratacao:=`CLT` && record.salario_min:>=1000 && record.salario_max:<=5000"
	if got != want {
		t.Fatalf("filter_by = %q, esperado %q", got, want)
	}
}

func TestCursoSearchParameters_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CursoSearchParameters
		wantErr bool
	}{
		{"sem filtros", CursoSearchParameters{Q: "curso"}, false},
		{"filtros comuns", CursoSearchParameters{Q: "curso", OrgaoID: "SMAS", Status: "opened", Modalidade: "PRESENCIAL", Turno: "MANHA", FormatoAula: "teórico e prático"}, false},
		{"operador sem crase continua aceito (vira literal)", CursoSearchParameters{Q: "curso", OrgaoID: "x || action:=write"}, false},
		{"crase em orgao_id", CursoSearchParameters{Q: "curso", OrgaoID: "x` || action:=write"}, true},
		{"crase em status", CursoSearchParameters{Q: "curso", Status: "`"}, true},
		{"crase em modalidade", CursoSearchParameters{Q: "curso", Modalidade: "a`b"}, true},
		{"crase em turno", CursoSearchParameters{Q: "curso", Turno: "`MANHA`"}, true},
		{"crase em formato_aula", CursoSearchParameters{Q: "curso", FormatoAula: "x`"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidSearchFilter) {
					t.Fatalf("esperado ErrInvalidSearchFilter, veio %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}

func TestEmpregoSearchParameters_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  EmpregoSearchParameters
		wantErr bool
	}{
		{"sem filtros", EmpregoSearchParameters{Q: "vaga"}, false},
		{"filtros comuns", EmpregoSearchParameters{Q: "vaga", OrgaoID: "SMTE", Status: "publicado", TipoContratacao: "CLT", JornadaTrabalho: "40h semanais", Turno: "INTEGRAL"}, false},
		{"crase em orgao_id", EmpregoSearchParameters{Q: "vaga", OrgaoID: "`"}, true},
		{"crase em status", EmpregoSearchParameters{Q: "vaga", Status: "a`"}, true},
		{"crase em tipo_contratacao", EmpregoSearchParameters{Q: "vaga", TipoContratacao: "`CLT"}, true},
		{"crase em jornada_trabalho", EmpregoSearchParameters{Q: "vaga", JornadaTrabalho: "x`y"}, true},
		{"crase em turno", EmpregoSearchParameters{Q: "vaga", Turno: "`"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidSearchFilter) {
					t.Fatalf("esperado ErrInvalidSearchFilter, veio %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}
