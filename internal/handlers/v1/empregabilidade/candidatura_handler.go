package empregabilidade

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type CandidaturaHandler struct {
	service *services.CandidaturaService
}

func NewCandidaturaHandler(service *services.CandidaturaService) *CandidaturaHandler {
	return &CandidaturaHandler{service: service}
}

// @Summary      Criar candidatura
// @Description  Cria uma nova candidatura para uma vaga
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Candidatura  true  "Dados da candidatura"
// @Success      201      {object}  empregabilidade.Candidatura
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas [post]
func (h *CandidaturaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Candidatura
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// A origem é sempre decidida aqui, nunca pelo corpo da requisição.
	entity.AssociadoPorCPF = nil
	if (middlewares.IsAdmin(c) || middlewares.HasRole(c, "go:empregabilidade:admin")) && entity.CPF != "" {
		// Admin criando candidatura para outro usuário — usa CPF do body, sem sobrescrever nome/email
		origem := empregabilidade.OrigemCandidaturaSecretaria
		entity.Origem = &origem
	} else {
		origem := empregabilidade.OrigemCandidaturaPortalCidadao
		entity.Origem = &origem
		cpf := middlewares.GetUserCPF(c)
		if cpf == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "CPF não encontrado no token de autenticação"})
			return
		}
		entity.CPF = cpf
		if nome := middlewares.GetUserName(c); nome != "" {
			entity.Nome = &nome
		}
		if email := middlewares.GetUserEmail(c); email != "" {
			entity.Email = &email
		}
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar candidaturas (admin)
// @Description  Retorna lista paginada de candidaturas com filtros e busca universal. Restrito a administradores e editores de empregabilidade; com o RBAC ligado, o editor só vê candidaturas de vagas dos órgãos do seu escopo.
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10)"
// @Param        cpf       query     string  false  "Filtrar por CPF"
// @Param        vagaId    query     string  false  "Filtrar por ID da vaga"
// @Param        status    query     string  false  "Filtrar por status"
// @Param        search    query     string  false  "Busca parcial por CPF, nome ou email"
// @Param        etapa_id  query     string  false  "Filtrar por ID da etapa atual"
// @Param        genero               query  []string  false  "Gêneros autodeclarados (repetível ou separado por vírgula)"  collectionFormat(multi)
// @Param        idade_min            query  int       false  "Idade mínima"
// @Param        idade_max            query  int       false  "Idade máxima"
// @Param        tempo_experiencia    query  string    false  "Soma das experiências"  Enums(sem_experiencia, ate_2_anos, 2_a_4_anos, 4_a_6_anos, 6_a_8_anos, 8_a_10_anos, mais_de_10_anos)
// @Param        experiencia_carteira query  string    false  "Experiência comprovada em carteira"  Enums(sim, 6_meses)
// @Param        escolaridade         query  string    false  "Escolaridade do cadastro (exata)"
// @Param        id_idioma            query  string    false  "Idioma do currículo"
// @Param        id_nivel_idioma      query  string    false  "Nível mínimo do idioma (exige id_idioma)"
// @Param        bairro               query  string    false  "Bairro do cadastro"
// @Param        pcd                  query  string    false  "Pessoa com deficiência"  Enums(sim, nao)
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]string
// @Failure      403       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas [get]
func (h *CandidaturaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	candidato, err := lerFiltroCandidato(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	filter := empregabilidade.CandidaturaFilter{
		Status:           c.Query("status"),
		CPF:              c.Query("cpf"),
		Search:           c.Query("search"),
		Candidato:        candidato,
		OrgaoParceiroIDs: middlewares.GetVagaOrgaoParceiroIDs(c),
	}

	if raw := c.Query("vagaId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
			return
		}
		filter.VagaID = &id
	}

	if raw := c.Query("etapa_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da etapa inválido"})
			return
		}
		filter.EtapaID = &id
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	resumoFilter := filter
	resumoFilter.Status = ""
	resumo, err := h.service.CountByStatus(c.Request.Context(), resumoFilter)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	forceRefresh := c.Query("refresh") == "true" || c.Query("sync") == "true"
	h.service.EnrichRespostasWithTituloMultiple(entities)
	h.service.EnrichMultipleWithPersonalInfo(c.Request.Context(), entities, forceRefresh)
	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
		"resumo": resumo,
	})
}

// @Summary      Listar candidaturas por CPF
// @Description  Retorna lista paginada de candidaturas de um usuário pelo CPF. Usuários só podem acessar o próprio CPF; admins podem acessar qualquer CPF; com o RBAC ligado, o editor de empregabilidade que consulta outro CPF só vê candidaturas de vagas dos órgãos do seu escopo. O campo atualizacoes conta as candidaturas do CPF com mudança de status ou de etapa (registrada no histórico) que o candidato ainda não viu (ignora os filtros da listagem).
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        cpf       path      string  true   "CPF do usuário"
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10)"
// @Param        vagaId    query     string  false  "Filtrar por ID da vaga"
// @Param        status    query     string  false  "Filtrar por status"
// @Param        etapa_id  query     string  false  "Filtrar por ID da etapa atual"
// @Success      200       {object}  map[string]interface{}
// @Failure      401       {object}  map[string]string
// @Failure      403       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/usuario/{cpf} [get]
func (h *CandidaturaHandler) ListByCPF(c *gin.Context) {
	cpf := c.Param("cpf")

	if !middlewares.IsAdmin(c) && !middlewares.IsEmpregabilidadeRole(c) {
		userCPF := middlewares.GetUserCPF(c)
		if userCPF == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
			return
		}
		if userCPF != cpf {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := empregabilidade.CandidaturaFilter{
		CPF:    cpf,
		Status: c.Query("status"),
	}
	// O próprio cidadão vê todas as suas candidaturas; quem consulta o CPF de
	// outra pessoa fica restrito às vagas do seu escopo de órgão (DEV-517).
	if middlewares.GetUserCPF(c) != cpf {
		filter.OrgaoParceiroIDs = middlewares.GetVagaOrgaoParceiroIDs(c)
	}

	if raw := c.Query("vagaId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
			return
		}
		filter.VagaID = &id
	}

	if raw := c.Query("etapa_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da etapa inválido"})
			return
		}
		filter.EtapaID = &id
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	resumoFilter := filter
	resumoFilter.Status = ""
	resumo, err := h.service.CountByStatus(c.Request.Context(), resumoFilter)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	atualizacoes, err := h.service.CountAtualizacoesNaoVistas(c.Request.Context(), cpf)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	forceRefresh := c.Query("refresh") == "true" || c.Query("sync") == "true"
	h.service.EnrichRespostasWithTituloMultiple(entities)
	h.service.EnrichMultipleWithPersonalInfo(c.Request.Context(), entities, forceRefresh)
	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
		"resumo":       resumo,
		"atualizacoes": atualizacoes,
	})
}

// @Summary      Marcar atualizações das candidaturas como vistas
// @Description  Marca como vistas as mudanças de status e de etapa das candidaturas do CPF, zerando o campo atualizacoes da listagem por CPF. Só o próprio candidato pode marcar.
// @Tags         empregabilidade-candidaturas
// @Param        cpf  path  string  true  "CPF do usuário"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/usuario/{cpf}/visualizar [put]
func (h *CandidaturaHandler) MarcarAtualizacoesVistas(c *gin.Context) {
	cpf := c.Param("cpf")

	// Ver as atualizações é do candidato: nem admin marca no lugar dele.
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}
	if userCPF != cpf {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	if err := h.service.MarcarAtualizacoesComoVistas(c.Request.Context(), cpf); err != nil {
		response.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Buscar candidatura
// @Description  Retorna uma candidatura pelo ID
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  empregabilidade.Candidatura
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [get]
func (h *CandidaturaHandler) GetByID(c *gin.Context) {
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

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		return
	}

	if !middlewares.HasRole(c, "go:empregabilidade:admin") && !requireOwnership(c, entity.CPF) {
		return
	}

	forceRefresh := c.Query("refresh") == "true" || c.Query("sync") == "true"
	h.service.EnrichRespostasWithTitulo(entity)
	h.service.EnrichWithPersonalInfo(c.Request.Context(), entity, forceRefresh)
	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar candidatura
// @Description  Atualiza uma candidatura existente
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "ID da candidatura"
// @Param        request  body      empregabilidade.Candidatura  true  "Dados da candidatura"
// @Success      200      {object}  empregabilidade.Candidatura
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [put]
func (h *CandidaturaHandler) Update(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		return
	}

	if !middlewares.HasRole(c, "go:empregabilidade:admin") && !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.Candidatura
	if err := c.ShouldBindJSON(&entity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	entity.ID = id

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir candidatura
// @Description  Remove uma candidatura pelo ID
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [delete]
func (h *CandidaturaHandler) Delete(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		return
	}

	if !middlewares.HasRole(c, "go:empregabilidade:admin") && !requireOwnership(c, existing.CPF) {
		return
	}

	// Histórico (DEV-532): o próprio candidato cancela; qualquer outro é o operador.
	ator := empregabilidade.AtorOperador
	if userCPF := middlewares.GetUserCPF(c); userCPF != "" && userCPF == existing.CPF {
		ator = empregabilidade.AtorCandidato
	}
	ctx := empregabilidade.ComOrigemEvento(c.Request.Context(), empregabilidade.OrigemEventoCandidatura{Ator: ator})

	if err := h.service.Delete(ctx, id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura excluída com sucesso"})
}

type UpdateStatusRequest struct {
	Status empregabilidade.StatusCandidatura `json:"status"`
}

// @Summary      Atualizar status da candidatura
// @Description  Atualiza o status de uma candidatura
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string               true  "ID da candidatura"
// @Param        request  body      UpdateStatusRequest  true  "Novo status"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/status [put]
func (h *CandidaturaHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !h.exigirCandidaturaNoEscopo(c, id) {
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, request.Status); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// @Summary      Aprovar candidatura
// @Description  Aprova uma candidatura
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/approve [put]
func (h *CandidaturaHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if !h.exigirCandidaturaNoEscopo(c, id) {
		return
	}

	if err := h.service.Approve(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura aprovada com sucesso"})
}

// @Summary      Reprovar candidatura
// @Description  Reprova uma candidatura
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/reject [put]
func (h *CandidaturaHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if !h.exigirCandidaturaNoEscopo(c, id) {
		return
	}

	if err := h.service.Reject(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura reprovada com sucesso"})
}

type BulkUpdateStatusRequest struct {
	CPFs   []string                          `json:"cpfs"`
	VagaID uuid.UUID                         `json:"vaga_id"`
	Status empregabilidade.StatusCandidatura `json:"status"`
}

// @Summary      Atualizar status de candidaturas em lote
// @Description  Atualiza o status de múltiplas candidaturas de uma vaga identificadas por lista de CPFs
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        request  body      BulkUpdateStatusRequest  true  "Lista de CPFs, ID da vaga e novo status"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/bulk-status [put]
func (h *CandidaturaHandler) BulkUpdateStatus(c *gin.Context) {
	var request BulkUpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !h.exigirVagaNoEscopo(c, request.VagaID) {
		return
	}

	result, err := h.service.BulkUpdateStatus(c.Request.Context(), request.VagaID, request.CPFs, request.Status)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":     result.Updated,
		"failed_cpfs": result.FailedCPFs,
	})
}

type BulkUpdateEtapaRequest struct {
	CPFs    []string  `json:"cpfs"`
	VagaID  uuid.UUID `json:"vaga_id"`
	EtapaID uuid.UUID `json:"id_etapa"`
}

// @Summary      Atualizar etapa de candidaturas em lote
// @Description  Atualiza a etapa atual de múltiplas candidaturas de uma vaga. Todos os candidatos devem estar na mesma etapa atual.
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        request  body      BulkUpdateEtapaRequest  true  "Lista de CPFs, ID da vaga e nova etapa"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/bulk-etapa [put]
func (h *CandidaturaHandler) BulkUpdateEtapa(c *gin.Context) {
	var request BulkUpdateEtapaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !h.exigirVagaNoEscopo(c, request.VagaID) {
		return
	}

	result, err := h.service.BulkUpdateEtapa(c.Request.Context(), request.VagaID, request.CPFs, request.EtapaID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":     result.Updated,
		"failed_cpfs": result.FailedCPFs,
	})
}

type UpdateEtapaRequest struct {
	EtapaID uuid.UUID `json:"id_etapa_atual"`
}

// @Summary      Avançar etapa da candidatura
// @Description  Atualiza a etapa atual de uma candidatura
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string             true  "ID da candidatura"
// @Param        request  body      UpdateEtapaRequest true  "Nova etapa"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/etapa [put]
func (h *CandidaturaHandler) UpdateEtapa(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateEtapaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !h.exigirCandidaturaNoEscopo(c, id) {
		return
	}

	if err := h.service.UpdateEtapa(c.Request.Context(), id, request.EtapaID); err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Etapa atualizada com sucesso"})
}

// orgaoNoEscopo reporta se o órgão parceiro está no escopo que VagaListFilter
// injetou. Sem escopo injetado (admin, go:empregabilidade:admin ou
// RBAC_ENABLED=false) não há restrição.
func orgaoNoEscopo(c *gin.Context, orgaoParceiroID string) bool {
	escopo := middlewares.GetVagaOrgaoParceiroIDs(c)
	return escopo == nil || slices.Contains(escopo, orgaoParceiroID)
}

// exigirCandidaturaNoEscopo barra a mudança de uma candidatura cuja vaga é de
// órgão fora do escopo do editor (DEV-517). Responde e devolve false quando
// barra.
func (h *CandidaturaHandler) exigirCandidaturaNoEscopo(c *gin.Context, id uuid.UUID) bool {
	if middlewares.GetVagaOrgaoParceiroIDs(c) == nil {
		return true
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err)
		return false
	}
	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		return false
	}
	if entity.Vaga == nil || !orgaoNoEscopo(c, entity.Vaga.IDOrgaoParceiro) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: candidatura de vaga fora da sua secretaria"})
		return false
	}
	return true
}

// exigirVagaNoEscopo barra a mudança em lote das candidaturas de uma vaga de
// órgão fora do escopo do editor (DEV-517). Responde e devolve false quando
// barra.
func (h *CandidaturaHandler) exigirVagaNoEscopo(c *gin.Context, vagaID uuid.UUID) bool {
	if middlewares.GetVagaOrgaoParceiroIDs(c) == nil {
		return true
	}

	orgaoParceiroID, err := h.service.GetVagaOrgaoParceiroID(c.Request.Context(), vagaID)
	if err != nil {
		response.RespondError(c, err)
		return false
	}
	if !orgaoNoEscopo(c, orgaoParceiroID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: vaga não pertence à sua secretaria"})
		return false
	}
	return true
}
