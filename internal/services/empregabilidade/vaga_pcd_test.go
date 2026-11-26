package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cnpjPCD = "12.345.678/0001-90"

func acessibilidadePCD(a empregabilidade.AcessibilidadePCD) *empregabilidade.AcessibilidadePCD {
	return &a
}

func tiposPCD(n int) []empregabilidade.TipoPCD {
	tipos := make([]empregabilidade.TipoPCD, 0, n)
	for i := 0; i < n; i++ {
		tipos = append(tipos, empregabilidade.TipoPCD{ID: uuid.New()})
	}
	return tipos
}

func requireErroValidacao(t *testing.T, err error, msg string) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperrors.Error
	require.True(t, errors.As(err, &appErr), "erro precisa ser de validação: %v", err)
	assert.Equal(t, apperrors.KindInvalid, appErr.Kind)
	assert.Contains(t, err.Error(), msg)
}

func TestVagaService_Create_OpcaoPCD(t *testing.T) {
	tests := []struct {
		name              string
		opcao             empregabilidade.OpcaoPCD
		acessibilidade    *empregabilidade.AcessibilidadePCD
		tipos             []empregabilidade.TipoPCD
		wantOpcao         empregabilidade.OpcaoPCD
		wantAcessibilidad *empregabilidade.AcessibilidadePCD
		wantErr           string
	}{
		{name: "sem nada é aberta", wantOpcao: empregabilidade.OpcaoPCDAberta},
		{name: "aberta limpa o campo antigo", opcao: empregabilidade.OpcaoPCDAberta, acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD), wantOpcao: empregabilidade.OpcaoPCDAberta},
		{name: "aberta com tipos", opcao: empregabilidade.OpcaoPCDAberta, tipos: tiposPCD(1), wantOpcao: empregabilidade.OpcaoPCDAberta},
		{name: "preferencial sem tipos", opcao: empregabilidade.OpcaoPCDPreferencial, wantOpcao: empregabilidade.OpcaoPCDPreferencial, wantAcessibilidad: acessibilidadePCD(empregabilidade.AcessibilidadePreferencialPCD)},
		{name: "exclusiva com tipos", opcao: empregabilidade.OpcaoPCDExclusiva, tipos: tiposPCD(2), wantOpcao: empregabilidade.OpcaoPCDExclusiva, wantAcessibilidad: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD)},
		{name: "exclusiva sem tipos é recusada", opcao: empregabilidade.OpcaoPCDExclusiva, wantErr: empregabilidade.MsgTipoPCDObrigatorio},
		{name: "exclusiva com lista vazia é recusada", opcao: empregabilidade.OpcaoPCDExclusiva, tipos: []empregabilidade.TipoPCD{}, wantErr: empregabilidade.MsgTipoPCDObrigatorio},
		{name: "opção desconhecida é recusada", opcao: "exclusivo_pcd", wantErr: "opcao_pcd inválida"},
		{name: "cliente antigo: exclusivo_pcd vira exclusiva", acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD), tipos: tiposPCD(1), wantOpcao: empregabilidade.OpcaoPCDExclusiva, wantAcessibilidad: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD)},
		{name: "cliente antigo: preferencial_pcd vira preferencial", acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadePreferencialPCD), wantOpcao: empregabilidade.OpcaoPCDPreferencial, wantAcessibilidad: acessibilidadePCD(empregabilidade.AcessibilidadePreferencialPCD)},
		{name: "cliente antigo: para_pcd vira preferencial", acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadeParaPCD), wantOpcao: empregabilidade.OpcaoPCDPreferencial, wantAcessibilidad: acessibilidadePCD(empregabilidade.AcessibilidadeParaPCD)},
		{name: "cliente antigo: exclusiva sem tipos é recusada", acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD), wantErr: empregabilidade.MsgTipoPCDObrigatorio},
		{name: "cliente antigo: acessibilidade desconhecida é recusada", acessibilidade: acessibilidadePCD("todos"), wantErr: "acessibilidade_pcd inválido"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjPCD)
			vaga := &empregabilidade.Vaga{
				Titulo:            "Repositor",
				IDContratante:     cnpjPCD,
				OpcaoPCD:          tt.opcao,
				AcessibilidadePCD: tt.acessibilidade,
				TiposPCD:          tt.tipos,
			}

			id, err := svc.Create(context.Background(), vaga)
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Equal(t, uuid.Nil, id)
				assert.Empty(t, vagaRepo.vagas, "vaga inválida não pode ser gravada")
				return
			}
			require.NoError(t, err)
			saved := vagaRepo.vagas[id]
			require.NotNil(t, saved)
			assert.Equal(t, tt.wantOpcao, saved.OpcaoPCD)
			assert.Equal(t, tt.wantAcessibilidad, saved.AcessibilidadePCD)
		})
	}
}

func TestVagaService_Update_OpcaoPCD(t *testing.T) {
	tests := []struct {
		name           string
		atualOpcao     empregabilidade.OpcaoPCD
		atualTipos     []empregabilidade.TipoPCD
		opcao          empregabilidade.OpcaoPCD
		acessibilidade *empregabilidade.AcessibilidadePCD
		tipos          []empregabilidade.TipoPCD
		wantOpcao      empregabilidade.OpcaoPCD
		wantErr        string
	}{
		{name: "exclusiva com tipos no body", atualOpcao: empregabilidade.OpcaoPCDAberta, opcao: empregabilidade.OpcaoPCDExclusiva, tipos: tiposPCD(1), wantOpcao: empregabilidade.OpcaoPCDExclusiva},
		{name: "exclusiva sem tipos no body usa os gravados", atualOpcao: empregabilidade.OpcaoPCDPreferencial, atualTipos: tiposPCD(1), opcao: empregabilidade.OpcaoPCDExclusiva, wantOpcao: empregabilidade.OpcaoPCDExclusiva},
		{name: "exclusiva sem tipos em lugar nenhum é recusada", atualOpcao: empregabilidade.OpcaoPCDAberta, opcao: empregabilidade.OpcaoPCDExclusiva, wantErr: empregabilidade.MsgTipoPCDObrigatorio},
		{name: "exclusiva tirando todos os tipos é recusada", atualOpcao: empregabilidade.OpcaoPCDExclusiva, atualTipos: tiposPCD(2), opcao: empregabilidade.OpcaoPCDExclusiva, tipos: []empregabilidade.TipoPCD{}, wantErr: empregabilidade.MsgTipoPCDObrigatorio},
		{name: "volta para aberta", atualOpcao: empregabilidade.OpcaoPCDExclusiva, atualTipos: tiposPCD(1), opcao: empregabilidade.OpcaoPCDAberta, wantOpcao: empregabilidade.OpcaoPCDAberta},
		{name: "opção desconhecida é recusada", atualOpcao: empregabilidade.OpcaoPCDAberta, opcao: "Exclusiva", wantErr: "opcao_pcd inválida"},
		{name: "cliente antigo sem acessibilidade fica aberta", atualOpcao: empregabilidade.OpcaoPCDPreferencial, wantOpcao: empregabilidade.OpcaoPCDAberta},
		{name: "cliente antigo com exclusivo_pcd e tipos gravados", atualOpcao: empregabilidade.OpcaoPCDExclusiva, atualTipos: tiposPCD(1), acessibilidade: acessibilidadePCD(empregabilidade.AcessibilidadeExclusivoPCD), wantOpcao: empregabilidade.OpcaoPCDExclusiva},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjPCD)
			vagaID := uuid.New()
			original := &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Repositor",
				Status:            empregabilidade.StatusVagaEmEdicao,
				OpcaoPCD:          tt.atualOpcao,
				AcessibilidadePCD: tt.atualOpcao.AcessibilidadeLegada(),
				TiposPCD:          tt.atualTipos,
			}
			vagaRepo.vagas[vagaID] = original

			err := svc.Update(context.Background(), &empregabilidade.Vaga{
				ID:                vagaID,
				Titulo:            "Repositor II",
				OpcaoPCD:          tt.opcao,
				AcessibilidadePCD: tt.acessibilidade,
				TiposPCD:          tt.tipos,
			})
			if tt.wantErr != "" {
				requireErroValidacao(t, err, tt.wantErr)
				assert.Same(t, original, vagaRepo.vagas[vagaID], "vaga não pode ser alterada")
				return
			}
			require.NoError(t, err)
			saved := vagaRepo.vagas[vagaID]
			assert.Equal(t, tt.wantOpcao, saved.OpcaoPCD)
			assert.Equal(t, tt.wantOpcao, empregabilidade.OpcaoPCDDeAcessibilidade(saved.AcessibilidadePCD),
				"campo antigo precisa continuar coerente com opcao_pcd")
		})
	}
}

func TestVagaService_UpdateTiposPCD_ExclusivaExigeTipo(t *testing.T) {
	tests := []struct {
		name      string
		opcao     empregabilidade.OpcaoPCD
		ids       []uuid.UUID
		wantErr   bool
		wantCalls int
	}{
		{name: "exclusiva sem tipos é recusada", opcao: empregabilidade.OpcaoPCDExclusiva, ids: []uuid.UUID{}, wantErr: true},
		{name: "exclusiva com nil é recusada", opcao: empregabilidade.OpcaoPCDExclusiva, ids: nil, wantErr: true},
		{name: "exclusiva com tipo", opcao: empregabilidade.OpcaoPCDExclusiva, ids: []uuid.UUID{uuid.New()}, wantCalls: 1},
		{name: "preferencial pode ficar sem tipos", opcao: empregabilidade.OpcaoPCDPreferencial, ids: []uuid.UUID{}, wantCalls: 1},
		{name: "aberta pode ficar sem tipos", opcao: empregabilidade.OpcaoPCDAberta, ids: nil, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, vagaRepo, _ := newVagaServiceForQtd(cnpjPCD)
			vagaID := uuid.New()
			vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{ID: vagaID, OpcaoPCD: tt.opcao, TiposPCD: tiposPCD(1)}

			err := svc.UpdateTiposPCD(context.Background(), vagaID, tt.ids)
			if tt.wantErr {
				requireErroValidacao(t, err, empregabilidade.MsgTipoPCDObrigatorio)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, vagaRepo.updateTiposPCDCalls, tt.wantCalls)
		})
	}
}

func TestVagaService_UpdateTiposPCD_ErroAoLerVaga(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjPCD)
	vagaRepo.getError = errors.New("banco fora")

	err := svc.UpdateTiposPCD(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.Empty(t, vagaRepo.updateTiposPCDCalls)
}

func TestVagaService_Update_CriterioBloqueioPCD(t *testing.T) {
	svc, vagaRepo, _ := newVagaServiceForQtd(cnpjPCD)
	vagaID := uuid.New()
	vagaRepo.vagas[vagaID] = &empregabilidade.Vaga{
		ID:       vagaID,
		Status:   empregabilidade.StatusVagaEmEdicao,
		OpcaoPCD: empregabilidade.OpcaoPCDExclusiva,
		TiposPCD: tiposPCD(1),
	}

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                vagaID,
		Titulo:            "Repositor",
		OpcaoPCD:          empregabilidade.OpcaoPCDExclusiva,
		CriteriosBloqueio: []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioPCD},
	})
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.CriterioBloqueio{empregabilidade.CriterioBloqueioPCD}, vagaRepo.vagas[vagaID].CriteriosBloqueio)
}
