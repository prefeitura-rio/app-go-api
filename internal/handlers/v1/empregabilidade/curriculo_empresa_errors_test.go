package empregabilidade_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// DEV-44: currículo e empresa não repassam o erro interno no corpo do 500
// ──────────────────────────────────────────────────────────────────────────────

// curriculoDriverErr imita o texto que o GORM/pq devolve numa falha de banco:
// tabela, constraint e o valor da chave em DETAIL. Nada disso pode chegar ao cliente.
const curriculoDriverErr = `ERROR: duplicate key value violates unique constraint "uni_emp_curriculo_formacoes_cpf" (SQLSTATE 23505) DETAIL: Key (cpf)=(00000000000) already exists. relation "emp_curriculo_formacoes"`

var curriculoLeakMarkers = []string{"SQLSTATE", "constraint", "DETAIL", "emp_curriculo", "00000000000", "erro ao persistir"}

func assertGeneric500(t *testing.T, w *httptest.ResponseRecorder, wantMessage string, leakMarkers []string) {
	t.Helper()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body %q: %v", w.Body.String(), err)
	}
	if body["error"] != wantMessage {
		t.Errorf("expected error %q, got %q", wantMessage, body["error"])
	}
	for _, leaked := range leakMarkers {
		if strings.Contains(w.Body.String(), leaked) {
			t.Errorf("response body leaks internal error detail %q: %s", leaked, w.Body.String())
		}
	}
}

func TestCurriculoHandler_ServiceError_DoesNotLeakInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const cpf = "12345678900"
	dbErr := fmt.Errorf("erro ao persistir: %w", fmt.Errorf("%s", curriculoDriverErr))
	id := uuid.New().String()

	// readErr: a busca inicial falha. writeErr: a busca acha o registro do próprio
	// usuário e a escrita (update/delete) falha.
	readErr := func() *mockCurriculoRepoH { return &mockCurriculoRepoH{err: dbErr} }
	writeErr := func() *mockCurriculoRepoH {
		return &mockCurriculoRepoH{
			writeErr:    dbErr,
			formacao:    &empmodels.CurriculoFormacao{CPF: cpf},
			idioma:      &empmodels.CurriculoIdioma{CPF: cpf},
			curso:       &empmodels.CurriculoCursoComplementar{CPF: cpf},
			experiencia: &empmodels.CurriculoExperiencia{CPF: cpf},
			conquista:   &empmodels.CurriculoConquista{CPF: cpf},
		}
	}

	tests := []struct {
		name        string
		repo        *mockCurriculoRepoH
		method      string
		path        string
		body        string
		wantMessage string
	}{
		{"GetCurriculoCompleto", readErr(), http.MethodGet, "/curriculo/" + cpf, "", "Erro ao buscar currículo"},

		{"CreateFormacao", readErr(), http.MethodPost, "/curriculo/formacoes", `{}`, "Erro ao criar formação"},
		{"GetFormacaoByID", readErr(), http.MethodGet, "/curriculo/formacoes/" + id, "", "Erro ao buscar formação"},
		{"UpdateFormacao lookup", readErr(), http.MethodPut, "/curriculo/formacoes/" + id, `{}`, "Erro ao buscar formação"},
		{"UpdateFormacao write", writeErr(), http.MethodPut, "/curriculo/formacoes/" + id, `{}`, "Erro ao atualizar formação"},
		{"DeleteFormacao lookup", readErr(), http.MethodDelete, "/curriculo/formacoes/" + id, "", "Erro ao buscar formação"},
		{"DeleteFormacao write", writeErr(), http.MethodDelete, "/curriculo/formacoes/" + id, "", "Erro ao excluir formação"},
		{"ListFormacoesByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/formacoes", "", "Erro ao listar formações"},

		{"CreateIdioma", readErr(), http.MethodPost, "/curriculo/idiomas", `{}`, "Erro ao criar idioma"},
		{"GetIdiomaByID", readErr(), http.MethodGet, "/curriculo/idiomas/" + id, "", "Erro ao buscar idioma"},
		{"UpdateIdioma lookup", readErr(), http.MethodPut, "/curriculo/idiomas/" + id, `{}`, "Erro ao buscar idioma"},
		{"UpdateIdioma write", writeErr(), http.MethodPut, "/curriculo/idiomas/" + id, `{}`, "Erro ao atualizar idioma"},
		{"DeleteIdioma lookup", readErr(), http.MethodDelete, "/curriculo/idiomas/" + id, "", "Erro ao buscar idioma"},
		{"DeleteIdioma write", writeErr(), http.MethodDelete, "/curriculo/idiomas/" + id, "", "Erro ao excluir idioma"},
		{"ListIdiomasByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/idiomas", "", "Erro ao listar idiomas"},

		{"CreateCursoComplementar", readErr(), http.MethodPost, "/curriculo/cursos", `{}`, "Erro ao criar curso complementar"},
		{"GetCursoComplementarByID", readErr(), http.MethodGet, "/curriculo/cursos/" + id, "", "Erro ao buscar curso complementar"},
		{"UpdateCursoComplementar lookup", readErr(), http.MethodPut, "/curriculo/cursos/" + id, `{}`, "Erro ao buscar curso complementar"},
		{"UpdateCursoComplementar write", writeErr(), http.MethodPut, "/curriculo/cursos/" + id, `{}`, "Erro ao atualizar curso complementar"},
		{"DeleteCursoComplementar lookup", readErr(), http.MethodDelete, "/curriculo/cursos/" + id, "", "Erro ao buscar curso complementar"},
		{"DeleteCursoComplementar write", writeErr(), http.MethodDelete, "/curriculo/cursos/" + id, "", "Erro ao excluir curso complementar"},
		{"ListCursosComplementaresByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/cursos", "", "Erro ao listar cursos complementares"},

		{"CreateExperiencia", readErr(), http.MethodPost, "/curriculo/experiencias", `{}`, "Erro ao criar experiência"},
		{"GetExperienciaByID", readErr(), http.MethodGet, "/curriculo/experiencias/" + id, "", "Erro ao buscar experiência"},
		{"UpdateExperiencia lookup", readErr(), http.MethodPut, "/curriculo/experiencias/" + id, `{}`, "Erro ao buscar experiência"},
		{"UpdateExperiencia write", writeErr(), http.MethodPut, "/curriculo/experiencias/" + id, `{}`, "Erro ao atualizar experiência"},
		{"DeleteExperiencia lookup", readErr(), http.MethodDelete, "/curriculo/experiencias/" + id, "", "Erro ao buscar experiência"},
		{"DeleteExperiencia write", writeErr(), http.MethodDelete, "/curriculo/experiencias/" + id, "", "Erro ao excluir experiência"},
		{"ListExperienciasByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/experiencias", "", "Erro ao listar experiências"},

		{"CreateConquista", readErr(), http.MethodPost, "/curriculo/conquistas", `{}`, "Erro ao criar conquista"},
		{"GetConquistaByID", readErr(), http.MethodGet, "/curriculo/conquistas/" + id, "", "Erro ao buscar conquista"},
		{"UpdateConquista lookup", readErr(), http.MethodPut, "/curriculo/conquistas/" + id, `{}`, "Erro ao buscar conquista"},
		{"UpdateConquista write", writeErr(), http.MethodPut, "/curriculo/conquistas/" + id, `{}`, "Erro ao atualizar conquista"},
		{"DeleteConquista lookup", readErr(), http.MethodDelete, "/curriculo/conquistas/" + id, "", "Erro ao buscar conquista"},
		{"DeleteConquista write", writeErr(), http.MethodDelete, "/curriculo/conquistas/" + id, "", "Erro ao excluir conquista"},
		{"ListConquistasByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/conquistas", "", "Erro ao listar conquistas"},

		{"ReplaceAllFormacoesByCPF", readErr(), http.MethodPut, "/curriculo/" + cpf + "/formacoes", `{}`, "Erro ao salvar formações e idiomas"},
		{"ReplaceAllExperienciasByCPF", readErr(), http.MethodPut, "/curriculo/" + cpf + "/experiencias", `{}`, "Erro ao salvar experiências e conquistas"},
		{"ReplaceAllConquistasByCPF", readErr(), http.MethodPut, "/curriculo/" + cpf + "/conquistas", `[]`, "Erro ao salvar conquistas"},
		{"ReplaceAllIdiomasByCPF", readErr(), http.MethodPut, "/curriculo/" + cpf + "/idiomas", `[]`, "Erro ao salvar idiomas"},
		{"ReplaceAllCursosComplementaresByCPF", readErr(), http.MethodPut, "/curriculo/" + cpf + "/cursos-complementares", `[]`, "Erro ao salvar cursos complementares"},

		{"UpsertSituacaoInteresses", readErr(), http.MethodPut, "/curriculo/" + cpf + "/situacao-interesses", `{}`, "Erro ao salvar situação e interesses"},
		{"GetSituacaoInteressesByCPF", readErr(), http.MethodGet, "/curriculo/" + cpf + "/situacao-interesses", "", "Erro ao buscar situação e interesses"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupCurriculoRouter(tt.repo, cpf, false)
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assertGeneric500(t, w, tt.wantMessage, curriculoLeakMarkers)
		})
	}
}

func TestEmpresaHandler_ConsultaCNPJ_ServiceError_DoesNotLeakInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Sem token manager, o serviço falha com "service account não configurado":
	// um detalhe de configuração interna que não pode chegar ao cliente.
	svc := services.NewEmpresaServiceWithInterface(&mockEmpresaRepoEmpH{})
	h := handlers.NewEmpresaHandler(svc).WithCNPJConsulta(&services.CNPJConsultaService{})

	r := gin.New()
	r.GET("/empresas/consulta-cnpj/:cnpj", h.ConsultaCNPJ)

	req := httptest.NewRequest(http.MethodGet, "/empresas/consulta-cnpj/11222333000181", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertGeneric500(t, w, "Erro ao consultar CNPJ", []string{"service account", "configurado", "token"})
}
