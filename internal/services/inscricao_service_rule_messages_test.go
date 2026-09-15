package services_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// As mensagens de EnrollmentRuleError são devolvidas ao portal em 400 e exibidas
// ao cidadão como estão. Elas não podem carregar vocabulário interno — nome de
// campo do payload, de coluna ou de tipo. Uma delas nasceu assim
// ("schedule_id fornecido não pertence a este curso") e só não vazou antes
// porque ficava escondida atrás de um 500 com mensagem genérica no portal.
//
// Este teste lê o fonte e barra a reintrodução do problema.
func TestEnrollmentRuleMessagesAreCitizenFacing(t *testing.T) {
	source, err := os.ReadFile("inscricao_service.go")
	if err != nil {
		t.Fatalf("não foi possível ler o fonte do serviço: %v", err)
	}

	// Captura o literal de cada enrollmentRuleErrorf("...")
	re := regexp.MustCompile(`enrollmentRuleErrorf\("([^"]*)"`)
	matches := re.FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		t.Fatal("nenhuma mensagem encontrada — o teste perdeu o alvo, revise o regexp")
	}

	// Termos que denunciam implementação em vez de falar com o cidadão
	forbidden := []string{
		"_id",
		"schedule",
		"payload",
		"json",
		"uuid",
		"struct",
		"nil",
		"query",
		"column",
		"coluna",
	}

	for _, match := range matches {
		message := match[1]
		lower := strings.ToLower(message)
		for _, term := range forbidden {
			if strings.Contains(lower, term) {
				t.Errorf(
					"mensagem exibida ao cidadão contém termo interno %q: %q",
					term, message,
				)
			}
		}
	}

	t.Logf("%d mensagens verificadas", len(matches))
}
