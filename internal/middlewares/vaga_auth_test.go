package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func injectVagaRoles(role string, orgaoID string, secretariaIDs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if role != "" {
			c.Set(middlewares.UserRolesKey, []string{role})
		}
		if orgaoID != "" {
			c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + orgaoID})
		}
		if secretariaIDs != nil {
			c.Set(middlewares.UserSecretariaOrgaoIDsKey, secretariaIDs)
		}
		c.Next()
	}
}

func TestVagaAuthorization_Admin_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	})
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_EmpAdmin_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:admin", "", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_SecretariaOnly_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("", "", []string{"orgao-1"}))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaAuthorization_Editor_WithSecretaria_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "", []string{"orgao-1"}))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_Editor_WithOrgao_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "orgao-42", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_EditorSemCuradoria_WithOrgao_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_sem_curadoria", "orgao-42", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_EditorComCuradoria_WithOrgao_Passes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "orgao-42", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVagaAuthorization_EditorComCuradoria_NoOrgao_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaAuthorization_NoRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaAuthorization_Editor_NoOrgao_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "", nil))
	r.Use(middlewares.VagaAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaListFilter_Admin_NoFilterInjected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	})
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaListFilter_EmpAdmin_NoFilterInjected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:admin", "", nil))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaListFilter_SecretariaOnly_InjectsEmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("", "", []string{"orgao-1", "orgao-2"}))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"len":0`)
	assert.NotContains(t, w.Body.String(), "orgao-1")
}

func TestVagaListFilter_Editor_WithSecretaria_InjectsIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "", []string{"orgao-1", "orgao-2"}))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), "orgao-1")
	assert.Contains(t, w.Body.String(), "orgao-2")
}

func TestVagaListFilter_Editor_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "orgao-42", nil))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), "orgao-42")
	assert.Contains(t, w.Body.String(), `"len":1`)
}

func TestVagaListFilter_EditorComCuradoria_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "orgao-77", nil))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), "orgao-77")
	assert.Contains(t, w.Body.String(), `"len":1`)
}

func TestVagaListFilter_EditorSemCuradoria_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_sem_curadoria", "orgao-99", nil))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), "orgao-99")
}

func TestGetVagaOrgaoParceiroIDs_KeyNotSet_ReturnsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaOrgaoInjector_Admin_NoRestriction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	})
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaOrgaoInjector_EmpAdmin_NoRestriction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:admin", "", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaOrgaoInjector_SecretariaOnly_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("", "", []string{"orgao-1", "orgao-2"}))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaOrgaoInjector_Editor_WithSecretaria_InjectsIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "", []string{"orgao-1", "orgao-2"}))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-1")
	assert.Contains(t, w.Body.String(), "orgao-2")
}

func TestVagaOrgaoInjector_SecretariaUser_EmptyList_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("", "", []string{}))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaOrgaoInjector_Editor_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "orgao-42", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-42")
	assert.Contains(t, w.Body.String(), `"len":1`)
}

func TestVagaOrgaoInjector_EditorComCuradoria_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "orgao-77", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-77")
	assert.Contains(t, w.Body.String(), `"len":1`)
}

func TestVagaOrgaoInjector_EditorComCuradoria_NoOrgao_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaOrgaoInjector_EditorSemCuradoria_InjectsSingleID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor_sem_curadoria", "orgao-99", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-99")
}

func TestVagaOrgaoInjector_Editor_NoOrgao_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "", nil))
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestVagaOrgaoInjector_NoRole_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.VagaOrgaoInjector())
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetVagaAllowedOrgaos_KeyNotSet_ReturnsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestVagaListFilter_NoRole_InjectsEmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":false`)
	assert.Contains(t, w.Body.String(), `"len":0`)
}

func TestVagaListFilter_Editor_NoOrgao_InjectsEmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectVagaRoles("go:empregabilidade:editor", "", nil))
	r.Use(middlewares.VagaListFilter())
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetVagaOrgaoParceiroIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil, "len": len(ids)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":false`)
	assert.Contains(t, w.Body.String(), `"len":0`)
}

// ──────────────────────────────────────────────────────────────────────────────
// VagaOwnershipCheck
// ──────────────────────────────────────────────────────────────────────────────

const ownershipVagaUUID = "550e8400-e29b-41d4-a716-446655440000"

// vagaLoaderStub returns a loader that always yields the given ownership info
// (or nil / error to simulate not-found / failure).
//
// A regra de "quem pode editar vaga publicada" NÃO é responsabilidade deste
// middleware (fica no VagaHandler.Update, coberto em vaga_handler_test.go), por
// isso o stub não carrega mais o status de publicação.
func vagaLoaderStub(orgaoID string, notFound bool, err error) middlewares.VagaLoaderFunc {
	return func(_ context.Context, _ uuid.UUID) (*middlewares.VagaOwnershipInfo, error) {
		if err != nil {
			return nil, err
		}
		if notFound {
			return nil, nil
		}
		return &middlewares.VagaOwnershipInfo{OrgaoParceiroID: orgaoID}, nil
	}
}

func runVagaOwnership(inject gin.HandlerFunc, loader middlewares.VagaLoaderFunc, id string) int {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if inject != nil {
		r.Use(inject)
	}
	r.PUT("/vagas/:id", middlewares.VagaOwnershipCheck(loader),
		func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/vagas/"+id, nil))
	return w.Code
}

func injectAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middlewares.UserRoleKey, "ADMIN")
		c.Next()
	}
}

func TestVagaOwnershipCheck_Admin_BypassesEvenCrossOrg(t *testing.T) {
	code := runVagaOwnership(injectAdminRole(),
		vagaLoaderStub("orgao-b", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusOK, code)
}

func TestVagaOwnershipCheck_EmpAdmin_Bypasses(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:admin", "", nil),
		vagaLoaderStub("orgao-b", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusOK, code)
}

// Non-admin editor acting on a same-org vaga (scope via go:orgao role).
func TestVagaOwnershipCheck_Editor_SameOrg_Passes(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor_com_curadoria", "orgao-a", nil),
		vagaLoaderStub("orgao-a", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusOK, code)
}

// Same as above, but scope comes from the CPF→secretaria mapping.
func TestVagaOwnershipCheck_Editor_SameSecretaria_Passes(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "", []string{"orgao-a", "orgao-x"}),
		vagaLoaderStub("orgao-a", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusOK, code)
}

// Escopo por secretaria e por órgão (role go:orgao) coexistem: uma vaga que casa
// apenas com o órgão do usuário (fora da lista de secretarias) ainda deve passar,
// alinhado ao que a VagaAuthorization autoriza (F4).
func TestVagaOwnershipCheck_Editor_MatchesOrgaoNotSecretaria_Passes(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "orgao-a", []string{"orgao-x", "orgao-y"}),
		vagaLoaderStub("orgao-a", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusOK, code)
}

func TestVagaOwnershipCheck_Editor_CrossOrg_Forbidden(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor_sem_curadoria", "orgao-a", nil),
		vagaLoaderStub("orgao-b", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestVagaOwnershipCheck_Editor_CrossSecretaria_Forbidden(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "", []string{"orgao-a", "orgao-x"}),
		vagaLoaderStub("orgao-b", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestVagaOwnershipCheck_UnrelatedRole_Forbidden(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:cursos:editor", "orgao-a", nil),
		vagaLoaderStub("orgao-a", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestVagaOwnershipCheck_Editor_NoScope_Forbidden(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "", nil),
		vagaLoaderStub("orgao-a", false, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusForbidden, code)
}

func TestVagaOwnershipCheck_NotFound(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "orgao-a", nil),
		vagaLoaderStub("", true, nil), ownershipVagaUUID)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestVagaOwnershipCheck_InvalidID(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "orgao-a", nil),
		vagaLoaderStub("orgao-a", false, nil), "not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestVagaOwnershipCheck_LoaderError(t *testing.T) {
	code := runVagaOwnership(injectVagaRoles("go:empregabilidade:editor", "orgao-a", nil),
		vagaLoaderStub("", false, errors.New("boom")), ownershipVagaUUID)
	assert.Equal(t, http.StatusInternalServerError, code)
}
