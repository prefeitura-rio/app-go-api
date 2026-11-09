package empregabilidade_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func opcaoJovemAprendiz(v bool) *bool { return &v }

func TestVagaService_Create_JovemAprendiz(t *testing.T) {
	jovem := empregabilidade.CriterioBloqueioJovemAprendiz
	idade18 := 18
	tests := []struct {
		name        string
		criterios   []empregabilidade.CriterioBloqueio
		enviado     *bool
		idadeMinima *int
		want        bool
		wantErr     string
	}{
		{name: "sem o campo nasce desligado", want: false},
		{name: "opção marcada sem bloqueio é gravada", enviado: opcaoJovemAprendiz(true), want: true},
		{name: "opção desmarcada", enviado: opcaoJovemAprendiz(false), want: false},
		{
			name:      "bloqueio ligado com a opção marcada",
			criterios: criterios(jovem),
			enviado:   opcaoJovemAprendiz(true),
			want:      true,
		},
		{
			name:        "Maiores de 18 (idade_minima) segue igual ao lado de Jovem Aprendiz",
			enviado:     opcaoJovemAprendiz(true),
			idadeMinima: &idade18,
			want:        true,
		},
		{
			name:      "bloqueio ligado sem a opção é recusado",
			criterios: criterios(jovem),
			wantErr:   empregabilidade.MsgJovemAprendizObrigatorio,
		},
		{
			name:      "bloqueio ligado com a opção desmarcada é recusado",
			criterios: criterios(jovem),
			enviado:   opcaoJovemAprendiz(false),
			wantErr:   empregabilidade.MsgJovemAprendizObrigatorio,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
			vaga := &empregabilidade.Vaga{
				Titulo:            "Auxiliar administrativo",
				IDContratante:     cnpjCriteriosBloqueio,
				CriteriosBloqueio: tt.criterios,
				JovemAprendiz:     tt.enviado,
				IdadeMinima:       tt.idadeMinima,
			}

			id, err := svc.Create(context.Background(), vaga)
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Equal(t, uuid.Nil, id)
				assert.Empty(t, vagaRepo.vagas, "vaga inválida não pode ser gravada")
				return
			}
			require.NoError(t, err)
			gravada := vagaRepo.vagas[id]
			require.NotNil(t, gravada.JovemAprendiz, "o campo é sempre gravado")
			assert.Equal(t, tt.want, *gravada.JovemAprendiz)
			assert.Equal(t, tt.idadeMinima, gravada.IdadeMinima, "idade_minima não muda")
		})
	}
}

func TestVagaService_Update_JovemAprendiz(t *testing.T) {
	jovem := empregabilidade.CriterioBloqueioJovemAprendiz

	tests := []struct {
		name           string
		atualCriterios []empregabilidade.CriterioBloqueio
		atual          *bool
		criterios      []empregabilidade.CriterioBloqueio
		enviado        *bool
		want           bool
		wantCriterios  []empregabilidade.CriterioBloqueio
		wantErr        string
	}{
		{
			name:           "campo ausente mantém a opção e o bloqueio gravados",
			atualCriterios: criterios(jovem),
			atual:          opcaoJovemAprendiz(true),
			want:           true,
			wantCriterios:  criterios(jovem),
		},
		{
			name:  "vaga existente sem o campo segue desligada",
			atual: opcaoJovemAprendiz(false),
			want:  false,
		},
		{
			name:          "marca a opção e liga o bloqueio",
			criterios:     criterios(jovem),
			enviado:       opcaoJovemAprendiz(true),
			want:          true,
			wantCriterios: criterios(jovem),
		},
		{
			name:          "liga o bloqueio com a opção já gravada",
			atual:         opcaoJovemAprendiz(true),
			criterios:     criterios(jovem),
			want:          true,
			wantCriterios: criterios(jovem),
		},
		{
			name:           "desliga o bloqueio e mantém a opção",
			atualCriterios: criterios(jovem),
			atual:          opcaoJovemAprendiz(true),
			criterios:      criterios(),
			want:           true,
		},
		{
			name:           "desmarca a opção com o bloqueio desligado",
			atualCriterios: criterios(jovem),
			atual:          opcaoJovemAprendiz(true),
			criterios:      criterios(),
			enviado:        opcaoJovemAprendiz(false),
			want:           false,
		},
		{
			name:           "desmarca a opção com o bloqueio mantido ligado é recusado",
			atualCriterios: criterios(jovem),
			atual:          opcaoJovemAprendiz(true),
			enviado:        opcaoJovemAprendiz(false),
			wantErr:        empregabilidade.MsgJovemAprendizObrigatorio,
		},
		{
			name:      "liga o bloqueio numa vaga sem a opção é recusado",
			atual:     opcaoJovemAprendiz(false),
			criterios: criterios(jovem),
			wantErr:   empregabilidade.MsgJovemAprendizObrigatorio,
		},
		{
			name:      "liga o bloqueio numa vaga lida sem valor é recusado",
			criterios: criterios(jovem),
			wantErr:   empregabilidade.MsgJovemAprendizObrigatorio,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjCriteriosBloqueio)
			vagaID := uuid.New()
			gravada := &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Auxiliar administrativo",
				Status:            empregabilidade.StatusVagaEmEdicao,
				CriteriosBloqueio: tt.atualCriterios,
				JovemAprendiz:     tt.atual,
			}
			vagaRepo.vagas[vagaID] = gravada

			err := svc.Update(context.Background(), &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Auxiliar administrativo II",
				CriteriosBloqueio: tt.criterios,
				JovemAprendiz:     tt.enviado,
			})
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Same(t, gravada, vagaRepo.vagas[vagaID], "vaga inválida não pode ser gravada")
				return
			}
			require.NoError(t, err)
			atualizada := vagaRepo.vagas[vagaID]
			require.NotNil(t, atualizada.JovemAprendiz, "o campo é sempre gravado")
			assert.Equal(t, tt.want, *atualizada.JovemAprendiz)
			assert.Equal(t, tt.wantCriterios, atualizada.CriteriosBloqueio)
		})
	}
}
