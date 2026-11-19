package empregabilidade_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaixaEtaria_LimitesIdade(t *testing.T) {
	minimo, maximo, ok := empregabilidade.FaixaEtariaMaioresDe18.LimitesIdade()
	require.True(t, ok)
	assert.Equal(t, 18, minimo)
	assert.Nil(t, maximo, "Maiores de 18 não tem teto")

	// Mesma faixa do critério Jovem Aprendiz da vaga.
	minimo, maximo, ok = empregabilidade.FaixaEtariaJovemAprendiz.LimitesIdade()
	require.True(t, ok)
	assert.Equal(t, empregabilidade.IdadeMinimaJovemAprendiz, minimo)
	require.NotNil(t, maximo)
	assert.Equal(t, empregabilidade.IdadeMaximaJovemAprendiz, *maximo)
}

func TestFaixaEtaria_IsValid(t *testing.T) {
	assert.True(t, empregabilidade.FaixaEtariaMaioresDe18.IsValid())
	assert.True(t, empregabilidade.FaixaEtariaJovemAprendiz.IsValid())
	for _, invalida := range []empregabilidade.FaixaEtaria{"", "30", "idosos", "MAIORES_18"} {
		assert.False(t, invalida.IsValid(), invalida)
	}
}

func TestFiltroCandidato_VazioComFaixaEtaria(t *testing.T) {
	assert.True(t, empregabilidade.FiltroCandidato{}.Vazio())
	assert.False(t, empregabilidade.FiltroCandidato{FaixaEtaria: empregabilidade.FaixaEtariaJovemAprendiz}.Vazio())
}
