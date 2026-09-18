package empregabilidade

import (
	"testing"

	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

func TestShouldSendProximaEtapaEmail(t *testing.T) {
	etapa1 := &empmodels.Etapa{Ordem: 1}
	etapa2 := &empmodels.Etapa{Ordem: 2}

	tests := []struct {
		name    string
		status  empmodels.StatusCandidatura
		current *empmodels.Etapa
		next    *empmodels.Etapa
		want    bool
	}{
		{"first assignment", empmodels.StatusCandidaturaEnviada, nil, etapa1, true},
		{"forward", empmodels.StatusCandidaturaEnviada, etapa1, etapa2, true},
		{"backward", empmodels.StatusCandidaturaEnviada, etapa2, etapa1, false},
		{"same ordem", empmodels.StatusCandidaturaEnviada, etapa1, &empmodels.Etapa{Ordem: 1}, false},
		{"aprovada", empmodels.StatusCandidaturaAprovada, etapa1, etapa2, false},
		{"reprovada", empmodels.StatusCandidaturaReprovada, etapa1, etapa2, false},
		{"nil next", empmodels.StatusCandidaturaEnviada, etapa1, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSendProximaEtapaEmail(tt.status, tt.current, tt.next); got != tt.want {
				t.Errorf("shouldSendProximaEtapaEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
