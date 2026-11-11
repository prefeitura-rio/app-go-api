package empregabilidade

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// O histórico de candidaturas (DEV-532) não guarda dado pessoal: só ids,
// códigos e datas.
func TestCandidaturaHistorico_SemDadoPessoal(t *testing.T) {
	proibidos := []string{"cpf", "nome", "email", "telefone", "endereco"}
	tipo := reflect.TypeOf(CandidaturaHistorico{})
	for i := 0; i < tipo.NumField(); i++ {
		campo := tipo.Field(i)
		for _, proibido := range proibidos {
			assert.NotContains(t, strings.ToLower(campo.Name), proibido, "campo %s", campo.Name)
			assert.NotContains(t, strings.ToLower(campo.Tag.Get("json")), proibido, "campo %s", campo.Name)
		}
	}
	assert.Equal(t, "emp_candidatura_historico", CandidaturaHistorico{}.TableName())
}

func TestTipoEventoPorStatus(t *testing.T) {
	assert.Equal(t, EventoCandidaturaAprovada, TipoEventoPorStatus(StatusCandidaturaAprovada))
	assert.Equal(t, EventoCandidaturaReprovada, TipoEventoPorStatus(StatusCandidaturaReprovada))
	assert.Equal(t, EventoStatusAlterado, TipoEventoPorStatus(StatusCandidaturaEnviada))
	assert.Equal(t, EventoStatusAlterado, TipoEventoPorStatus(StatusCandidaturaVagaCongelada))
	assert.Equal(t, EventoStatusAlterado, TipoEventoPorStatus(StatusCandidaturaDescontinuada))
}

func TestOrigemEventoDoContexto(t *testing.T) {
	_, ok := OrigemEventoDoContexto(context.Background())
	assert.False(t, ok, "contexto sem origem")

	_, ok = OrigemEventoDoContexto(ComOrigemEvento(context.Background(), OrigemEventoCandidatura{}))
	assert.False(t, ok, "origem sem ator não vale")

	origem := OrigemEventoCandidatura{Ator: AtorAutomacao, Motivo: MotivoVagaCongelada}
	lida, ok := OrigemEventoDoContexto(ComOrigemEvento(context.Background(), origem))
	assert.True(t, ok)
	assert.Equal(t, origem, lida)
}
