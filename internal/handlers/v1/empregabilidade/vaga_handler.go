package empregabilidade

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseCSVParam(param string) []string {
	if param == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var result []string
	for _, s := range strings.Split(param, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, exists := seen[s]; !exists {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

type VagaHandler struct {
	service *services.VagaService
}

func NewVagaHandler(service *services.VagaService) *VagaHandler {
	return &VagaHandler{service: service}
}

// @Summary      Criar vaga
// @Description  Cria uma nova vaga
// @Description  PcD: opcao_pcd (aberta, preferencial ou exclusiva; ausente = aberta) com os tipos em tipos_pcd; exclusiva sem nenhum tipo retorna 400 "Selecione ao menos um tipo de deficiência.". O bloqueio de candidatura PcD é a chave "pcd" de criterios_bloqueio. acessibilidade_pcd está descontinuado: é preenchido a partir de opcao_pcd e só vale como entrada quando opcao_pcd não é enviada.
// @Description  Experiência em carteira: experiencia_carteira (qualquer_tempo ou minimo_6_meses; ausente, null ou "" = sem exigência). O bloqueio de candidatura é a chave "experiencia_carteira" de criterios_bloqueio; ligado sem experiencia_carteira retorna 400.
// @Description  Idade: Maiores de 18 é idade_minima (como hoje); Jovem Aprendiz é jovem_aprendiz (booleano, faixa fixa de 14 a 24 anos, sem número livre; ausente = false). O bloqueio de candidatura é a chave "jovem_aprendiz" de criterios_bloqueio; ligado sem jovem_aprendiz true retorna 400. Um critério de idade por vez: idade_minima junto de jovem_aprendiz true com a chave "jovem_aprendiz" ligada retorna 400.
// @Description  Habilidades e competências: areas_atuacao (lista de {id_area_atuacao, conhecimento_ids}, cada área uma vez, cada conhecimento da área em que foi marcado) e comportamento_atitudes_ids. Opcionais no rascunho. Área repetida, área, conhecimento ou comportamento fora do catálogo, conhecimento de outra área e lista acima do limite (50 áreas, 300 conhecimentos por área, 100 comportamentos) retornam 400 com o erro do campo em "fields" (areas_atuacao, areas_atuacao.conhecimento_ids, comportamento_atitudes_ids). A resposta traz as listas com os nomes (nome, conhecimentos, comportamentos_atitudes), que são ignorados na entrada: para duplicar uma vaga, areas_atuacao e comportamento_atitudes_ids devolvidos pelo GET podem ser enviados como estão no POST do rascunho (sem o id da vaga).
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Vaga  true  "Dados da vaga"
// @Success      201      {object}  empregabilidade.Vaga
// @Failure      400      {object}  response.ErrorWithFields
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas [post]
// @Router       /api/v1/empregabilidade/vagas/draft [post]
func (h *VagaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Vaga
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	allowedOrgaos := middlewares.GetVagaAllowedOrgaos(c)
	if allowedOrgaos != nil && !containsString(allowedOrgaos, entity.IDOrgaoParceiro) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para criar vagas neste órgão parceiro"})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar vagas
// @Description  Retorna lista paginada de vagas com filtros opcionais
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        page                    query     int     false  "Número da página (default: 1)"
// @Param        pageSize                query     int     false  "Tamanho da página (default: 10, máx: 100)"
// @Param        status                  query     string  false  "Filtrar por status (em_edicao, em_aprovacao, publicado_ativo, publicado_expirado, vaga_congelada, vaga_descontinuada)"
// @Param        contratante             query     string  false  "Filtrar por CNPJ do contratante"
// @Param        orgao_parceiro_id       query     string  false  "Filtrar por ID do órgão parceiro (ignorado se middleware já restringiu)"
// @Param        search                  query     string  false  "Busca parcial no título da vaga"
// @Param        data_publicacao         query     string  false  "Filtrar por data de publicação (hoje, ultima_semana, ultimo_mes)"
// @Param        id_regime_contratacao   query     string  false  "UUID(s) do regime de contratação (CSV, ex: uuid1,uuid2)"
// @Param        id_modelo_trabalho      query     string  false  "UUID(s) do modelo de trabalho (CSV, ex: uuid1,uuid2)"
// @Param        acessibilidade_pcd      query     string  false  "Acessibilidade PCD (CSV, ex: para_pcd,exclusivo_pcd)"
// @Param        bairro                  query     string  false  "Bairro(s) — busca parcial (CSV, ex: Centro,Tijuca)"
// @Success      200                     {object}  map[string]interface{}
// @Failure      400                     {object}  map[string]string
// @Failure      500                     {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas [get]
func (h *VagaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	dataPublicacao := empregabilidade.DataPublicacaoRange(c.Query("data_publicacao"))
	if dataPublicacao != "" && !dataPublicacao.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_publicacao inválido: deve ser hoje, ultima_semana ou ultimo_mes"})
		return
	}

	acessibilidadePCDValues := parseCSVParam(c.Query("acessibilidade_pcd"))
	for _, v := range acessibilidadePCDValues {
		if !empregabilidade.AcessibilidadePCD(v).IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "acessibilidade_pcd inválido: deve ser para_pcd, preferencial_pcd ou exclusivo_pcd"})
			return
		}
	}

	orgaoParceiroIDs := middlewares.GetVagaOrgaoParceiroIDs(c)

	if orgaoParceiroIDs != nil && len(orgaoParceiroIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"data": []*empregabilidade.Vaga{},
			"meta": gin.H{"page": page, "page_size": pageSize, "total": 0},
		})
		return
	}

	filter := empregabilidade.VagaFilter{
		Status:              c.Query("status"),
		Contratante:         c.Query("contratante"),
		Search:              c.Query("search"),
		DataPublicacao:      dataPublicacao,
		IDRegimeContratacao: parseCSVParam(c.Query("id_regime_contratacao")),
		IDModeloTrabalho:    parseCSVParam(c.Query("id_modelo_trabalho")),
		AcessibilidadePCD:   acessibilidadePCDValues,
		Bairro:              parseCSVParam(c.Query("bairro")),
	}

	switch {
	case len(orgaoParceiroIDs) == 1:
		filter.OrgaoParceiroID = orgaoParceiroIDs[0]
	case len(orgaoParceiroIDs) > 1:
		filter.OrgaoParceiroIDs = orgaoParceiroIDs
	default:
		filter.OrgaoParceiroID = c.Query("orgao_parceiro_id")
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Buscar vaga
// @Description  Retorna uma vaga pelo ID, com as habilidades e competências cadastradas: areas_atuacao (cada área com id_area_atuacao, nome, conhecimento_ids e conhecimentos), comportamento_atitudes_ids e comportamentos_atitudes. Vaga sem registro devolve listas vazias. As rotas públicas não trazem esses campos.
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  empregabilidade.Vaga
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [get]
func (h *VagaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByIDGestao(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar vaga
// @Description  Atualiza uma vaga existente
// @Description  PcD: opcao_pcd (aberta, preferencial ou exclusiva) vale sobre acessibilidade_pcd (descontinuado); sem nenhum dos dois a vaga fica aberta. tipos_pcd ausente mantém os tipos gravados e [] remove todos; exclusiva sem nenhum tipo retorna 400 "Selecione ao menos um tipo de deficiência.".
// @Description  Experiência em carteira: experiencia_carteira (qualquer_tempo ou minimo_6_meses) ausente ou null mantém o valor gravado e "" remove a exigência; com a chave "experiencia_carteira" ligada em criterios_bloqueio (enviada ou mantida), a vaga sem experiencia_carteira retorna 400.
// @Description  Idade: jovem_aprendiz (Jovem Aprendiz, faixa fixa de 14 a 24 anos) ausente ou null mantém o valor gravado; com a chave "jovem_aprendiz" ligada em criterios_bloqueio (enviada ou mantida), a vaga sem jovem_aprendiz true retorna 400. Maiores de 18 segue em idade_minima; um critério de idade por vez: idade_minima junto de jovem_aprendiz true com a chave "jovem_aprendiz" ligada (enviados ou mantidos) retorna 400.
// @Description  Habilidades e competências: areas_atuacao e comportamento_atitudes_ids ausentes ou null mantêm o gravado; a lista enviada substitui o conjunto inteiro (apaga o que saiu, insere o que entrou; [] apaga tudo). Mesmas regras e erros por campo do POST. A resposta é a vaga relida, com as listas e os nomes.
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string                true  "ID da vaga"
// @Param        request  body      empregabilidade.Vaga  true  "Dados da vaga"
// @Success      200      {object}  empregabilidade.Vaga
// @Failure      400      {object}  response.ErrorWithFields
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [put]
func (h *VagaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err)
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	// A autorização por órgão/secretaria fica em VagaOwnershipCheck (middleware,
	// só ativo em dev/test). Já a regra de "quem pode editar vaga publicada"
	// depende do status da vaga — algo que o gateway Heimdall não conhece —, então
	// precisa ser aplicada aqui no handler, em qualquer ambiente.
	canEditPublished := middlewares.IsAdmin(c) ||
		middlewares.HasRole(c, "go:empregabilidade:admin") ||
		middlewares.HasRole(c, "go:empregabilidade:editor_sem_curadoria")
	if existing.Status.IsPublished() && !canEditPublished {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas administradores, empregabilidade:admin ou editor_sem_curadoria podem editar vagas publicadas"})
		return
	}

	var entity empregabilidade.Vaga
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Mesma lista do Create (VagaOrgaoInjector): admin / emp:admin sem restrição;
	// editor só nas secretarias/cd_ua dele. Campo vazio não troca o órgão.
	if entity.IDOrgaoParceiro != "" {
		allowedOrgaos := middlewares.GetVagaAllowedOrgaos(c)
		if allowedOrgaos != nil && !containsString(allowedOrgaos, entity.IDOrgaoParceiro) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para atribuir esta vaga a este órgão parceiro"})
			return
		}
	}

	entity.ID = id

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		response.RespondError(c, err)
		return
	}

	updated, err := h.service.GetByIDGestao(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary      Excluir vaga
// @Description  Remove uma vaga pelo ID
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [delete]
func (h *VagaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga excluída com sucesso"})
}

// @Summary      Retornar vaga para rascunho
// @Description  Retorna uma vaga em aprovação para o estado de edição (rascunho)
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/send-to-draft [put]
func (h *VagaHandler) SendToDraft(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.SendToDraft(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga retornada para rascunho com sucesso"})
}

// @Summary      Enviar vaga para aprovação
// @Description  Envia uma vaga em edição para o fluxo de aprovação. Exige ao menos uma área de atuação, ao menos um conhecimento em cada área e ao menos um comportamento ou atitude: sem eles retorna 400 com o erro de cada campo em "fields" (areas_atuacao, areas_atuacao.conhecimento_ids, comportamento_atitudes_ids).
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  response.ErrorWithFields
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/send-to-approval [put]
func (h *VagaHandler) SendToApproval(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.SendToApproval(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga enviada para aprovação com sucesso"})
}

// @Summary      Publicar vaga
// @Description  Publica uma vaga em edição ou em aprovação. Exige ao menos uma área de atuação, ao menos um conhecimento em cada área e ao menos um comportamento ou atitude: sem eles retorna 400 com o erro de cada campo em "fields" (areas_atuacao, areas_atuacao.conhecimento_ids, comportamento_atitudes_ids).
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  response.ErrorWithFields
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/publish [put]
func (h *VagaHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Publish(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga publicada com sucesso"})
}

type UpdateTiposPCDRequest struct {
	TiposPCDIDs []uuid.UUID `json:"tipos_pcd_ids"`
}

// @Summary      Atualizar tipos PCD da vaga
// @Description  Atualiza os tipos de PCD aceitos na vaga. Vaga exclusiva (opcao_pcd = exclusiva) não pode ficar sem tipo: lista vazia retorna 400 "Selecione ao menos um tipo de deficiência.".
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "ID da vaga"
// @Param        request  body      UpdateTiposPCDRequest  true  "IDs dos tipos PCD"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/tipos-pcd [put]
func (h *VagaHandler) UpdateTiposPCD(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateTiposPCDRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateTiposPCD(c.Request.Context(), id, request.TiposPCDIDs); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipos PCD atualizados com sucesso"})
}

type UpdateZonasRequest struct {
	ZonaIDs []uuid.UUID `json:"zona_ids"`
}

// @Summary      Atualizar zonas de elegibilidade da vaga
// @Description  Atualiza as zonas geográficas de elegibilidade da vaga
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string              true  "ID da vaga"
// @Param        request  body      UpdateZonasRequest  true  "IDs das zonas"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/zonas [put]
func (h *VagaHandler) UpdateZonas(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateZonasRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateZonas(c.Request.Context(), id, request.ZonaIDs); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zonas atualizadas com sucesso"})
}

type UpdateIdiomasRequisitoRequest struct {
	Requisitos []empregabilidade.VagaIdiomaRequisito `json:"requisitos"`
}

// @Summary      Atualizar requisitos de idioma da vaga
// @Description  Atualiza os idiomas e níveis mínimos exigidos como critério de elegibilidade da vaga
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string                         true  "ID da vaga"
// @Param        request  body      UpdateIdiomasRequisitoRequest  true  "Requisitos de idioma"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/idiomas-requisito [put]
func (h *VagaHandler) UpdateIdiomasRequisito(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateIdiomasRequisitoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	for i := range request.Requisitos {
		request.Requisitos[i].IDVaga = id
	}

	if err := h.service.UpdateIdiomasRequisito(c.Request.Context(), id, request.Requisitos); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Requisitos de idioma atualizados com sucesso"})
}

// @Summary      Congelar vaga
// @Description  Congela uma vaga publicada e atualiza o status de todas as candidaturas vinculadas para vaga_congelada
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/freeze [put]
func (h *VagaHandler) Freeze(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.FreezeVaga(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga congelada com sucesso"})
}

// @Summary      Descongelar vaga
// @Description  Descongela uma vaga congelada e restaura o status anterior das candidaturas
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/unfreeze [put]
func (h *VagaHandler) Unfreeze(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.UnfreezeVaga(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga descongelada com sucesso"})
}

// @Summary      Descontinuar vaga
// @Description  Descontinua uma vaga publicada ou congelada e atualiza o status de todas as candidaturas para vaga_descontinuada
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/discontinue [put]
func (h *VagaHandler) Discontinue(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.DiscontinueVaga(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga descontinuada com sucesso"})
}

// @Summary      Reativar vaga
// @Description  Reativa uma vaga descontinuada e restaura o status anterior das candidaturas
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/reactivate [put]
func (h *VagaHandler) Reactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.ReactivateVaga(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga reativada com sucesso"})
}

// ==================== Public Endpoints ====================

// @Summary      Listar vagas públicas
// @Description  Retorna lista paginada de vagas publicadas. Sem filtro retorna todos os status públicos (publicado_ativo, publicado_expirado, vaga_congelada, vaga_descontinuada). Use ?status= para filtrar por um status específico.
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        page                    query     int     false  "Número da página (default: 1)"
// @Param        pageSize                query     int     false  "Tamanho da página (default: 10)"
// @Param        status                  query     string  false  "Filtrar por status (publicado_ativo, publicado_expirado, vaga_congelada, vaga_descontinuada)"
// @Param        data_publicacao         query     string  false  "Filtrar por data de publicação (hoje, ultima_semana, ultimo_mes)"
// @Param        id_regime_contratacao   query     string  false  "UUID(s) do regime de contratação (CSV, ex: uuid1,uuid2)"
// @Param        id_modelo_trabalho      query     string  false  "UUID(s) do modelo de trabalho (CSV, ex: uuid1,uuid2)"
// @Param        contratante             query     string  false  "Nome(s) ou CNPJ(s) do contratante (CSV, ex: TechRio,12345678000100)"
// @Param        acessibilidade_pcd      query     string  false  "Acessibilidade PCD (CSV, ex: para_pcd,exclusivo_pcd)"
// @Param        bairro                  query     string  false  "Bairro(s) — busca parcial (CSV, ex: Centro,Tijuca)"
// @Success      200                     {object}  empregabilidade.VagaPublicListResponse
// @Failure      400                     {object}  map[string]string
// @Failure      500                     {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas [get]
func (h *VagaHandler) PublicList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	status := c.Query("status")
	if status != "" && !empregabilidade.StatusVaga(status).IsPublished() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status inválido: deve ser publicado_ativo, publicado_expirado, vaga_congelada ou vaga_descontinuada"})
		return
	}

	dataPublicacao := empregabilidade.DataPublicacaoRange(c.Query("data_publicacao"))
	if dataPublicacao != "" && !dataPublicacao.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_publicacao inválido: deve ser hoje, ultima_semana ou ultimo_mes"})
		return
	}

	acessibilidadePCDValues := parseCSVParam(c.Query("acessibilidade_pcd"))
	for _, v := range acessibilidadePCDValues {
		if !empregabilidade.AcessibilidadePCD(v).IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "acessibilidade_pcd inválido: deve ser para_pcd, preferencial_pcd ou exclusivo_pcd"})
			return
		}
	}

	filter := empregabilidade.VagaPublicFilter{
		Status:              status,
		DataPublicacao:      dataPublicacao,
		IDRegimeContratacao: parseCSVParam(c.Query("id_regime_contratacao")),
		IDModeloTrabalho:    parseCSVParam(c.Query("id_modelo_trabalho")),
		Contratante:         parseCSVParam(c.Query("contratante")),
		AcessibilidadePCD:   acessibilidadePCDValues,
		Bairro:              parseCSVParam(c.Query("bairro")),
	}

	entities, total, err := h.service.ListPublic(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Buscar vaga pública por slug
// @Description  Retorna uma vaga publicada pelo slug (apenas se estiver ativa). Slug no formato "{titulo-slugificado}-{12-chars-hex-do-uuid}". Se o slug for histórico (título mudou), retorna 301 redirect para o slug atual.
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        slug  path      string  true  "Slug da vaga (ex: analista-financeiro-jr-b06235c80d2a)"
// @Success      200   {object}  empregabilidade.Vaga
// @Success      301   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas/slug/{slug} [get]
func (h *VagaHandler) PublicGetBySlug(c *gin.Context) {
	vagaSlug := c.Param("slug")

	entity, err := h.service.GetBySlug(c.Request.Context(), vagaSlug)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	if entity == nil || !entity.Status.IsPublished() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	if entity.Slug != vagaSlug {
		originalPath := c.GetHeader("X-Envoy-Original-Path")
		if originalPath == "" {
			originalPath = c.Request.RequestURI
		}
		base := originalPath[:strings.LastIndex(originalPath, "/")]
		c.Redirect(http.StatusMovedPermanently, base+"/"+entity.Slug)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Buscar vaga pública
// @Description  Retorna uma vaga publicada pelo ID (publicado_ativo, publicado_expirado, vaga_congelada, vaga_descontinuada)
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  empregabilidade.Vaga
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas/{id} [get]
func (h *VagaHandler) PublicGetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	if entity == nil || !entity.Status.IsPublished() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}
