package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empModels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
)

// etapaGetRepoStub atende só o GetByID e o ListByVaga; os demais métodos da
// interface não são usados pelas rotas GET de etapas (chamá-los derruba o teste
// com panic).
type etapaGetRepoStub struct {
	empServices.EtapaRepositoryInterface
	etapas []*empModels.Etapa
}

func (s *etapaGetRepoStub) GetByID(_ context.Context, id uuid.UUID) (*empModels.Etapa, error) {
	for _, e := range s.etapas {
		if e.ID == id {
			copied := *e
			return &copied, nil
		}
	}
	return nil, nil
}

func (s *etapaGetRepoStub) ListByVaga(_ context.Context, vagaID uuid.UUID) ([]*empModels.Etapa, error) {
	out := []*empModels.Etapa{}
	for _, e := range s.etapas {
		if e.IDVaga == vagaID {
			copied := *e
			out = append(out, &copied)
		}
	}
	return out, nil
}

// GET /api/v1/empregabilidade/vagas/:id/etapas e /etapas/:etapaId seguem a
// regra do GET /vagas/:id: vagaAuth e, para vaga ainda não publicada, o órgão
// parceiro no escopo do chamador (DEV-137).
func TestRegisterEmpregabilidadeRoutes_EtapaGet_Authorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rascunhoID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	aprovacaoID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	publicadaID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	inexistenteID := uuid.MustParse("44444444-4444-4444-8444-444444444444")

	etapaRascunhoID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	etapaAprovacaoID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	etapaPublicadaID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	etapaInexistenteID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")

	vagaRepo := &vagaGetRepoStub{vagas: map[uuid.UUID]*empModels.Vaga{
		rascunhoID:  {ID: rascunhoID, Titulo: "Rascunho", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaEmEdicao},
		aprovacaoID: {ID: aprovacaoID, Titulo: "Em aprovação", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaEmAprovacao},
		publicadaID: {ID: publicadaID, Titulo: "Publicada", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaPublicadoAtivo},
	}}
	vagaSvc := empServices.NewVagaServiceWithInterfaces(vagaRepo, nil, nil)

	etapaRepo := &etapaGetRepoStub{etapas: []*empModels.Etapa{
		{ID: etapaRascunhoID, IDVaga: rascunhoID, Titulo: "Triagem", Ordem: 1},
		{ID: etapaAprovacaoID, IDVaga: aprovacaoID, Titulo: "Entrevista", Ordem: 1},
		{ID: etapaPublicadaID, IDVaga: publicadaID, Titulo: "Prova", Ordem: 1},
	}}
	etapaSvc := empServices.NewEtapaServiceWithInterface(etapaRepo)

	listURL := func(vagaID uuid.UUID) string {
		return "/api/v1/empregabilidade/vagas/" + vagaID.String() + "/etapas"
	}
	getURL := func(vagaID, etapaID uuid.UUID) string {
		return listURL(vagaID) + "/" + etapaID.String()
	}

	editor := []string{"go:empregabilidade:editor"}
	cidadao := injectCaller("", nil, "")
	editorOutroOrgao := injectCaller("", editor, "orgao-a")
	editorDoOrgao := injectCaller("", editor, "orgao-b")
	empAdmin := injectCaller("", []string{"go:empregabilidade:admin"}, "")
	admin := injectCaller("ADMIN", nil, "")

	tests := []struct {
		name   string
		caller gin.HandlerFunc
		url    string
		want   int
	}{
		// Listagem
		{"lista: cidadão sem role, vaga em edição", cidadao, listURL(rascunhoID), http.StatusForbidden},
		{"lista: cidadão sem role, vaga em aprovação", cidadao, listURL(aprovacaoID), http.StatusForbidden},
		{"lista: cidadão sem role, vaga publicada", cidadao, listURL(publicadaID), http.StatusForbidden},
		{"lista: editor de outro órgão, vaga em edição", editorOutroOrgao, listURL(rascunhoID), http.StatusForbidden},
		{"lista: editor de outro órgão, vaga em aprovação", editorOutroOrgao, listURL(aprovacaoID), http.StatusForbidden},
		{"lista: editor de outro órgão, vaga publicada", editorOutroOrgao, listURL(publicadaID), http.StatusOK},
		{"lista: editor do órgão, vaga em edição", editorDoOrgao, listURL(rascunhoID), http.StatusOK},
		{"lista: editor do órgão, vaga em aprovação", editorDoOrgao, listURL(aprovacaoID), http.StatusOK},
		{"lista: emp admin, vaga em edição", empAdmin, listURL(rascunhoID), http.StatusOK},
		{"lista: admin, vaga em aprovação", admin, listURL(aprovacaoID), http.StatusOK},
		{"lista: editor do órgão, vaga inexistente", editorDoOrgao, listURL(inexistenteID), http.StatusNotFound},

		// Busca por ID
		{"etapa: cidadão sem role, vaga em edição", cidadao, getURL(rascunhoID, etapaRascunhoID), http.StatusForbidden},
		{"etapa: cidadão sem role, vaga publicada", cidadao, getURL(publicadaID, etapaPublicadaID), http.StatusForbidden},
		{"etapa: editor de outro órgão, vaga em edição", editorOutroOrgao, getURL(rascunhoID, etapaRascunhoID), http.StatusForbidden},
		{"etapa: editor de outro órgão, vaga em aprovação", editorOutroOrgao, getURL(aprovacaoID, etapaAprovacaoID), http.StatusForbidden},
		{"etapa: editor de outro órgão, vaga publicada", editorOutroOrgao, getURL(publicadaID, etapaPublicadaID), http.StatusOK},
		{"etapa: editor do órgão, vaga em edição", editorDoOrgao, getURL(rascunhoID, etapaRascunhoID), http.StatusOK},
		{"etapa: emp admin, vaga em aprovação", empAdmin, getURL(aprovacaoID, etapaAprovacaoID), http.StatusOK},
		{"etapa: admin, vaga em edição", admin, getURL(rascunhoID, etapaRascunhoID), http.StatusOK},
		{"etapa: editor do órgão, vaga inexistente", editorDoOrgao, getURL(inexistenteID, etapaRascunhoID), http.StatusNotFound},
		{"etapa: editor do órgão, etapa inexistente", editorDoOrgao, getURL(rascunhoID, etapaInexistenteID), http.StatusNotFound},
		// A vaga do path é publicada, mas a etapa é de uma vaga em edição: não
		// pode contornar a regra de leitura da vaga não publicada.
		{"etapa: editor de outro órgão, etapa de vaga em edição sob vaga publicada", editorOutroOrgao, getURL(publicadaID, etapaRascunhoID), http.StatusNotFound},
		{"etapa: admin, etapa de outra vaga", admin, getURL(publicadaID, etapaRascunhoID), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(tt.caller)
			app := createMockAppContainer()
			app.EmpVagaService = vagaSvc
			app.EmpVagaHandler = empHandlers.NewVagaHandler(vagaSvc)
			app.EmpEtapaService = etapaSvc
			app.EmpEtapaHandler = empHandlers.NewEtapaHandler(etapaSvc)
			cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: true}}
			registerEmpregabilidadeRoutes(router.Group("/api/v1"), router.Group("/api/public"), app, cfg)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.url, nil))
			assert.Equal(t, tt.want, w.Code, w.Body.String())
		})
	}
}
