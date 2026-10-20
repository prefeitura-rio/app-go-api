package empregabilidade_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBancoCurriculoHandler_List_FiltrosAvancados(t *testing.T) {
	idioma, nivel := uuid.New(), uuid.New()
	svc := &fakeBancoCurriculoService{items: []*empmodels.BancoCurriculoItem{}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos?genero=Mulher%20cisg%C3%AAnero,Mulher%20transg%C3%AAnero&genero=N%C3%A3o%20bin%C3%A1rio"+
		"&idade_min=18&idade_max=30&tempo_experiencia=2_a_4_anos&experiencia_carteira=6_meses"+
		"&escolaridade=M%C3%A9dio%20completo&id_idioma="+idioma.String()+"&id_nivel_idioma="+nivel.String()+
		"&bairro=Andara%C3%AD&pcd=sim")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	f := svc.gotFilter.Candidato
	assert.Equal(t, []string{"Mulher cisgênero", "Mulher transgênero", "Não binário"}, f.Generos)
	require.NotNil(t, f.IdadeMinima)
	assert.Equal(t, 18, *f.IdadeMinima)
	require.NotNil(t, f.IdadeMaxima)
	assert.Equal(t, 30, *f.IdadeMaxima)
	assert.Equal(t, empmodels.Faixa2a4Anos, f.TempoExperiencia)
	assert.Equal(t, empmodels.CarteiraSeisMeses, f.Carteira)
	assert.Equal(t, "Médio completo", f.Escolaridade)
	assert.Equal(t, &idioma, f.IDIdioma)
	assert.Equal(t, &nivel, f.IDNivelIdiomaMinimo)
	assert.Equal(t, "Andaraí", f.Bairro)
	require.NotNil(t, f.PCD)
	assert.True(t, *f.PCD)
}

func TestBancoCurriculoHandler_List_SemFiltros(t *testing.T) {
	svc := &fakeBancoCurriculoService{items: []*empmodels.BancoCurriculoItem{}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos?pcd=nao")

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, svc.gotFilter.Candidato.PCD)
	assert.False(t, *svc.gotFilter.Candidato.PCD)
	svc.gotFilter.Candidato.PCD = nil
	assert.True(t, svc.gotFilter.Candidato.Vazio())
}

func TestBancoCurriculoHandler_List_FiltroInvalido(t *testing.T) {
	cases := map[string]string{
		"idade não numérica":       "idade_min=abc",
		"idade negativa":           "idade_max=-1",
		"faixa de idade invertida": "idade_min=40&idade_max=30",
		"faixa de experiência":     "tempo_experiencia=muito",
		"carteira":                 "experiencia_carteira=talvez",
		"idioma inválido":          "id_idioma=xyz",
		"nível sem idioma":         "id_nivel_idioma=" + uuid.NewString(),
		"pcd fora de sim/nao":      "pcd=talvez",
	}
	for nome, query := range cases {
		t.Run(nome, func(t *testing.T) {
			svc := &fakeBancoCurriculoService{}
			w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos?"+query)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Zero(t, svc.gotPage, "não chega ao service")
		})
	}
}
