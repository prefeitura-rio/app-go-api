package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func exigenciaCarteira(e empregabilidade.ExperienciaCarteira) *empregabilidade.ExperienciaCarteira {
	return &e
}

func TestVagaService_Create_ExperienciaCarteira(t *testing.T) {
	carteira := empregabilidade.CriterioBloqueioExperienciaCarteira
	tests := []struct {
		name      string
		criterios []empregabilidade.CriterioBloqueio
		enviado   *empregabilidade.ExperienciaCarteira
		want      *empregabilidade.ExperienciaCarteira
		wantErr   string
	}{
		{name: "sem exigência e sem bloqueio", want: nil},
		{
			name:      "bloqueio ligado com mínimo de 6 meses",
			criterios: criterios(carteira),
			enviado:   exigenciaCarteira(empregabilidade.ExperienciaCarteiraMinimo6Meses),
			want:      exigenciaCarteira(empregabilidade.ExperienciaCarteiraMinimo6Meses),
		},
		{
			name:      "bloqueio ligado com qualquer tempo",
			criterios: criterios(carteira),
			enviado:   exigenciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo),
			want:      exigenciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo),
		},
		{
			name:    "opção sem bloqueio ligado é gravada",
			enviado: exigenciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo),
			want:    exigenciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo),
		},
		{name: "string vazia é sem exigência", enviado: exigenciaCarteira(""), want: nil},
		{
			name:    "opção desconhecida é recusada",
			enviado: exigenciaCarteira("minimo_12_meses"),
			wantErr: "experiencia_carteira inválida: deve ser qualquer_tempo ou minimo_6_meses",
		},
		{
			name:      "bloqueio ligado sem opção é recusado",
			criterios: criterios(carteira),
			wantErr:   empregabilidade.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:      "bloqueio ligado com string vazia é recusado",
			criterios: criterios(carteira),
			enviado:   exigenciaCarteira(""),
			wantErr:   empregabilidade.MsgExperienciaCarteiraObrigatoria,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
			vaga := &empregabilidade.Vaga{
				Titulo:              "Repositor",
				IDContratante:       cnpjCriteriosBloqueio,
				CriteriosBloqueio:   tt.criterios,
				ExperienciaCarteira: tt.enviado,
			}

			id, err := svc.Create(context.Background(), vaga)
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Equal(t, uuid.Nil, id)
				assert.Empty(t, vagaRepo.vagas, "vaga inválida não pode ser gravada")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, vagaRepo.vagas[id].ExperienciaCarteira)
		})
	}
}

func TestVagaService_Update_ExperienciaCarteira(t *testing.T) {
	carteira := empregabilidade.CriterioBloqueioExperienciaCarteira
	seisMeses := exigenciaCarteira(empregabilidade.ExperienciaCarteiraMinimo6Meses)
	qualquer := exigenciaCarteira(empregabilidade.ExperienciaCarteiraQualquerTempo)

	tests := []struct {
		name           string
		atualCriterios []empregabilidade.CriterioBloqueio
		atual          *empregabilidade.ExperienciaCarteira
		criterios      []empregabilidade.CriterioBloqueio
		enviado        *empregabilidade.ExperienciaCarteira
		want           *empregabilidade.ExperienciaCarteira
		wantCriterios  []empregabilidade.CriterioBloqueio
		wantErr        string
	}{
		{
			name:           "campo ausente mantém a opção e o bloqueio gravados",
			atualCriterios: criterios(carteira),
			atual:          seisMeses,
			want:           seisMeses,
			wantCriterios:  criterios(carteira),
		},
		{
			name:           "troca a opção",
			atualCriterios: criterios(carteira),
			atual:          seisMeses,
			enviado:        qualquer,
			want:           qualquer,
			wantCriterios:  criterios(carteira),
		},
		{
			name:          "liga o bloqueio com a opção",
			criterios:     criterios(carteira),
			enviado:       seisMeses,
			want:          seisMeses,
			wantCriterios: criterios(carteira),
		},
		{
			name:           "desliga o bloqueio e mantém a opção",
			atualCriterios: criterios(carteira),
			atual:          seisMeses,
			criterios:      criterios(),
			want:           seisMeses,
			wantCriterios:  nil,
		},
		{
			name:           "string vazia remove a exigência com o bloqueio desligado",
			atualCriterios: criterios(carteira),
			atual:          seisMeses,
			criterios:      criterios(),
			enviado:        exigenciaCarteira(""),
			want:           nil,
			wantCriterios:  nil,
		},
		{
			name:           "string vazia com o bloqueio mantido ligado é recusada",
			atualCriterios: criterios(carteira),
			atual:          seisMeses,
			enviado:        exigenciaCarteira(""),
			wantErr:        empregabilidade.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:      "liga o bloqueio numa vaga sem opção é recusado",
			criterios: criterios(carteira),
			wantErr:   empregabilidade.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:    "opção desconhecida é recusada",
			atual:   seisMeses,
			enviado: exigenciaCarteira("6_meses"),
			wantErr: "experiencia_carteira inválida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
			vagaID := uuid.New()
			gravada := &empregabilidade.Vaga{
				ID:                  vagaID,
				Titulo:              "Repositor",
				Status:              empregabilidade.StatusVagaEmEdicao,
				CriteriosBloqueio:   tt.atualCriterios,
				ExperienciaCarteira: tt.atual,
			}
			vagaRepo.vagas[vagaID] = gravada

			err := svc.Update(context.Background(), &empregabilidade.Vaga{
				ID:                  vagaID,
				Titulo:              "Repositor II",
				CriteriosBloqueio:   tt.criterios,
				ExperienciaCarteira: tt.enviado,
			})
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Same(t, gravada, vagaRepo.vagas[vagaID], "vaga inválida não pode ser gravada")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, vagaRepo.vagas[vagaID].ExperienciaCarteira)
			assert.Equal(t, tt.wantCriterios, vagaRepo.vagas[vagaID].CriteriosBloqueio)
		})
	}
}
