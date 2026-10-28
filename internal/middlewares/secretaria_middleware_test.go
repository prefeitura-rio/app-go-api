package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

// injectCPF is a test helper that sets the user CPF in the gin context.
func injectCPF(cpf string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cpf != "" {
			c.Set(middlewares.UserCPFKey, cpf)
		}
		c.Next()
	}
}

func TestExtractSecretariaOrgaoIDs_NoCPF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
		return []string{"orgao-1"}, nil
	}))
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestExtractSecretariaOrgaoIDs_NilResolver(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectCPF("12345678900"))
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(nil))
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestExtractSecretariaOrgaoIDs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectCPF("12345678900"))
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
		return []string{"orgao-1", "orgao-2"}, nil
	}))
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids, "is_nil": ids == nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "orgao-1")
	assert.Contains(t, w.Body.String(), "orgao-2")
	assert.Contains(t, w.Body.String(), `"is_nil":false`)
}

func TestExtractSecretariaOrgaoIDs_EmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectCPF("12345678900"))
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
		return []string{}, nil
	}))
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil, "len": len(ids)})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// key is set, so not nil — but length is zero
	assert.Contains(t, w.Body.String(), `"is_nil":false`)
	assert.Contains(t, w.Body.String(), `"len":0`)
}

func TestExtractSecretariaOrgaoIDs_ResolverError_ReadFailOpen(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(injectCPF("12345678900"))
			r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
				return nil, errors.New("app-rmi unavailable")
			}))
			var keySet bool
			r.Handle(method, "/test", func(c *gin.Context) {
				_, keySet = c.Get(middlewares.UserSecretariaOrgaoIDsKey)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			// fail-open on reads: the secretaria key is NOT set, so the request is
			// handled as if the user had no secretaria mapping (full public lists).
			assert.False(t, keySet, "secretaria key must not be set on read when the resolver fails")
		})
	}
}

func TestExtractSecretariaOrgaoIDs_ResolverError_WriteFailClosed(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(injectCPF("12345678900"))
			r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
				return nil, errors.New("app-rmi unavailable")
			}))
			r.Handle(method, "/test", func(c *gin.Context) {
				ids := middlewares.GetUserSecretariaOrgaoIDs(c)
				c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil, "len": len(ids)})
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			// fail-closed on writes: key IS set to an empty slice, not nil — no scope.
			assert.Contains(t, w.Body.String(), `"is_nil":false`)
			assert.Contains(t, w.Body.String(), `"len":0`)
		})
	}
}

func TestExtractSecretariaOrgaoIDs_ResolverError_EditorWriteForbidden(t *testing.T) {
	failing := func(ctx context.Context, cpf string) ([]string, error) {
		return nil, errors.New("app-rmi unavailable")
	}
	tests := []struct {
		name  string
		role  string
		path  string
		guard gin.HandlerFunc
	}{
		{"cursos editor draft", "go:cursos:editor", "/courses/draft", middlewares.CourseDraftAuthorization()},
		{"cursos editor update", "go:cursos:editor", "/courses/1", middlewares.CourseAuthorization()},
		{"vagas editor create", "go:empregabilidade:editor", "/vagas", middlewares.VagaOrgaoInjector()},
		{"vagas editor update", "go:empregabilidade:editor", "/vagas/1", middlewares.VagaAuthorization()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(injectCPF("12345678900"))
			r.Use(func(c *gin.Context) {
				c.Set(middlewares.UserRolesKey, []string{tt.role})
				c.Next()
			})
			r.Use(middlewares.ExtractSecretariaOrgaoIDs(failing))
			r.POST(tt.path, tt.guard, func(c *gin.Context) { c.Status(http.StatusCreated) })

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

func TestExtractSecretariaOrgaoIDs_CPFPassedToResolver(t *testing.T) {
	var resolvedCPF string
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectCPF("98765432100"))
	r.Use(middlewares.ExtractSecretariaOrgaoIDs(func(ctx context.Context, cpf string) ([]string, error) {
		resolvedCPF = cpf
		return []string{"orgao-1"}, nil
	}))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "98765432100", resolvedCPF)
}

func TestGetUserSecretariaOrgaoIDs_KeyNotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"is_nil": ids == nil})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), `"is_nil":true`)
}

func TestGetUserSecretariaOrgaoIDs_KeySetWithList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserSecretariaOrgaoIDsKey, []string{"orgao-x", "orgao-y"})
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		ids := middlewares.GetUserSecretariaOrgaoIDs(c)
		c.JSON(http.StatusOK, gin.H{"ids": ids})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "orgao-x")
	assert.Contains(t, w.Body.String(), "orgao-y")
}

func TestHasSecretariaFilter(t *testing.T) {
	tests := []struct {
		name      string
		setKey    bool
		orgaoIDs  []string
		wantValue bool
	}{
		{"key not set", false, nil, false},
		{"empty list", true, []string{}, false},
		{"single orgao", true, []string{"orgao-1"}, true},
		{"multiple orgaos", true, []string{"orgao-1", "orgao-2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			if tt.setKey {
				ids := tt.orgaoIDs
				r.Use(func(c *gin.Context) {
					c.Set(middlewares.UserSecretariaOrgaoIDsKey, ids)
					c.Next()
				})
			}
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"has_filter": middlewares.HasSecretariaFilter(c)})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if tt.wantValue {
				assert.Contains(t, w.Body.String(), `"has_filter":true`)
			} else {
				assert.Contains(t, w.Body.String(), `"has_filter":false`)
			}
		})
	}
}
