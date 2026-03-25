package models_test

import (
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func TestInscricao_TableName(t *testing.T) {
	inscricao := &models.Inscricao{}
	if got := inscricao.TableName(); got != "inscricoes" {
		t.Errorf("Inscricao.TableName() = %v, want inscricoes", got)
	}
}

func TestStatusInscricao_Values(t *testing.T) {
	// Test that status constants are properly defined
	tests := []struct {
		name   string
		status models.StatusInscricao
		want   string
	}{
		{"Pending", models.StatusInscricaoPending, "pending"},
		{"Approved", models.StatusInscricaoApproved, "approved"},
		{"Rejected", models.StatusInscricaoRejected, "rejected"},
		{"Cancelled", models.StatusInscricaoCancelled, "cancelled"},
		{"Concluded", models.StatusInscricaoConcluded, "concluded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("StatusInscricao %v = %v, want %v", tt.name, tt.status, tt.want)
			}
		})
	}
}
