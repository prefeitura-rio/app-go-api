package empregabilidade_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// recordingHabilidadeRepo registra o limite repassado ao repositório pelas listas de catálogo.
type recordingHabilidadeRepo struct {
	mockHabilidadeRepo
	lastLimit int
}

func (m *recordingHabilidadeRepo) ListHabilidades(ctx context.Context, f empmodels.HabilidadeFilter, limit, offset int) ([]*empmodels.Habilidade, int64, error) {
	m.lastLimit = limit
	return m.mockHabilidadeRepo.ListHabilidades(ctx, f, limit, offset)
}

func (m *recordingHabilidadeRepo) ListAreasAtuacao(ctx context.Context, f empmodels.AreaAtuacaoFilter, limit, offset int) ([]*empmodels.AreaAtuacao, int64, error) {
	m.lastLimit = limit
	return m.mockHabilidadeRepo.ListAreasAtuacao(ctx, f, limit, offset)
}

func (m *recordingHabilidadeRepo) ListHabilidadesDaArea(ctx context.Context, areaID int64, f empmodels.HabilidadeFilter, limit, offset int) ([]*empmodels.CatalogoItem, int64, error) {
	m.lastLimit = limit
	return m.mockHabilidadeRepo.ListHabilidadesDaArea(ctx, areaID, f, limit, offset)
}

type recordingComportamentoAtitudesRepo struct {
	mockComportamentoAtitudesRepo
	lastLimit int
}

func (m *recordingComportamentoAtitudesRepo) ListComportamentoAtitudes(ctx context.Context, f empmodels.ComportamentoAtitudesFilter, limit, offset int) ([]*empmodels.ComportamentoAtitudes, int64, error) {
	m.lastLimit = limit
	return m.mockComportamentoAtitudesRepo.ListComportamentoAtitudes(ctx, f, limit, offset)
}

func TestCatalogoList_PageSize_LimitaA100(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pageSizes := []struct {
		query string
		want  int
	}{
		{query: "0", want: 20},
		{query: "-5", want: 20},
		{query: "invalid", want: 20},
		{query: "1", want: 1},
		{query: "100", want: 100},
		{query: "101", want: 100},
		{query: "150", want: 100},
	}

	endpoints := []struct {
		name  string
		path  string
		setup func() (*gin.Engine, func() int)
	}{
		{
			name: "habilidades",
			path: "/habilidades",
			setup: func() (*gin.Engine, func() int) {
				repo := &recordingHabilidadeRepo{}
				h := handler.NewHabilidadeHandler(services.NewHabilidadeServiceWithInterface(repo), nil)
				r := gin.New()
				r.GET("/habilidades", h.ListHabilidades)
				return r, func() int { return repo.lastLimit }
			},
		},
		{
			name: "areas de atuacao",
			path: "/habilidades/areas-atuacao",
			setup: func() (*gin.Engine, func() int) {
				repo := &recordingHabilidadeRepo{}
				h := handler.NewHabilidadeHandler(services.NewHabilidadeServiceWithInterface(repo), nil)
				r := gin.New()
				r.GET("/habilidades/areas-atuacao", h.ListAreasAtuacao)
				return r, func() int { return repo.lastLimit }
			},
		},
		{
			name: "conhecimentos da area",
			path: "/habilidades/areas-atuacao/10/habilidades",
			setup: func() (*gin.Engine, func() int) {
				repo := &recordingHabilidadeRepo{}
				h := handler.NewHabilidadeHandler(services.NewHabilidadeServiceWithInterface(repo), nil)
				r := gin.New()
				r.GET("/habilidades/areas-atuacao/:id/habilidades", h.ListHabilidadesDaArea)
				return r, func() int { return repo.lastLimit }
			},
		},
		{
			name: "comportamentos e atitudes",
			path: "/comportamentos-atitudes",
			setup: func() (*gin.Engine, func() int) {
				repo := &recordingComportamentoAtitudesRepo{}
				h := handler.NewComportamentoAtitudesHandler(services.NewComportamentoAtitudesService(repo))
				r := gin.New()
				r.GET("/comportamentos-atitudes", h.ListComportamentoAtitudes)
				return r, func() int { return repo.lastLimit }
			},
		},
	}

	for _, ep := range endpoints {
		for _, ps := range pageSizes {
			t.Run(ep.name+"/pageSize="+ps.query, func(t *testing.T) {
				r, lastLimit := ep.setup()

				req := httptest.NewRequest(http.MethodGet, ep.path+"?page=1&pageSize="+ps.query, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				var body struct {
					PageSize int `json:"pageSize"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, ps.want, body.PageSize, "pageSize na resposta")
				assert.Equal(t, ps.want, lastLimit(), "limite repassado ao repositório")
			})
		}
	}
}
