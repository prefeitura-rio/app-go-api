package empregabilidade_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
)

func TestVaga_Slug_FormatoCorreto(t *testing.T) {
	id := uuid.MustParse("f3d23675-97e5-4d57-8892-bff6ba805d6d")
	v := &empregabilidade.Vaga{ID: id, Titulo: "Analista de TI Júnior"}
	assert.Equal(t, "analista-de-ti-junior-f3d23675", v.Slug())
}

func TestVaga_Slug_TerminaComOitoCharsDoUUID(t *testing.T) {
	id := uuid.New()
	v := &empregabilidade.Vaga{ID: id, Titulo: "Desenvolvedor Go"}
	parts := strings.Split(v.Slug(), "-")
	suffix := parts[len(parts)-1]
	assert.Equal(t, id.String()[:8], suffix)
}

func TestVaga_Slug_RemoveAcentos(t *testing.T) {
	id := uuid.MustParse("aaaabbbb-0000-0000-0000-000000000000")
	v := &empregabilidade.Vaga{ID: id, Titulo: "Técnico em Administração"}
	assert.True(t, strings.HasPrefix(v.Slug(), "tecnico-em-administracao-"))
}

func TestVaga_Slug_UnicoPorUUID(t *testing.T) {
	titulo := "Mesmo Título"
	v1 := &empregabilidade.Vaga{ID: uuid.New(), Titulo: titulo}
	v2 := &empregabilidade.Vaga{ID: uuid.New(), Titulo: titulo}
	assert.NotEqual(t, v1.Slug(), v2.Slug())
}
