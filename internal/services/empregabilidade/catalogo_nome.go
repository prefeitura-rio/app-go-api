package empregabilidade

import (
	"strings"
	"unicode/utf8"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
)

const catalogoNomeMaxRunes = 250

func normalizeCatalogoNome(nome string) (string, error) {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return "", apperrors.Validation("nome é obrigatório")
	}
	if utf8.RuneCountInString(nome) > catalogoNomeMaxRunes {
		return "", apperrors.Validation("nome excede o tamanho máximo de 250 caracteres")
	}
	return nome, nil
}
