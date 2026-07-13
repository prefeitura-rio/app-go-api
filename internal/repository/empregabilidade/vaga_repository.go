package empregabilidade

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

var nonDigit = regexp.MustCompile(`\D`)

type VagaRepository struct {
	db *gorm.DB
}

func NewVagaRepository(db *gorm.DB) *VagaRepository {
	return &VagaRepository{db: db}
}

func (r *VagaRepository) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar vaga: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *VagaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	var entity empregabilidade.Vaga
	result := r.db.WithContext(ctx).
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Preload("TiposPCD").
		Preload("Zonas").
		Preload("IdiomasRequisito").
		Preload("Etapas", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Preload("InformacoesComplementares").
		First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar vaga: %w", result.Error)
	}
	return &entity, nil
}

func (r *VagaRepository) Update(ctx context.Context, entity *empregabilidade.Vaga) error {
	updates := map[string]interface{}{
		"titulo":                           entity.Titulo,
		"descricao":                        entity.Descricao,
		"id_contratante":                   entity.IDContratante,
		"id_regime_contratacao":            entity.IDRegimeContratacao,
		"id_modelo_trabalho":               entity.IDModeloTrabalho,
		"acessibilidade_pcd":               entity.AcessibilidadePCD,
		"valor_vaga":                       entity.ValorVaga,
		"quantidade_estimada_contratacoes": entity.QuantidadeEstimadaContratacoes,
		"bairro":                           entity.Bairro,
		"data_limite":                      entity.DataLimite,
		"requisitos":                       entity.Requisitos,
		"diferenciais":                     entity.Diferenciais,
		"responsabilidades":                entity.Responsabilidades,
		"beneficios":                       entity.Beneficios,
		"id_orgao_parceiro":                entity.IDOrgaoParceiro,
		"status":                           entity.Status,
		"idade_minima":                     entity.IdadeMinima,
		"idade_maxima":                     entity.IdadeMaxima,
		"bairros_elegibilidade":            entity.BairrosElegibilidade,
		"id_escolaridade_minima":           entity.IDEscolaridadeMinima,
		"areas_formacao_elegibilidade":     entity.AreasFormacaoElegibilidade,
	}

	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Vaga{}).
		Where("id = ?", entity.ID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar vaga: %w", result.Error)
	}
	return nil
}

func (r *VagaRepository) UpdateWithAssociations(ctx context.Context, entity *empregabilidade.Vaga) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"titulo":                           entity.Titulo,
			"descricao":                        entity.Descricao,
			"id_contratante":                   entity.IDContratante,
			"id_regime_contratacao":            entity.IDRegimeContratacao,
			"id_modelo_trabalho":               entity.IDModeloTrabalho,
			"acessibilidade_pcd":               entity.AcessibilidadePCD,
			"valor_vaga":                       entity.ValorVaga,
			"quantidade_estimada_contratacoes": entity.QuantidadeEstimadaContratacoes,
			"bairro":                           entity.Bairro,
			"data_limite":                      entity.DataLimite,
			"requisitos":                       entity.Requisitos,
			"diferenciais":                     entity.Diferenciais,
			"responsabilidades":                entity.Responsabilidades,
			"beneficios":                       entity.Beneficios,
			"id_orgao_parceiro":                entity.IDOrgaoParceiro,
			"status":                           entity.Status,
			"idade_minima":                     entity.IdadeMinima,
			"idade_maxima":                     entity.IdadeMaxima,
			"bairros_elegibilidade":            entity.BairrosElegibilidade,
			"id_escolaridade_minima":           entity.IDEscolaridadeMinima,
			"areas_formacao_elegibilidade":     entity.AreasFormacaoElegibilidade,
		}

		if err := tx.Model(&empregabilidade.Vaga{}).Where("id = ?", entity.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("erro ao atualizar vaga: %w", err)
		}

		if entity.Etapas != nil {
			// Clear id_etapa_atual from all candidaturas of this vaga before deleting the steps.
			// We must clear all statuses (not only active ones) because the FK has no ON DELETE SET NULL.
			if err := tx.Model(&empregabilidade.Candidatura{}).
				Where("id_etapa_atual IN (SELECT id FROM emp_etapas WHERE id_vaga = ?)", entity.ID).
				Update("id_etapa_atual", nil).Error; err != nil {
				return fmt.Errorf("erro ao desvincular etapas das candidaturas: %w", err)
			}

			if err := tx.Where("id_vaga = ?", entity.ID).Delete(&empregabilidade.Etapa{}).Error; err != nil {
				return fmt.Errorf("erro ao remover etapas existentes: %w", err)
			}

			for i := range entity.Etapas {
				entity.Etapas[i].ID = uuid.Nil
				entity.Etapas[i].IDVaga = entity.ID
				entity.Etapas[i].Vaga = nil
			}

			if len(entity.Etapas) > 0 {
				if err := tx.Create(&entity.Etapas).Error; err != nil {
					return fmt.Errorf("erro ao criar etapas: %w", err)
				}
			}
		}

		if entity.InformacoesComplementares != nil {
			if err := syncInformacoesComplementares(tx, entity.ID, entity.InformacoesComplementares); err != nil {
				return err
			}
		}

		if entity.TiposPCD != nil {
			if err := tx.Exec("DELETE FROM emp_vagas_tipos_pcd WHERE id_vaga = ?", entity.ID).Error; err != nil {
				return fmt.Errorf("erro ao remover tipos PCD: %w", err)
			}
			for _, tipo := range entity.TiposPCD {
				if err := tx.Exec("INSERT INTO emp_vagas_tipos_pcd (id_vaga, id_tipo_pcd) VALUES (?, ?)", entity.ID, tipo.ID).Error; err != nil {
					return fmt.Errorf("erro ao inserir tipo PCD: %w", err)
				}
			}
		}

		if entity.Zonas != nil {
			if err := tx.Exec("DELETE FROM emp_vagas_zonas WHERE id_vaga = ?", entity.ID).Error; err != nil {
				return fmt.Errorf("erro ao remover zonas: %w", err)
			}
			for _, zona := range entity.Zonas {
				if err := tx.Exec("INSERT INTO emp_vagas_zonas (id_vaga, id_zona) VALUES (?, ?)", entity.ID, zona.ID).Error; err != nil {
					return fmt.Errorf("erro ao inserir zona: %w", err)
				}
			}
		}

		return nil
	})
}

// syncInformacoesComplementares reconcilia as informações complementares de uma vaga
// pelo UUID em vez de deletar e recriar tudo: itens com ID preenchido são atualizados
// in-place (preservando o UUID referenciado pelas respostas de candidaturas já
// submetidas); itens ausentes do payload são removidos; itens com ID == uuid.Nil são
// inseridos como novos, recebendo um UUID gerado pelo banco.
func syncInformacoesComplementares(tx *gorm.DB, vagaID uuid.UUID, informacoes []empregabilidade.InformacaoComplementar) error {
	existentesIDs := make([]uuid.UUID, 0, len(informacoes))
	for i := range informacoes {
		if informacoes[i].ID != uuid.Nil {
			existentesIDs = append(existentesIDs, informacoes[i].ID)
		}
	}

	deleteQuery := tx.Where("id_vaga = ?", vagaID)
	if len(existentesIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", existentesIDs)
	}
	if err := deleteQuery.Delete(&empregabilidade.InformacaoComplementar{}).Error; err != nil {
		return fmt.Errorf("erro ao remover informações complementares existentes: %w", err)
	}

	var novos []empregabilidade.InformacaoComplementar
	for i := range informacoes {
		informacoes[i].IDVaga = vagaID
		informacoes[i].Vaga = nil

		if informacoes[i].ID == uuid.Nil {
			novos = append(novos, informacoes[i])
			continue
		}

		// Update baseado em struct (não em map) é obrigatório aqui: apenas o update
		// baseado em struct aplica o serializer:json do GORM ao campo Opcoes. Um
		// map[string]interface{} ignora o serializer e expande o []string em uma
		// expressão SQL inválida para a coluna jsonb. Select(...) força a
		// atualização de todos os campos listados mesmo quando estão em zero-value
		// (ex.: Obrigatorio == false), que de outra forma seria omitido pelo GORM.
		if err := tx.Model(&empregabilidade.InformacaoComplementar{}).
			Where("id = ? AND id_vaga = ?", informacoes[i].ID, vagaID).
			Select("titulo", "obrigatorio", "tipo_campo", "valor_minimo", "valor_maximo", "opcoes").
			Updates(&informacoes[i]).Error; err != nil {
			return fmt.Errorf("erro ao atualizar informação complementar: %w", err)
		}
	}

	if len(novos) > 0 {
		if err := tx.Create(&novos).Error; err != nil {
			return fmt.Errorf("erro ao criar informações complementares: %w", err)
		}

		novoIdx := 0
		for i := range informacoes {
			if informacoes[i].ID == uuid.Nil {
				informacoes[i].ID = novos[novoIdx].ID
				novoIdx++
			}
		}
	}

	return nil
}

func (r *VagaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Vaga{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir vaga: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("vaga não encontrada")
	}
	return nil
}

func (r *VagaRepository) List(ctx context.Context, filter empregabilidade.VagaFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Status != "" {
			switch filter.Status {
			case string(empregabilidade.StatusVagaPublicadoAtivo):
				db = db.Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)
			case string(empregabilidade.StatusVagaPublicadoExpirado):
				db = db.Where("status = ? AND data_limite IS NOT NULL AND data_limite <= NOW()", empregabilidade.StatusVagaPublicadoAtivo)
			default:
				db = db.Where("status = ?", filter.Status)
			}
		}
		if filter.Contratante != "" {
			db = db.Where("id_contratante = ?", filter.Contratante)
		}
		switch {
		case len(filter.OrgaoParceiroIDs) == 1:
			db = db.Where("id_orgao_parceiro = ?", filter.OrgaoParceiroIDs[0])
		case len(filter.OrgaoParceiroIDs) > 1:
			db = db.Where("id_orgao_parceiro IN ?", filter.OrgaoParceiroIDs)
		case filter.OrgaoParceiroID != "":
			db = db.Where("id_orgao_parceiro = ?", filter.OrgaoParceiroID)
		}
		if filter.Search != "" {
			db = db.Where("titulo ILIKE ?", fmt.Sprintf("%%%s%%", filter.Search))
		}
		switch filter.DataPublicacao {
		case empregabilidade.DataPublicacaoHoje:
			db = db.Where("created_at >= ?", time.Now().Truncate(24*time.Hour))
		case empregabilidade.DataPublicacaoUltimaSemana:
			db = db.Where("created_at >= ?", time.Now().AddDate(0, 0, -7))
		case empregabilidade.DataPublicacaoUltimoMes:
			db = db.Where("created_at >= ?", time.Now().AddDate(0, 0, -30))
		}
		if len(filter.IDRegimeContratacao) == 1 {
			db = db.Where("id_regime_contratacao = ?", filter.IDRegimeContratacao[0])
		} else if len(filter.IDRegimeContratacao) > 1 {
			db = db.Where("id_regime_contratacao IN ?", filter.IDRegimeContratacao)
		}
		if len(filter.IDModeloTrabalho) == 1 {
			db = db.Where("id_modelo_trabalho = ?", filter.IDModeloTrabalho[0])
		} else if len(filter.IDModeloTrabalho) > 1 {
			db = db.Where("id_modelo_trabalho IN ?", filter.IDModeloTrabalho)
		}
		if len(filter.AcessibilidadePCD) == 1 {
			db = db.Where("acessibilidade_pcd = ?", filter.AcessibilidadePCD[0])
		} else if len(filter.AcessibilidadePCD) > 1 {
			db = db.Where("acessibilidade_pcd IN ?", filter.AcessibilidadePCD)
		}
		if len(filter.Bairro) > 0 {
			conditions := make([]string, len(filter.Bairro))
			args := make([]interface{}, len(filter.Bairro))
			for i, b := range filter.Bairro {
				conditions[i] = "bairro ILIKE ?"
				args[i] = "%" + b + "%"
			}
			db = db.Where(strings.Join(conditions, " OR "), args...)
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	result := findDB.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) ListPublicActive(ctx context.Context, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).
		Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas públicas ativas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM emp_vagas_tipos_pcd WHERE id_vaga = ?", vagaID).Error; err != nil {
			return fmt.Errorf("erro ao remover tipos PCD: %w", err)
		}

		for _, tipoPCDID := range tiposPCDIDs {
			if err := tx.Exec("INSERT INTO emp_vagas_tipos_pcd (id_vaga, id_tipo_pcd) VALUES (?, ?)", vagaID, tipoPCDID).Error; err != nil {
				return fmt.Errorf("erro ao inserir tipo PCD: %w", err)
			}
		}

		return nil
	})
}

func (r *VagaRepository) UpdateZonas(ctx context.Context, vagaID uuid.UUID, zonaIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM emp_vagas_zonas WHERE id_vaga = ?", vagaID).Error; err != nil {
			return fmt.Errorf("erro ao remover zonas: %w", err)
		}

		for _, zonaID := range zonaIDs {
			if err := tx.Exec("INSERT INTO emp_vagas_zonas (id_vaga, id_zona) VALUES (?, ?)", vagaID, zonaID).Error; err != nil {
				return fmt.Errorf("erro ao inserir zona: %w", err)
			}
		}

		return nil
	})
}

func (r *VagaRepository) UpdateIdiomasRequisito(ctx context.Context, vagaID uuid.UUID, requisitos []empregabilidade.VagaIdiomaRequisito) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM emp_vagas_idiomas_requisitos WHERE id_vaga = ?", vagaID).Error; err != nil {
			return fmt.Errorf("erro ao remover requisitos de idioma: %w", err)
		}

		for _, requisito := range requisitos {
			if err := tx.Exec(
				"INSERT INTO emp_vagas_idiomas_requisitos (id_vaga, id_idioma, id_nivel_minimo) VALUES (?, ?, ?)",
				vagaID, requisito.IDIdioma, requisito.IDNivelMinimo,
			).Error; err != nil {
				return fmt.Errorf("erro ao inserir requisito de idioma: %w", err)
			}
		}

		return nil
	})
}

func (r *VagaRepository) ListPublic(ctx context.Context, filter empregabilidade.VagaPublicFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		switch filter.Status {
		case string(empregabilidade.StatusVagaPublicadoAtivo):
			db = db.Where("emp_vagas.status = ? AND (emp_vagas.data_limite IS NULL OR emp_vagas.data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)
		case string(empregabilidade.StatusVagaPublicadoExpirado):
			db = db.Where("emp_vagas.status = ? AND emp_vagas.data_limite IS NOT NULL AND emp_vagas.data_limite <= NOW()", empregabilidade.StatusVagaPublicadoAtivo)
		case string(empregabilidade.StatusVagaCongelada):
			db = db.Where("emp_vagas.status = ?", empregabilidade.StatusVagaCongelada)
		case string(empregabilidade.StatusVagaDescontinuada):
			db = db.Where("emp_vagas.status = ?", empregabilidade.StatusVagaDescontinuada)
		default:
			db = db.Where("emp_vagas.status IN ?", []string{
				string(empregabilidade.StatusVagaPublicadoAtivo),
				string(empregabilidade.StatusVagaCongelada),
				string(empregabilidade.StatusVagaDescontinuada),
			})
		}

		switch filter.DataPublicacao {
		case empregabilidade.DataPublicacaoHoje:
			hoje := time.Now().Truncate(24 * time.Hour)
			db = db.Where("emp_vagas.created_at >= ?", hoje)
		case empregabilidade.DataPublicacaoUltimaSemana:
			db = db.Where("emp_vagas.created_at >= ?", time.Now().AddDate(0, 0, -7))
		case empregabilidade.DataPublicacaoUltimoMes:
			db = db.Where("emp_vagas.created_at >= ?", time.Now().AddDate(0, 0, -30))
		}

		if len(filter.IDRegimeContratacao) == 1 {
			db = db.Where("emp_vagas.id_regime_contratacao = ?", filter.IDRegimeContratacao[0])
		} else if len(filter.IDRegimeContratacao) > 1 {
			db = db.Where("emp_vagas.id_regime_contratacao IN ?", filter.IDRegimeContratacao)
		}

		if len(filter.IDModeloTrabalho) == 1 {
			db = db.Where("emp_vagas.id_modelo_trabalho = ?", filter.IDModeloTrabalho[0])
		} else if len(filter.IDModeloTrabalho) > 1 {
			db = db.Where("emp_vagas.id_modelo_trabalho IN ?", filter.IDModeloTrabalho)
		}

		if len(filter.AcessibilidadePCD) == 1 {
			db = db.Where("emp_vagas.acessibilidade_pcd = ?", filter.AcessibilidadePCD[0])
		} else if len(filter.AcessibilidadePCD) > 1 {
			db = db.Where("emp_vagas.acessibilidade_pcd IN ?", filter.AcessibilidadePCD)
		}

		if len(filter.Bairro) > 0 {
			conditions := make([]string, len(filter.Bairro))
			args := make([]interface{}, len(filter.Bairro))
			for i, b := range filter.Bairro {
				conditions[i] = "emp_vagas.bairro ILIKE ?"
				args[i] = "%" + b + "%"
			}
			db = db.Where(strings.Join(conditions, " OR "), args...)
		}

		if len(filter.Contratante) > 0 {
			var cnpjs []string
			var nameConditions []string
			var nameArgs []interface{}
			for _, c := range filter.Contratante {
				digits := nonDigit.ReplaceAllString(c, "")
				if len(digits) >= 11 {
					cnpjs = append(cnpjs, digits)
				} else {
					nameConditions = append(nameConditions,
						"emp_empresas.razao_social ILIKE ? OR emp_empresas.nome_fantasia ILIKE ?")
					nameArgs = append(nameArgs, "%"+c+"%", "%"+c+"%")
				}
			}
			if len(nameConditions) > 0 {
				db = db.Joins("JOIN emp_empresas ON emp_empresas.cnpj = emp_vagas.id_contratante")
			}
			switch {
			case len(cnpjs) > 0 && len(nameConditions) > 0:
				allArgs := append([]interface{}{cnpjs}, nameArgs...)
				db = db.Where(
					"emp_vagas.id_contratante IN ? OR ("+strings.Join(nameConditions, " OR ")+")",
					allArgs...,
				)
			case len(cnpjs) > 0:
				if len(cnpjs) == 1 {
					db = db.Where("emp_vagas.id_contratante = ?", cnpjs[0])
				} else {
					db = db.Where("emp_vagas.id_contratante IN ?", cnpjs)
				}
			default:
				db = db.Where(strings.Join(nameConditions, " OR "), nameArgs...)
			}
		}

		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas públicas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	result := findDB.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("emp_vagas.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas públicas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) ListByContratante(ctx context.Context, cnpj string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).Where("id_contratante = ?", cnpj)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas por contratante: %w", err)
	}

	result := db.
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por contratante: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) GetByIDPrefix(ctx context.Context, idPrefix string) (*empregabilidade.Vaga, error) {
	var entity empregabilidade.Vaga
	result := r.db.WithContext(ctx).
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Preload("TiposPCD").
		Preload("Zonas").
		Preload("IdiomasRequisito").
		Preload("Etapas", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Preload("InformacoesComplementares").
		First(&entity, "replace(id::text, '-', '') LIKE ?", idPrefix+"%")
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar vaga por slug: %w", result.Error)
	}
	return &entity, nil
}

func (r *VagaRepository) ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).Where("id_orgao_parceiro = ?", orgaoID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas por órgão parceiro: %w", err)
	}

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por órgão parceiro: %w", result.Error)
	}

	return entities, int(total), nil
}
