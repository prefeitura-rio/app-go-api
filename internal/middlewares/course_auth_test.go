// course_auth_test.go
package middlewares_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

// Helper functions for injecting test data
func injectCourseRoles(role string, orgaoID string, secretariaIDs []string) gin.HandlerFunc {
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

func setupCourseTestRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	for _, m := range middlewares {
		r.Use(m)
	}
	return r
}

// ============ CourseAuthorization Tests ============

func TestCourseAuthorization_Admin_Passes(t *testing.T) {
	r := setupCourseTestRouter(
		func(c *gin.Context) {
			c.Set(middlewares.UserRoleKey, "ADMIN")
			c.Next()
		},
		middlewares.CourseAuthorization(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseAuthorization_CasaCivil_Passes(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:casa_civil", "", nil),
		middlewares.CourseAuthorization(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseAuthorization_SecretariaUser_Passes(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{"orgao-1", "orgao-2"}),
		middlewares.CourseAuthorization(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseAuthorization_Editor_WithOrgao_Passes(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "orgao-42", nil),
		middlewares.CourseAuthorization(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseAuthorization_NoRole_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(middlewares.CourseAuthorization())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Sem permissão para acessar este recurso")
}

func TestCourseAuthorization_Editor_NoOrgao_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", nil),
		middlewares.CourseAuthorization(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ============ CourseListFilter Tests ============

func TestCourseListFilter_Admin_NoFilterInjected(t *testing.T) {
	r := setupCourseTestRouter(
		func(c *gin.Context) {
			c.Set(middlewares.UserRoleKey, "ADMIN")
			c.Next()
		},
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filter_count": len(filters)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["filter_count"])
}

func TestCourseListFilter_CasaCivil_NoFilterInjected(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:casa_civil", "", nil),
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filter_count": len(filters)})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["filter_count"])
}

func TestCourseListFilter_SecretariaUser_InjectsINFilter(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{"orgao-1", "orgao-2"}),
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, filters, "orgao_id IN")

	ids, ok := filters["orgao_id IN"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []interface{}{"orgao-1", "orgao-2"}, ids)
}

func TestCourseListFilter_SecretariaUser_EmptyIDs_InjectsNoResults(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{}),
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, filters, "_no_results")
	assert.Equal(t, true, filters["_no_results"])
}

func TestCourseListFilter_Editor_InjectsOrgaoIDFilter(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "orgao-42", nil),
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, filters, "orgao_id")
	assert.Equal(t, "orgao-42", filters["orgao_id"])
}

func TestCourseListFilter_Editor_NoOrgao_InjectsNoResults(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", nil),
		middlewares.CourseListFilter(),
	)
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, filters, "_no_results")
	assert.Equal(t, true, filters["_no_results"])
}

// ============ GetCourseFilters Tests ============

func TestGetCourseFilters_KeyNotSet_ReturnsEmptyMap(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Empty(t, filters)
}

func TestGetCourseFilters_KeyExists_ReturnsFilters(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_filters", map[string]interface{}{
			"orgao_id": "orgao-123",
			"status":   "active",
		})
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "orgao-123", filters["orgao_id"])
	assert.Equal(t, "active", filters["status"])
}

func TestGetCourseFilters_KeyExistsWithWrongType_ReturnsEmptyMap(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_filters", "invalid_type")
		filters := middlewares.GetCourseFilters(c)
		c.JSON(http.StatusOK, gin.H{"filters": filters})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	filters, ok := response["filters"].(map[string]interface{})
	assert.True(t, ok)
	assert.Empty(t, filters)
}

// ============ CourseOrgaoInjector Tests ============

func TestCourseOrgaoInjector_Admin_SetsNilAllowed(t *testing.T) {
	r := setupCourseTestRouter(
		func(c *gin.Context) {
			c.Set(middlewares.UserRoleKey, "ADMIN")
			c.Next()
		},
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) {
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Nil(t, response["allowed"])
}

func TestCourseOrgaoInjector_CasaCivil_SetsAllowedOrgaosNil(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:casa_civil", "", nil),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) {
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Nil(t, response["allowed"])
}

func TestCourseOrgaoInjector_SecretariaUser_SetsAllowedOrgaosList(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{"orgao-1", "orgao-2"}),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) {
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	allowed, ok := response["allowed"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []interface{}{"orgao-1", "orgao-2"}, allowed)
}

func TestCourseOrgaoInjector_SecretariaUser_EmptyIDs_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{}),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Sem permissão para criar: nenhuma secretaria associada")
}

func TestCourseOrgaoInjector_Editor_WithOrgao_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "orgao-42", nil),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Editores não têm permissão para criar cursos")
}

func TestCourseOrgaoInjector_Editor_NoOrgao_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", nil),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Editores não têm permissão para criar cursos")
}

func TestCourseOrgaoInjector_Editor_WithSecretaria_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", []string{"orgao-1", "orgao-2"}),
		middlewares.CourseOrgaoInjector(),
	)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Editores não têm permissão para criar cursos")
}

func TestCourseOrgaoInjector_NoRole_Forbidden(t *testing.T) {
	r := setupCourseTestRouter(middlewares.CourseOrgaoInjector())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Sem permissão para criar: órgão não identificado")
}

// ============ GetUserAllowedOrgaos Tests ============

func TestGetUserAllowedOrgaos_KeyNotSet_ReturnsNil(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Nil(t, response["allowed"])
}

func TestGetUserAllowedOrgaos_KeyExists_ReturnsList(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_allowed_orgaos", []string{"a", "b"})
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	allowed, ok := response["allowed"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []interface{}{"a", "b"}, allowed)
}

func TestGetUserAllowedOrgaos_KeyExistsWithWrongType_ReturnsNil(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_allowed_orgaos", 123) // tipo errado
		allowed := middlewares.GetUserAllowedOrgaos(c)
		c.JSON(http.StatusOK, gin.H{"allowed": allowed})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Nil(t, response["allowed"])
}

// ============ CourseOwnershipCheck Tests ============

func TestCourseOwnershipCheck_Admin_Passes(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		assert.Equal(t, 123, id)
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		func(c *gin.Context) {
			c.Set(middlewares.UserRoleKey, "ADMIN")
			c.Next()
		},
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) {
		orgaoID := middlewares.GetLoadedCursoOrgaoID(c)
		c.JSON(http.StatusOK, gin.H{"orgao_id": orgaoID})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "orgao-456", response["orgao_id"])
}

func TestCourseOwnershipCheck_CasaCivil_Passes(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:casa_civil", "", nil),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseOwnershipCheck_SecretariaUser_MatchingOrgao_Passes(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{"orgao-456", "orgao-789"}),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseOwnershipCheck_SecretariaUser_NonMatchingOrgao_Forbidden(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{"orgao-789", "orgao-999"}),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Acesso negado: curso não pertence à sua secretaria")
}

func TestCourseOwnershipCheck_EditorWithSecretaria_SecretariaMatchesOrgao_Passes(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", []string{"orgao-456", "orgao-789"}),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseOwnershipCheck_EditorWithSecretaria_SecretariaNoMatch_Forbidden(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "", []string{"orgao-999"}),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Acesso negado: curso não pertence à sua secretaria")
}

func TestCourseOwnershipCheck_EditorUser_NonMatchingOrgao_Forbidden(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}
	r := setupCourseTestRouter(
		injectCourseRoles("go:cursos:editor", "orgao-123", nil),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCourseOwnershipCheck_SecretariaUser_EmptySecretariaIDs_Forbidden(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "orgao-456", true, nil
	}

	r := setupCourseTestRouter(
		injectCourseRoles("", "", []string{}),
		middlewares.CourseOwnershipCheck(loader),
	)
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCourseOwnershipCheck_InvalidCourseID_BadRequest(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "", false, nil
	}

	r := setupCourseTestRouter(middlewares.CourseOwnershipCheck(loader))
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/invalid", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ID do curso inválido")
}

func TestCourseOwnershipCheck_CourseNotFound_NotFound(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "", false, nil
	}

	r := setupCourseTestRouter(middlewares.CourseOwnershipCheck(loader))
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Curso não encontrado")
}

func TestCourseOwnershipCheck_LoaderError_InternalServerError(t *testing.T) {
	loader := func(ctx context.Context, id int) (string, bool, error) {
		return "", false, errors.New("database error")
	}

	r := setupCourseTestRouter(middlewares.CourseOwnershipCheck(loader))
	r.GET("/test/:courseId", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test/123", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Erro ao buscar curso")
	assert.Contains(t, w.Body.String(), "database error")
}

// ============ GetLoadedCursoOrgaoID Tests ============

func TestGetLoadedCursoOrgaoID_KeyNotSet_ReturnsEmpty(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		orgaoID := middlewares.GetLoadedCursoOrgaoID(c)
		c.JSON(http.StatusOK, gin.H{"orgao_id": orgaoID})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "", response["orgao_id"])
}

func TestGetLoadedCursoOrgaoID_KeyExists_ReturnsValue(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_loaded_curso", "orgao-456")
		orgaoID := middlewares.GetLoadedCursoOrgaoID(c)
		c.JSON(http.StatusOK, gin.H{"orgao_id": orgaoID})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "orgao-456", response["orgao_id"])
}

func TestGetLoadedCursoOrgaoID_KeyExistsWithWrongType_ReturnsEmpty(t *testing.T) {
	r := setupCourseTestRouter()
	r.GET("/test", func(c *gin.Context) {
		c.Set("course_loaded_curso", 456) // Wrong type
		orgaoID := middlewares.GetLoadedCursoOrgaoID(c)
		c.JSON(http.StatusOK, gin.H{"orgao_id": orgaoID})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "", response["orgao_id"])
}
