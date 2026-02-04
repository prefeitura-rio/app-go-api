package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

func handleLookupError(c *gin.Context, err error) {
	dbErr := utils.ParseDatabaseError(err)
	c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
}

// ModeloTrabalhoHandler

type ModeloTrabalhoHandler struct {
	service *services.ModeloTrabalhoService
}

func NewModeloTrabalhoHandler(service *services.ModeloTrabalhoService) *ModeloTrabalhoHandler {
	return &ModeloTrabalhoHandler{service: service}
}

// @Summary      Criar modelo de trabalho
// @Description  Cria um novo modelo de trabalho
// @Tags         empregabilidade-modelos-trabalho
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.ModeloTrabalho  true  "Dados do modelo de trabalho"
// @Success      201   {object}  empregabilidade.ModeloTrabalho
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/modelos-trabalho [post]
func (h *ModeloTrabalhoHandler) Create(c *gin.Context) {
	var entity empregabilidade.ModeloTrabalho
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar modelos de trabalho
// @Description  Retorna uma lista paginada de modelos de trabalho
// @Tags         empregabilidade-modelos-trabalho
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/modelos-trabalho [get]
func (h *ModeloTrabalhoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar modelo de trabalho por ID
// @Description  Retorna um modelo de trabalho específico
// @Tags         empregabilidade-modelos-trabalho
// @Produce      json
// @Param        id   path      string  true  "ID do modelo de trabalho"
// @Success      200  {object}  empregabilidade.ModeloTrabalho
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/modelos-trabalho/{id} [get]
func (h *ModeloTrabalhoHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Modelo de trabalho não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar modelo de trabalho
// @Description  Atualiza um modelo de trabalho existente
// @Tags         empregabilidade-modelos-trabalho
// @Accept       json
// @Produce      json
// @Param        id    path      string                          true  "ID do modelo de trabalho"
// @Param        body  body      empregabilidade.ModeloTrabalho  true  "Dados do modelo de trabalho"
// @Success      200   {object}  empregabilidade.ModeloTrabalho
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/modelos-trabalho/{id} [put]
func (h *ModeloTrabalhoHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.ModeloTrabalho
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir modelo de trabalho
// @Description  Remove um modelo de trabalho
// @Tags         empregabilidade-modelos-trabalho
// @Produce      json
// @Param        id   path      string  true  "ID do modelo de trabalho"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/modelos-trabalho/{id} [delete]
func (h *ModeloTrabalhoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Modelo de trabalho excluído com sucesso"})
}

// TipoPCDHandler

type TipoPCDHandler struct {
	service *services.TipoPCDService
}

func NewTipoPCDHandler(service *services.TipoPCDService) *TipoPCDHandler {
	return &TipoPCDHandler{service: service}
}

// @Summary      Criar tipo de PCD
// @Description  Cria um novo tipo de PCD
// @Tags         empregabilidade-tipos-pcd
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.TipoPCD  true  "Dados do tipo de PCD"
// @Success      201   {object}  empregabilidade.TipoPCD
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-pcd [post]
func (h *TipoPCDHandler) Create(c *gin.Context) {
	var entity empregabilidade.TipoPCD
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar tipos de PCD
// @Description  Retorna uma lista paginada de tipos de PCD
// @Tags         empregabilidade-tipos-pcd
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-pcd [get]
func (h *TipoPCDHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar tipo de PCD por ID
// @Description  Retorna um tipo de PCD específico
// @Tags         empregabilidade-tipos-pcd
// @Produce      json
// @Param        id   path      string  true  "ID do tipo de PCD"
// @Success      200  {object}  empregabilidade.TipoPCD
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-pcd/{id} [get]
func (h *TipoPCDHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tipo PCD não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar tipo de PCD
// @Description  Atualiza um tipo de PCD existente
// @Tags         empregabilidade-tipos-pcd
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "ID do tipo de PCD"
// @Param        body  body      empregabilidade.TipoPCD  true  "Dados do tipo de PCD"
// @Success      200   {object}  empregabilidade.TipoPCD
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-pcd/{id} [put]
func (h *TipoPCDHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.TipoPCD
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir tipo de PCD
// @Description  Remove um tipo de PCD
// @Tags         empregabilidade-tipos-pcd
// @Produce      json
// @Param        id   path      string  true  "ID do tipo de PCD"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-pcd/{id} [delete]
func (h *TipoPCDHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo PCD excluído com sucesso"})
}

// IdiomaHandler

type IdiomaHandler struct {
	service *services.IdiomaService
}

func NewIdiomaHandler(service *services.IdiomaService) *IdiomaHandler {
	return &IdiomaHandler{service: service}
}

// @Summary      Criar idioma
// @Description  Cria um novo idioma
// @Tags         empregabilidade-idiomas
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.Idioma  true  "Dados do idioma"
// @Success      201   {object}  empregabilidade.Idioma
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/idiomas [post]
func (h *IdiomaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Idioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar idiomas
// @Description  Retorna uma lista paginada de idiomas
// @Tags         empregabilidade-idiomas
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/idiomas [get]
func (h *IdiomaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar idioma por ID
// @Description  Retorna um idioma específico
// @Tags         empregabilidade-idiomas
// @Produce      json
// @Param        id   path      string  true  "ID do idioma"
// @Success      200  {object}  empregabilidade.Idioma
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/idiomas/{id} [get]
func (h *IdiomaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Idioma não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar idioma
// @Description  Atualiza um idioma existente
// @Tags         empregabilidade-idiomas
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "ID do idioma"
// @Param        body  body      empregabilidade.Idioma  true  "Dados do idioma"
// @Success      200   {object}  empregabilidade.Idioma
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/idiomas/{id} [put]
func (h *IdiomaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Idioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir idioma
// @Description  Remove um idioma
// @Tags         empregabilidade-idiomas
// @Produce      json
// @Param        id   path      string  true  "ID do idioma"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/idiomas/{id} [delete]
func (h *IdiomaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Idioma excluído com sucesso"})
}

// NivelIdiomaHandler

type NivelIdiomaHandler struct {
	service *services.NivelIdiomaService
}

func NewNivelIdiomaHandler(service *services.NivelIdiomaService) *NivelIdiomaHandler {
	return &NivelIdiomaHandler{service: service}
}

// @Summary      Criar nível de idioma
// @Description  Cria um novo nível de idioma
// @Tags         empregabilidade-niveis-idioma
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.NivelIdioma  true  "Dados do nível de idioma"
// @Success      201   {object}  empregabilidade.NivelIdioma
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/niveis-idioma [post]
func (h *NivelIdiomaHandler) Create(c *gin.Context) {
	var entity empregabilidade.NivelIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar níveis de idioma
// @Description  Retorna uma lista paginada de níveis de idioma
// @Tags         empregabilidade-niveis-idioma
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/niveis-idioma [get]
func (h *NivelIdiomaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar nível de idioma por ID
// @Description  Retorna um nível de idioma específico
// @Tags         empregabilidade-niveis-idioma
// @Produce      json
// @Param        id   path      string  true  "ID do nível de idioma"
// @Success      200  {object}  empregabilidade.NivelIdioma
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/niveis-idioma/{id} [get]
func (h *NivelIdiomaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nível de idioma não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar nível de idioma
// @Description  Atualiza um nível de idioma existente
// @Tags         empregabilidade-niveis-idioma
// @Accept       json
// @Produce      json
// @Param        id    path      string                       true  "ID do nível de idioma"
// @Param        body  body      empregabilidade.NivelIdioma  true  "Dados do nível de idioma"
// @Success      200   {object}  empregabilidade.NivelIdioma
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/niveis-idioma/{id} [put]
func (h *NivelIdiomaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.NivelIdioma
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir nível de idioma
// @Description  Remove um nível de idioma
// @Tags         empregabilidade-niveis-idioma
// @Produce      json
// @Param        id   path      string  true  "ID do nível de idioma"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/niveis-idioma/{id} [delete]
func (h *NivelIdiomaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nível de idioma excluído com sucesso"})
}

// EscolaridadeHandler

type EscolaridadeHandler struct {
	service *services.EscolaridadeService
}

func NewEscolaridadeHandler(service *services.EscolaridadeService) *EscolaridadeHandler {
	return &EscolaridadeHandler{service: service}
}

// @Summary      Criar escolaridade
// @Description  Cria uma nova escolaridade
// @Tags         empregabilidade-escolaridades
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.Escolaridade  true  "Dados da escolaridade"
// @Success      201   {object}  empregabilidade.Escolaridade
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/escolaridades [post]
func (h *EscolaridadeHandler) Create(c *gin.Context) {
	var entity empregabilidade.Escolaridade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar escolaridades
// @Description  Retorna uma lista paginada de escolaridades
// @Tags         empregabilidade-escolaridades
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/escolaridades [get]
func (h *EscolaridadeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar escolaridade por ID
// @Description  Retorna uma escolaridade específica
// @Tags         empregabilidade-escolaridades
// @Produce      json
// @Param        id   path      string  true  "ID da escolaridade"
// @Success      200  {object}  empregabilidade.Escolaridade
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/escolaridades/{id} [get]
func (h *EscolaridadeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Escolaridade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar escolaridade
// @Description  Atualiza uma escolaridade existente
// @Tags         empregabilidade-escolaridades
// @Accept       json
// @Produce      json
// @Param        id    path      string                        true  "ID da escolaridade"
// @Param        body  body      empregabilidade.Escolaridade  true  "Dados da escolaridade"
// @Success      200   {object}  empregabilidade.Escolaridade
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/escolaridades/{id} [put]
func (h *EscolaridadeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Escolaridade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir escolaridade
// @Description  Remove uma escolaridade
// @Tags         empregabilidade-escolaridades
// @Produce      json
// @Param        id   path      string  true  "ID da escolaridade"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/escolaridades/{id} [delete]
func (h *EscolaridadeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Escolaridade excluída com sucesso"})
}

// TipoConquistaHandler

type TipoConquistaHandler struct {
	service *services.TipoConquistaService
}

func NewTipoConquistaHandler(service *services.TipoConquistaService) *TipoConquistaHandler {
	return &TipoConquistaHandler{service: service}
}

// @Summary      Criar tipo de conquista
// @Description  Cria um novo tipo de conquista
// @Tags         empregabilidade-tipos-conquista
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.TipoConquista  true  "Dados do tipo de conquista"
// @Success      201   {object}  empregabilidade.TipoConquista
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-conquista [post]
func (h *TipoConquistaHandler) Create(c *gin.Context) {
	var entity empregabilidade.TipoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar tipos de conquista
// @Description  Retorna uma lista paginada de tipos de conquista
// @Tags         empregabilidade-tipos-conquista
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-conquista [get]
func (h *TipoConquistaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar tipo de conquista por ID
// @Description  Retorna um tipo de conquista específico
// @Tags         empregabilidade-tipos-conquista
// @Produce      json
// @Param        id   path      string  true  "ID do tipo de conquista"
// @Success      200  {object}  empregabilidade.TipoConquista
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-conquista/{id} [get]
func (h *TipoConquistaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tipo de conquista não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar tipo de conquista
// @Description  Atualiza um tipo de conquista existente
// @Tags         empregabilidade-tipos-conquista
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "ID do tipo de conquista"
// @Param        body  body      empregabilidade.TipoConquista  true  "Dados do tipo de conquista"
// @Success      200   {object}  empregabilidade.TipoConquista
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-conquista/{id} [put]
func (h *TipoConquistaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.TipoConquista
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir tipo de conquista
// @Description  Remove um tipo de conquista
// @Tags         empregabilidade-tipos-conquista
// @Produce      json
// @Param        id   path      string  true  "ID do tipo de conquista"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/tipos-conquista/{id} [delete]
func (h *TipoConquistaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipo de conquista excluído com sucesso"})
}

// SituacaoAtualHandler

type SituacaoAtualHandler struct {
	service *services.SituacaoAtualService
}

func NewSituacaoAtualHandler(service *services.SituacaoAtualService) *SituacaoAtualHandler {
	return &SituacaoAtualHandler{service: service}
}

// @Summary      Criar situação atual
// @Description  Cria uma nova situação atual
// @Tags         empregabilidade-situacoes-atual
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.SituacaoAtual  true  "Dados da situação atual"
// @Success      201   {object}  empregabilidade.SituacaoAtual
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/situacoes-atual [post]
func (h *SituacaoAtualHandler) Create(c *gin.Context) {
	var entity empregabilidade.SituacaoAtual
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar situações atuais
// @Description  Retorna uma lista paginada de situações atuais
// @Tags         empregabilidade-situacoes-atual
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/situacoes-atual [get]
func (h *SituacaoAtualHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar situação atual por ID
// @Description  Retorna uma situação atual específica
// @Tags         empregabilidade-situacoes-atual
// @Produce      json
// @Param        id   path      string  true  "ID da situação atual"
// @Success      200  {object}  empregabilidade.SituacaoAtual
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/situacoes-atual/{id} [get]
func (h *SituacaoAtualHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Situação atual não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar situação atual
// @Description  Atualiza uma situação atual existente
// @Tags         empregabilidade-situacoes-atual
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "ID da situação atual"
// @Param        body  body      empregabilidade.SituacaoAtual  true  "Dados da situação atual"
// @Success      200   {object}  empregabilidade.SituacaoAtual
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/situacoes-atual/{id} [put]
func (h *SituacaoAtualHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.SituacaoAtual
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir situação atual
// @Description  Remove uma situação atual
// @Tags         empregabilidade-situacoes-atual
// @Produce      json
// @Param        id   path      string  true  "ID da situação atual"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/situacoes-atual/{id} [delete]
func (h *SituacaoAtualHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Situação atual excluída com sucesso"})
}

// DisponibilidadeHandler

type DisponibilidadeHandler struct {
	service *services.DisponibilidadeService
}

func NewDisponibilidadeHandler(service *services.DisponibilidadeService) *DisponibilidadeHandler {
	return &DisponibilidadeHandler{service: service}
}

// @Summary      Criar disponibilidade
// @Description  Cria uma nova disponibilidade
// @Tags         empregabilidade-disponibilidades
// @Accept       json
// @Produce      json
// @Param        body  body      empregabilidade.Disponibilidade  true  "Dados da disponibilidade"
// @Success      201   {object}  empregabilidade.Disponibilidade
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/disponibilidades [post]
func (h *DisponibilidadeHandler) Create(c *gin.Context) {
	var entity empregabilidade.Disponibilidade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar disponibilidades
// @Description  Retorna uma lista paginada de disponibilidades
// @Tags         empregabilidade-disponibilidades
// @Produce      json
// @Param        page      query     int  false  "Número da página"  default(1)
// @Param        pageSize  query     int  false  "Tamanho da página"  default(100)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/disponibilidades [get]
func (h *DisponibilidadeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities, "meta": gin.H{"page": page, "page_size": pageSize, "total": total}})
}

// @Summary      Buscar disponibilidade por ID
// @Description  Retorna uma disponibilidade específica
// @Tags         empregabilidade-disponibilidades
// @Produce      json
// @Param        id   path      string  true  "ID da disponibilidade"
// @Success      200  {object}  empregabilidade.Disponibilidade
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/empregabilidade/disponibilidades/{id} [get]
func (h *DisponibilidadeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Disponibilidade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar disponibilidade
// @Description  Atualiza uma disponibilidade existente
// @Tags         empregabilidade-disponibilidades
// @Accept       json
// @Produce      json
// @Param        id    path      string                           true  "ID da disponibilidade"
// @Param        body  body      empregabilidade.Disponibilidade  true  "Dados da disponibilidade"
// @Success      200   {object}  empregabilidade.Disponibilidade
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/v1/empregabilidade/disponibilidades/{id} [put]
func (h *DisponibilidadeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Disponibilidade
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir disponibilidade
// @Description  Remove uma disponibilidade
// @Tags         empregabilidade-disponibilidades
// @Produce      json
// @Param        id   path      string  true  "ID da disponibilidade"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/disponibilidades/{id} [delete]
func (h *DisponibilidadeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Disponibilidade excluída com sucesso"})
}
