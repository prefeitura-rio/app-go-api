package empregabilidade

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type CurriculoHandler struct {
	service *services.CurriculoService
}

func NewCurriculoHandler(service *services.CurriculoService) *CurriculoHandler {
	return &CurriculoHandler{service: service}
}

// requireOwnership checks that the authenticated user owns the entity (identified by entityCPF).
// Returns false and writes the HTTP error response if the check fails.
func requireOwnership(c *gin.Context, entityCPF string) bool {
	if middlewares.IsAdmin(c) {
		return true
	}
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return false
	}
	if userCPF != entityCPF {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode modificar seus próprios dados"})
		return false
	}
	return true
}

// requirePathCPFOwnership checks that the authenticated user matches the :cpf path param.
// Returns false and writes the HTTP error response if the check fails.
func requirePathCPFOwnership(c *gin.Context) bool {
	pathCPF := c.Param("cpf")
	if middlewares.IsAdmin(c) {
		return true
	}
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return false
	}
	if userCPF != pathCPF {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode acessar seus próprios dados"})
		return false
	}
	return true
}

// @Summary      Buscar currículo completo
// @Description  Retorna o currículo completo de um usuário
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf   path      string  true  "CPF do usuário"
// @Success      200   {object}  empregabilidade.CurriculoCompleto
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf} [get]
func (h *CurriculoHandler) GetCurriculoCompleto(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}

	cpf := c.Param("cpf")

	curriculo, err := h.service.GetCurriculoCompleto(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, curriculo)
}

// Formações

// @Summary      Criar formação
// @Description  Cria uma nova formação no currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoFormacao  true  "Dados da formação"
// @Success      201   {object}  empregabilidade.CurriculoFormacao
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/formacoes [post]
func (h *CurriculoHandler) CreateFormacao(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoFormacao
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	id, err := h.service.CreateFormacao(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Buscar formação por ID
// @Description  Retorna uma formação específica
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da formação"
// @Success      200  {object}  empregabilidade.CurriculoFormacao
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/formacoes/{id} [get]
func (h *CurriculoHandler) GetFormacaoByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetFormacaoByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Formação não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar formação
// @Description  Atualiza uma formação existente do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "ID da formação"
// @Param        body  body      empregabilidade.CurriculoFormacao  true  "Dados da formação"
// @Success      200   {object}  empregabilidade.CurriculoFormacao
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/formacoes/{id} [put]
func (h *CurriculoHandler) UpdateFormacao(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetFormacaoByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Formação não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.CurriculoFormacao
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.CPF = existing.CPF
	if err := h.service.UpdateFormacao(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir formação
// @Description  Remove uma formação do currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da formação"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/formacoes/{id} [delete]
func (h *CurriculoHandler) DeleteFormacao(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetFormacaoByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Formação não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	if err := h.service.DeleteFormacao(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Formação excluída com sucesso"})
}

// @Summary      Listar formações por CPF
// @Description  Retorna todas as formações do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/formacoes [get]
func (h *CurriculoHandler) ListFormacoesByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entities, err := h.service.ListFormacoesByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// Idiomas

// @Summary      Criar idioma no currículo
// @Description  Adiciona um novo idioma ao currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoIdioma  true  "Dados do idioma"
// @Success      201   {object}  empregabilidade.CurriculoIdioma
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/idiomas [post]
func (h *CurriculoHandler) CreateIdioma(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	id, err := h.service.CreateIdioma(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Buscar idioma por ID
// @Description  Retorna um idioma específico do currículo
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID do idioma"
// @Success      200  {object}  empregabilidade.CurriculoIdioma
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/idiomas/{id} [get]
func (h *CurriculoHandler) GetIdiomaByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetIdiomaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idioma não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar idioma no currículo
// @Description  Atualiza um idioma existente do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        id    path      string                           true  "ID do idioma"
// @Param        body  body      empregabilidade.CurriculoIdioma  true  "Dados do idioma"
// @Success      200   {object}  empregabilidade.CurriculoIdioma
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/idiomas/{id} [put]
func (h *CurriculoHandler) UpdateIdioma(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetIdiomaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idioma não encontrado"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.CurriculoIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.CPF = existing.CPF
	if err := h.service.UpdateIdioma(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir idioma do currículo
// @Description  Remove um idioma do currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID do idioma"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/idiomas/{id} [delete]
func (h *CurriculoHandler) DeleteIdioma(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetIdiomaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idioma não encontrado"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	if err := h.service.DeleteIdioma(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idioma excluído com sucesso"})
}

// @Summary      Listar idiomas por CPF
// @Description  Retorna todos os idiomas do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/idiomas [get]
func (h *CurriculoHandler) ListIdiomasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entities, err := h.service.ListIdiomasByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// Cursos Complementares

// @Summary      Criar curso complementar
// @Description  Adiciona um novo curso complementar ao currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoCursoComplementar  true  "Dados do curso complementar"
// @Success      201   {object}  empregabilidade.CurriculoCursoComplementar
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/cursos-complementares [post]
func (h *CurriculoHandler) CreateCursoComplementar(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoCursoComplementar
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	id, err := h.service.CreateCursoComplementar(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Buscar curso complementar por ID
// @Description  Retorna um curso complementar específico do currículo
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID do curso complementar"
// @Success      200  {object}  empregabilidade.CurriculoCursoComplementar
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/cursos-complementares/{id} [get]
func (h *CurriculoHandler) GetCursoComplementarByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetCursoComplementarByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso complementar não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar curso complementar
// @Description  Atualiza um curso complementar existente do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        id    path      string                                      true  "ID do curso complementar"
// @Param        body  body      empregabilidade.CurriculoCursoComplementar  true  "Dados do curso complementar"
// @Success      200   {object}  empregabilidade.CurriculoCursoComplementar
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/cursos-complementares/{id} [put]
func (h *CurriculoHandler) UpdateCursoComplementar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetCursoComplementarByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso complementar não encontrado"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.CurriculoCursoComplementar
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.CPF = existing.CPF
	if err := h.service.UpdateCursoComplementar(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir curso complementar
// @Description  Remove um curso complementar do currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID do curso complementar"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/cursos-complementares/{id} [delete]
func (h *CurriculoHandler) DeleteCursoComplementar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetCursoComplementarByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso complementar não encontrado"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	if err := h.service.DeleteCursoComplementar(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Curso complementar excluído com sucesso"})
}

// @Summary      Listar cursos complementares por CPF
// @Description  Retorna todos os cursos complementares do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/cursos-complementares [get]
func (h *CurriculoHandler) ListCursosComplementaresByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entities, err := h.service.ListCursosComplementaresByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// Experiências

// @Summary      Criar experiência
// @Description  Adiciona uma nova experiência profissional ao currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoExperiencia  true  "Dados da experiência"
// @Success      201   {object}  empregabilidade.CurriculoExperiencia
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/experiencias [post]
func (h *CurriculoHandler) CreateExperiencia(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoExperiencia
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	id, err := h.service.CreateExperiencia(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Buscar experiência por ID
// @Description  Retorna uma experiência específica do currículo
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da experiência"
// @Success      200  {object}  empregabilidade.CurriculoExperiencia
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/experiencias/{id} [get]
func (h *CurriculoHandler) GetExperienciaByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetExperienciaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiência não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar experiência
// @Description  Atualiza uma experiência existente do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        id    path      string                                true  "ID da experiência"
// @Param        body  body      empregabilidade.CurriculoExperiencia  true  "Dados da experiência"
// @Success      200   {object}  empregabilidade.CurriculoExperiencia
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/experiencias/{id} [put]
func (h *CurriculoHandler) UpdateExperiencia(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetExperienciaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiência não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.CurriculoExperiencia
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.CPF = existing.CPF
	if err := h.service.UpdateExperiencia(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir experiência
// @Description  Remove uma experiência do currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da experiência"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/experiencias/{id} [delete]
func (h *CurriculoHandler) DeleteExperiencia(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetExperienciaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiência não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	if err := h.service.DeleteExperiencia(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Experiência excluída com sucesso"})
}

// @Summary      Listar experiências por CPF
// @Description  Retorna todas as experiências profissionais do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/experiencias [get]
func (h *CurriculoHandler) ListExperienciasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entities, err := h.service.ListExperienciasByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// Conquistas

// @Summary      Criar conquista
// @Description  Adiciona uma nova conquista ao currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoConquista  true  "Dados da conquista"
// @Success      201   {object}  empregabilidade.CurriculoConquista
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/conquistas [post]
func (h *CurriculoHandler) CreateConquista(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	id, err := h.service.CreateConquista(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Buscar conquista por ID
// @Description  Retorna uma conquista específica do currículo
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da conquista"
// @Success      200  {object}  empregabilidade.CurriculoConquista
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/conquistas/{id} [get]
func (h *CurriculoHandler) GetConquistaByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetConquistaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conquista não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar conquista
// @Description  Atualiza uma conquista existente do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        id    path      string                              true  "ID da conquista"
// @Param        body  body      empregabilidade.CurriculoConquista  true  "Dados da conquista"
// @Success      200   {object}  empregabilidade.CurriculoConquista
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/conquistas/{id} [put]
func (h *CurriculoHandler) UpdateConquista(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetConquistaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conquista não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	var entity empregabilidade.CurriculoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.CPF = existing.CPF
	if err := h.service.UpdateConquista(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir conquista
// @Description  Remove uma conquista do currículo do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      string  true  "ID da conquista"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/conquistas/{id} [delete]
func (h *CurriculoHandler) DeleteConquista(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, err := h.service.GetConquistaByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conquista não encontrada"})
		return
	}
	if !requireOwnership(c, existing.CPF) {
		return
	}

	if err := h.service.DeleteConquista(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conquista excluída com sucesso"})
}

// @Summary      Listar conquistas por CPF
// @Description  Retorna todas as conquistas do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/conquistas [get]
func (h *CurriculoHandler) ListConquistasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entities, err := h.service.ListConquistasByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// Bulk replace-all (accordion save)

// @Summary      Substituir formações e idiomas por CPF
// @Description  Remove todas as formações e idiomas do CPF e insere os novos em uma única transação
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        cpf   path      string                                      true  "CPF do usuário"
// @Param        body  body      empregabilidade.FormacaoAccordionRequest    true  "Formações e idiomas"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/formacoes [put]
func (h *CurriculoHandler) ReplaceAllFormacoesByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	var req empregabilidade.FormacaoAccordionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReplaceAllFormacaoAccordionByCPF(c.Request.Context(), cpf, req.Formacoes, req.Idiomas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"formacoes": req.Formacoes, "idiomas": req.Idiomas})
}

// @Summary      Substituir experiências e conquistas por CPF
// @Description  Remove todas as experiências e conquistas do CPF e insere as novas em uma transação
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        cpf   path      string                                                          true  "CPF do usuário"
// @Param        body  body      empregabilidade.ExperienciaProfissionalAccordionRequest         true  "Experiências e conquistas"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/experiencias [put]
func (h *CurriculoHandler) ReplaceAllExperienciasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	var req empregabilidade.ExperienciaProfissionalAccordionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReplaceAllExperienciaProfissionalAccordionByCPF(c.Request.Context(), cpf, req.Experiencias, req.Conquistas, req.ResumoProfissional); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"experiencias": req.Experiencias, "conquistas": req.Conquistas})
}

// @Summary      Substituir conquistas por CPF
// @Description  Remove todas as conquistas do CPF e insere as novas em uma transação
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        cpf   path      string                               true  "CPF do usuário"
// @Param        body  body      []empregabilidade.CurriculoConquista true  "Lista de conquistas"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/conquistas [put]
func (h *CurriculoHandler) ReplaceAllConquistasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	var items []*empregabilidade.CurriculoConquista
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReplaceAllConquistasByCPF(c.Request.Context(), cpf, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Substituir idiomas por CPF
// @Description  Remove todos os idiomas do CPF e insere os novos em uma transação
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        cpf   path      string                             true  "CPF do usuário"
// @Param        body  body      []empregabilidade.CurriculoIdioma  true  "Lista de idiomas"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/idiomas [put]
func (h *CurriculoHandler) ReplaceAllIdiomasByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	var items []*empregabilidade.CurriculoIdioma
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReplaceAllIdiomasByCPF(c.Request.Context(), cpf, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary      Substituir cursos complementares por CPF
// @Description  Remove todos os cursos complementares do CPF e insere os novos em uma transação
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        cpf   path      string                                          true  "CPF do usuário"
// @Param        body  body      []empregabilidade.CurriculoCursoComplementar    true  "Lista de cursos complementares"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/cursos-complementares [put]
func (h *CurriculoHandler) ReplaceAllCursosComplementaresByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	var items []*empregabilidade.CurriculoCursoComplementar
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReplaceAllCursosComplementaresByCPF(c.Request.Context(), cpf, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Situação e Interesses

// @Summary      Criar ou atualizar situação e interesses
// @Description  Cria ou atualiza a situação atual e interesses do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.CurriculoSituacaoInteresses  true  "Dados da situação e interesses"
// @Success      200   {object}  empregabilidade.CurriculoSituacaoInteresses
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/situacao-interesses [put]
func (h *CurriculoHandler) UpsertSituacaoInteresses(c *gin.Context) {
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
		return
	}

	var entity empregabilidade.CurriculoSituacaoInteresses
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.CPF = userCPF

	if err := h.service.UpsertSituacaoInteresses(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Buscar situação e interesses por CPF
// @Description  Retorna a situação atual e interesses do usuário autenticado
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        cpf  path      string  true  "CPF do usuário"
// @Success      200  {object}  empregabilidade.CurriculoSituacaoInteresses
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/curriculo/{cpf}/situacao-interesses [get]
func (h *CurriculoHandler) GetSituacaoInteressesByCPF(c *gin.Context) {
	if !requirePathCPFOwnership(c) {
		return
	}
	cpf := c.Param("cpf")

	entity, err := h.service.GetSituacaoInteressesByCPF(c.Request.Context(), cpf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Situação e interesses não encontrados"})
		return
	}

	c.JSON(http.StatusOK, entity)
}
