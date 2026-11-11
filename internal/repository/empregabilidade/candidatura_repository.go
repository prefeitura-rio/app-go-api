package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type BulkUpdateResult struct {
	Updated    int
	FailedCPFs []string
}

type CandidaturaRepository struct {
	db *gorm.DB
}

func NewCandidaturaRepository(db *gorm.DB) *CandidaturaRepository {
	return &CandidaturaRepository{db: db}
}

// Create grava a candidatura e, na mesma transação, o evento candidatura_criada
// do histórico, atribuído à origem do contexto.
func (r *CandidaturaRepository) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	origem := origemEvento(ctx)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return fmt.Errorf("erro ao criar candidatura: %w", err)
		}
		status := entity.Status
		if status == "" {
			status = empregabilidade.StatusCandidaturaEnviada
		}
		return inserirEventos(tx, []empregabilidade.CandidaturaHistorico{{
			IDCandidatura: entity.ID,
			IDVaga:        entity.IDVaga,
			TipoEvento:    empregabilidade.EventoCandidaturaCriada,
			Ator:          origem.Ator,
			StatusNovo:    status,
			IDEtapaNova:   copiarUUID(entity.IDEtapaAtual),
			Motivo:        motivoEvento(origem),
		}})
	})
	if err != nil {
		return uuid.Nil, err
	}
	return entity.ID, nil
}

func (r *CandidaturaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	var entity empregabilidade.Candidatura
	result := r.db.WithContext(ctx).
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("Vaga.Etapas").
		Preload("Vaga.InformacoesComplementares").
		Preload("EtapaAtual").
		First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar candidatura: %w", result.Error)
	}
	return &entity, nil
}

// Update grava a edição genérica da candidatura. Status e etapa ficam de fora:
// só mudam pelas transições, que registram o histórico; regravá-los aqui
// desfaria uma transição concorrente sem evento.
func (r *CandidaturaRepository) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	result := r.db.WithContext(ctx).Omit("status", "status_anterior", "id_etapa_atual").Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar candidatura: %w", result.Error)
	}
	return nil
}

// Delete exclui (soft delete) a candidatura e registra, na mesma transação, o
// evento candidatura_cancelada.
func (r *CandidaturaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id = ?", id)
		if err != nil {
			return err
		}
		if len(estados) == 0 {
			return fmt.Errorf("candidatura não encontrada")
		}
		result := tx.Delete(&empregabilidade.Candidatura{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("erro ao excluir candidatura: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("candidatura não encontrada")
		}
		evento := eventoStatus(estados[0], estados[0].Status, origem)
		evento.TipoEvento = empregabilidade.EventoCandidaturaCancelada
		return inserirEventos(tx, []empregabilidade.CandidaturaHistorico{evento})
	})
}

func (r *CandidaturaRepository) List(ctx context.Context, filter empregabilidade.CandidaturaFilter, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	var entities []*empregabilidade.Candidatura
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.CPF != "" {
			db = db.Where("cpf = ?", filter.CPF)
		}
		if filter.VagaID != nil {
			db = db.Where("id_vaga = ?", *filter.VagaID)
		}
		if filter.Status != "" {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.EtapaID != nil {
			db = db.Where("id_etapa_atual = ?", *filter.EtapaID)
		}
		if filter.Search != "" {
			searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
		}
		if condicao, args := condicaoFiltroCandidato(filter.Candidato, cpfDigitosCandidatura); condicao != "" {
			db = db.Where(condicao, args...)
		}
		return restringirAoEscopoDeOrgao(db, filter.OrgaoParceiroIDs)
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar candidaturas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}))
	result := findDB.
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("Vaga.Etapas").
		Preload("Vaga.InformacoesComplementares").
		Preload("EtapaAtual").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar candidaturas: %w", result.Error)
	}

	return entities, int(total), nil
}

// restringirAoEscopoDeOrgao limita às candidaturas de vagas dos órgãos parceiros
// do escopo do editor: nil não restringe (admin), lista vazia não devolve nada.
func restringirAoEscopoDeOrgao(db *gorm.DB, orgaoParceiroIDs []string) *gorm.DB {
	if orgaoParceiroIDs == nil {
		return db
	}
	if len(orgaoParceiroIDs) == 0 {
		return db.Where("1 = 0")
	}
	return db.Where("id_vaga IN (SELECT id FROM emp_vagas WHERE id_orgao_parceiro IN ?)", orgaoParceiroIDs)
}

// BulkUpdateStatus muda o status das candidaturas da vaga com esses CPFs e
// registra, na mesma transação, um evento por candidatura cujo status mudou.
func (r *CandidaturaRepository) BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (BulkUpdateResult, error) {
	var result BulkUpdateResult
	origem := origemEvento(ctx)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id_vaga = ? AND cpf IN ?", vagaID, cpfs)
		if err != nil {
			return err
		}

		porCPF := make(map[string]estadoCandidatura, len(estados))
		for _, estado := range estados {
			porCPF[estado.CPF] = estado
		}

		var updateIDs []uuid.UUID
		var eventos []empregabilidade.CandidaturaHistorico
		// CPF repetido no pedido é a mesma candidatura: um só evento.
		vistas := make(map[uuid.UUID]struct{}, len(estados))
		for _, cpf := range cpfs {
			estado, found := porCPF[cpf]
			if !found {
				result.FailedCPFs = append(result.FailedCPFs, cpf)
				continue
			}
			if _, repetida := vistas[estado.ID]; repetida {
				continue
			}
			vistas[estado.ID] = struct{}{}
			updateIDs = append(updateIDs, estado.ID)
			if estado.Status != status {
				eventos = append(eventos, eventoStatus(estado, status, origem))
			}
		}

		if len(updateIDs) == 0 {
			return nil
		}

		res := tx.Model(&empregabilidade.Candidatura{}).
			Where("id IN ?", updateIDs).
			Update("status", status)
		if res.Error != nil {
			return fmt.Errorf("erro ao atualizar status em lote: %w", res.Error)
		}
		result.Updated = int(res.RowsAffected)

		return inserirEventos(tx, eventos)
	})
	if err != nil {
		return BulkUpdateResult{}, err
	}
	return result, nil
}

func (r *CandidaturaRepository) CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("cpf = ? AND id_vaga = ?", cpf, vagaID).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("erro ao verificar candidatura existente: %w", result.Error)
	}
	return count > 0, nil
}

// UpdateStatus muda o status da candidatura e registra, na mesma transação, o
// evento da mudança (nenhum quando o status já era esse).
func (r *CandidaturaRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id = ?", id)
		if err != nil {
			return err
		}
		result := tx.Model(&empregabilidade.Candidatura{}).
			Where("id = ?", id).
			Update("status", status)
		if result.Error != nil {
			return fmt.Errorf("erro ao atualizar status da candidatura: %w", result.Error)
		}
		if len(estados) == 0 || estados[0].Status == status {
			return nil
		}
		return inserirEventos(tx, []empregabilidade.CandidaturaHistorico{eventoStatus(estados[0], status, origem)})
	})
}

func (r *CandidaturaRepository) BulkGetByCPFs(ctx context.Context, vagaID uuid.UUID, cpfs []string) ([]*empregabilidade.Candidatura, error) {
	var candidaturas []*empregabilidade.Candidatura
	// Vaga e Contratante são lidos pelos e-mails de aprovação/reprovação em lote.
	if err := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Where("id_vaga = ? AND cpf IN ?", vagaID, cpfs).
		Find(&candidaturas).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar candidaturas: %w", err)
	}
	return candidaturas, nil
}

// BulkUpdateEtapa leva as candidaturas ids para a etapa etapaID e registra, na
// mesma transação, um evento (avanço ou volta) por candidatura cuja etapa mudou.
func (r *CandidaturaRepository) BulkUpdateEtapa(ctx context.Context, ids []uuid.UUID, etapaID uuid.UUID) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id IN ?", ids)
		if err != nil {
			return err
		}
		result := tx.Model(&empregabilidade.Candidatura{}).
			Where("id IN ?", ids).
			Update("id_etapa_atual", etapaID)
		if result.Error != nil {
			return fmt.Errorf("erro ao atualizar etapa em lote: %w", result.Error)
		}
		return registrarMudancasEtapa(tx, estados, etapaID, origem)
	})
}

// UpdateEtapa leva a candidatura para a etapa etapaID e registra, na mesma
// transação, o evento de avanço ou volta (nenhum quando a etapa já era essa).
func (r *CandidaturaRepository) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id = ?", id)
		if err != nil {
			return err
		}
		result := tx.Model(&empregabilidade.Candidatura{}).
			Where("id = ?", id).
			Update("id_etapa_atual", etapaID)
		if result.Error != nil {
			return fmt.Errorf("erro ao atualizar etapa da candidatura: %w", result.Error)
		}
		return registrarMudancasEtapa(tx, estados, etapaID, origem)
	})
}

// registrarMudancasEtapa grava um evento para cada candidatura de estados que
// não estava na etapa etapaID.
func registrarMudancasEtapa(tx *gorm.DB, estados []estadoCandidatura, etapaID uuid.UUID, origem empregabilidade.OrigemEventoCandidatura) error {
	var mudaram []estadoCandidatura
	for _, estado := range estados {
		if estado.IDEtapaAtual == nil || *estado.IDEtapaAtual != etapaID {
			mudaram = append(mudaram, estado)
		}
	}
	if len(mudaram) == 0 {
		return nil
	}
	tipos, err := tiposMudancaEtapa(tx, mudaram, etapaID)
	if err != nil {
		return err
	}
	eventos := make([]empregabilidade.CandidaturaHistorico, 0, len(mudaram))
	for _, estado := range mudaram {
		eventos = append(eventos, eventoEtapa(estado, etapaID, tipos[estado.ID], origem))
	}
	return inserirEventos(tx, eventos)
}

// BulkSaveAndUpdateStatusByVagaID guarda o status atual em status_anterior,
// aplica newStatus a todas as candidaturas da vaga e registra, na mesma
// transação, um evento por candidatura cujo status mudou (automação, com o
// motivo do contexto).
func (r *CandidaturaRepository) BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, newStatus empregabilidade.StatusCandidatura) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id_vaga = ?", vagaID)
		if err != nil {
			return err
		}
		// Save current status into status_anterior, then set new status
		result := tx.Exec(
			`UPDATE emp_candidaturas SET status_anterior = status, status = ?, updated_at = NOW() WHERE id_vaga = ? AND deleted_at IS NULL`,
			string(newStatus), vagaID,
		)
		if result.Error != nil {
			return fmt.Errorf("erro ao congelar/descontinuar candidaturas: %w", result.Error)
		}
		var eventos []empregabilidade.CandidaturaHistorico
		for _, estado := range estados {
			if estado.Status != newStatus {
				eventos = append(eventos, eventoStatus(estado, newStatus, origem))
			}
		}
		return inserirEventos(tx, eventos)
	})
}

func (r *CandidaturaRepository) CountByStatus(ctx context.Context, filter empregabilidade.CandidaturaFilter) (map[empregabilidade.StatusCandidatura]int64, error) {
	type statusCount struct {
		Status empregabilidade.StatusCandidatura
		Count  int64
	}
	var results []statusCount

	db := r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{})

	if filter.CPF != "" {
		db = db.Where("cpf = ?", filter.CPF)
	}
	if filter.VagaID != nil {
		db = db.Where("id_vaga = ?", *filter.VagaID)
	}
	if filter.EtapaID != nil {
		db = db.Where("id_etapa_atual = ?", *filter.EtapaID)
	}
	if filter.Search != "" {
		searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
		db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
	}
	if condicao, args := condicaoFiltroCandidato(filter.Candidato, cpfDigitosCandidatura); condicao != "" {
		db = db.Where(condicao, args...)
	}
	db = restringirAoEscopoDeOrgao(db, filter.OrgaoParceiroIDs)
	// Status filter is intentionally ignored to count all statuses

	if err := db.Select("status, COUNT(*) as count").Group("status").Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar candidaturas por status: %w", err)
	}

	counts := map[empregabilidade.StatusCandidatura]int64{
		empregabilidade.StatusCandidaturaEnviada:       0,
		empregabilidade.StatusCandidaturaAprovada:      0,
		empregabilidade.StatusCandidaturaReprovada:     0,
		empregabilidade.StatusCandidaturaVagaCongelada: 0,
		empregabilidade.StatusCandidaturaDescontinuada: 0,
	}
	for _, r := range results {
		counts[r.Status] = r.Count
	}

	return counts, nil
}

// BulkRestoreStatusByVagaID devolve às candidaturas da vaga o status guardado em
// status_anterior e registra, na mesma transação, um evento por candidatura
// cujo status mudou (automação, com o motivo do contexto).
func (r *CandidaturaRepository) BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error {
	origem := origemEvento(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estados, err := travarCandidaturas(tx, "id_vaga = ? AND status_anterior IS NOT NULL", vagaID)
		if err != nil {
			return err
		}
		result := tx.Exec(
			`UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW() WHERE id_vaga = ? AND deleted_at IS NULL AND status_anterior IS NOT NULL`,
			vagaID,
		)
		if result.Error != nil {
			return fmt.Errorf("erro ao restaurar status das candidaturas: %w", result.Error)
		}
		var eventos []empregabilidade.CandidaturaHistorico
		for _, estado := range estados {
			if estado.StatusAnterior != nil && *estado.StatusAnterior != estado.Status {
				eventos = append(eventos, eventoStatus(estado, *estado.StatusAnterior, origem))
			}
		}
		return inserirEventos(tx, eventos)
	})
}
