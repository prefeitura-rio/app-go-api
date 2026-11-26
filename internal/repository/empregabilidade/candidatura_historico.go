package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// tamanhoLoteHistorico limita as linhas de um INSERT de eventos em lote (vaga
// congelada com milhares de candidaturas), abaixo do limite de parâmetros do
// PostgreSQL.
const tamanhoLoteHistorico = 500

// estadoCandidatura é o estado de uma candidatura lido, com a linha travada,
// antes de uma transição. O CPF serve só para o resultado das operações em lote
// por CPF e nunca vai para o histórico.
type estadoCandidatura struct {
	ID             uuid.UUID
	IDVaga         uuid.UUID
	CPF            string
	Status         empregabilidade.StatusCandidatura
	StatusAnterior *empregabilidade.StatusCandidatura
	IDEtapaAtual   *uuid.UUID
}

// travarCandidaturas lê e trava (SELECT ... FOR UPDATE) as candidaturas não
// excluídas que atendem à condição, em ordem de id para evitar deadlock entre
// operações em lote concorrentes.
func travarCandidaturas(tx *gorm.DB, query string, args ...interface{}) ([]estadoCandidatura, error) {
	var estados []estadoCandidatura
	err := tx.Model(&empregabilidade.Candidatura{}).
		Select("id", "id_vaga", "cpf", "status", "status_anterior", "id_etapa_atual").
		Where(query, args...).
		Order("id").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scan(&estados).Error
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar candidaturas: %w", err)
	}
	return estados, nil
}

// origemEvento é quem provocou a transição, lido do contexto. Sem origem no
// contexto, a transição é do operador: as rotas que mudam status e etapa são
// de gestão; candidato e automação são declarados por quem chama.
func origemEvento(ctx context.Context) empregabilidade.OrigemEventoCandidatura {
	if origem, ok := empregabilidade.OrigemEventoDoContexto(ctx); ok {
		return origem
	}
	return empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorOperador}
}

func motivoEvento(origem empregabilidade.OrigemEventoCandidatura) *string {
	if origem.Motivo == "" {
		return nil
	}
	motivo := origem.Motivo
	return &motivo
}

// eventoStatus registra a mudança de status de estado para novo; a etapa fica
// a atual nos dois lados.
func eventoStatus(estado estadoCandidatura, novo empregabilidade.StatusCandidatura, origem empregabilidade.OrigemEventoCandidatura) empregabilidade.CandidaturaHistorico {
	anterior := estado.Status
	return empregabilidade.CandidaturaHistorico{
		IDCandidatura:   estado.ID,
		IDVaga:          estado.IDVaga,
		TipoEvento:      empregabilidade.TipoEventoPorStatus(novo),
		Ator:            origem.Ator,
		StatusAnterior:  &anterior,
		StatusNovo:      novo,
		IDEtapaAnterior: copiarUUID(estado.IDEtapaAtual),
		IDEtapaNova:     copiarUUID(estado.IDEtapaAtual),
		Motivo:          motivoEvento(origem),
	}
}

// eventoEtapa registra a ida de estado para a etapa nova; o status fica o atual
// nos dois lados.
func eventoEtapa(estado estadoCandidatura, etapaNova uuid.UUID, tipo empregabilidade.TipoEventoCandidatura, origem empregabilidade.OrigemEventoCandidatura) empregabilidade.CandidaturaHistorico {
	status := estado.Status
	nova := etapaNova
	return empregabilidade.CandidaturaHistorico{
		IDCandidatura:   estado.ID,
		IDVaga:          estado.IDVaga,
		TipoEvento:      tipo,
		Ator:            origem.Ator,
		StatusAnterior:  &status,
		StatusNovo:      status,
		IDEtapaAnterior: copiarUUID(estado.IDEtapaAtual),
		IDEtapaNova:     &nova,
		Motivo:          motivoEvento(origem),
	}
}

func copiarUUID(id *uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	copia := *id
	return &copia
}

// inserirEventos grava os eventos na transação tx; sem eventos, não faz nada.
func inserirEventos(tx *gorm.DB, eventos []empregabilidade.CandidaturaHistorico) error {
	if len(eventos) == 0 {
		return nil
	}
	if err := tx.CreateInBatches(&eventos, tamanhoLoteHistorico).Error; err != nil {
		return fmt.Errorf("erro ao registrar histórico da candidatura: %w", err)
	}
	return nil
}

// tiposMudancaEtapa decide, para cada candidatura que muda para etapaNova, se
// é avanço (etapa de ordem maior, ou a primeira etapa) ou volta (ordem menor).
func tiposMudancaEtapa(tx *gorm.DB, estados []estadoCandidatura, etapaNova uuid.UUID) (map[uuid.UUID]empregabilidade.TipoEventoCandidatura, error) {
	ids := []uuid.UUID{etapaNova}
	vistos := map[uuid.UUID]struct{}{etapaNova: {}}
	for _, estado := range estados {
		if estado.IDEtapaAtual == nil {
			continue
		}
		if _, ok := vistos[*estado.IDEtapaAtual]; ok {
			continue
		}
		vistos[*estado.IDEtapaAtual] = struct{}{}
		ids = append(ids, *estado.IDEtapaAtual)
	}

	var etapas []struct {
		ID    uuid.UUID
		Ordem int
	}
	if err := tx.Model(&empregabilidade.Etapa{}).Select("id", "ordem").Where("id IN ?", ids).Scan(&etapas).Error; err != nil {
		return nil, fmt.Errorf("erro ao ler ordem das etapas: %w", err)
	}
	ordem := make(map[uuid.UUID]int, len(etapas))
	for _, etapa := range etapas {
		ordem[etapa.ID] = etapa.Ordem
	}

	tipos := make(map[uuid.UUID]empregabilidade.TipoEventoCandidatura, len(estados))
	for _, estado := range estados {
		tipos[estado.ID] = empregabilidade.EventoMudancaEtapa
		if estado.IDEtapaAtual == nil {
			continue
		}
		ordemAtual, okAtual := ordem[*estado.IDEtapaAtual]
		ordemNova, okNova := ordem[etapaNova]
		if okAtual && okNova && ordemNova < ordemAtual {
			tipos[estado.ID] = empregabilidade.EventoVoltaEtapa
		}
	}
	return tipos, nil
}

// registrarEtapasRemovidas grava, para cada candidatura que ficou sem etapa
// porque a etapa em que estava foi excluída da vaga, o evento etapa_removida
// (automação).
func registrarEtapasRemovidas(tx *gorm.DB, estados []estadoCandidatura) error {
	origem := empregabilidade.OrigemEventoCandidatura{Ator: empregabilidade.AtorAutomacao, Motivo: empregabilidade.MotivoEtapaExcluida}
	eventos := make([]empregabilidade.CandidaturaHistorico, 0, len(estados))
	for _, estado := range estados {
		evento := eventoStatus(estado, estado.Status, origem)
		evento.TipoEvento = empregabilidade.EventoEtapaRemovida
		evento.IDEtapaNova = nil
		eventos = append(eventos, evento)
	}
	return inserirEventos(tx, eventos)
}
