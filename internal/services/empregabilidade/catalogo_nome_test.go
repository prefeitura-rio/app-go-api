package empregabilidade

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCatalogoNome(t *testing.T) {
	t.Run("trims spaces", func(t *testing.T) {
		got, err := normalizeCatalogoNome("  Comunicação  ")
		require.NoError(t, err)
		assert.Equal(t, "Comunicação", got)
	})

	t.Run("empty after trim", func(t *testing.T) {
		_, err := normalizeCatalogoNome("   ")
		assert.EqualError(t, err, "nome é obrigatório")
	})

	t.Run("too long", func(t *testing.T) {
		_, err := normalizeCatalogoNome(strings.Repeat("a", 251))
		assert.EqualError(t, err, "nome excede o tamanho máximo de 250 caracteres")
	})

	t.Run("max length", func(t *testing.T) {
		nome := strings.Repeat("á", 250)
		got, err := normalizeCatalogoNome(nome)
		require.NoError(t, err)
		assert.Equal(t, nome, got)
	})
}
