package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typesense/typesense-go/v3/typesense"
)

// newTypesenseServiceWithFakeServer monta o serviço com um cliente Typesense
// apontado para um servidor HTTP falso que conta as requisições recebidas e
// guarda o caminho da última.
func newTypesenseServiceWithFakeServer(t *testing.T) (*TypesenseService, *int32, *atomic.Value) {
	t.Helper()
	var calls int32
	var lastPath atomic.Value
	lastPath.Store("")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		lastPath.Store(r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"found":1,"page":1,"hits":[{"document":{"id":"1","titulo":"x"},"highlights":[]}],"request_params":{"collection_name":"cursos","per_page":10,"q":"x"}}`))
	}))
	t.Cleanup(srv.Close)

	client := typesense.NewClient(
		typesense.WithServer(srv.URL),
		typesense.WithAPIKey("test-key"),
	)
	return &TypesenseService{client: client}, &calls, &lastPath
}

func TestIsSearchableCollection(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		want       bool
	}{
		{"cursos permitida", "cursos", true},
		{"empregos permitida", "empregos", true},
		{"vazio", "", false},
		{"coleção arbitrária", "usuarios_internos", false},
		{"caixa diferente", "Cursos", false},
		{"espaço ao redor", " cursos", false},
		{"sufixo", "cursos_backup", false},
		{"traversal", "../cursos", false},
		{"barra", "cursos/documents", false},
		{"alias com caractere codificado", "cursos%2F", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSearchableCollection(tt.collection))
		})
	}
}

func TestTypesenseService_SearchDocuments_CollectionNotAllowed_DoesNotCallTypesense(t *testing.T) {
	for _, collection := range []string{"usuarios_internos", "Cursos", "", "../empregos"} {
		t.Run(collection, func(t *testing.T) {
			svc, calls, _ := newTypesenseServiceWithFakeServer(t)

			res, err := svc.SearchDocuments(context.Background(), collection, models.SearchParameters{Q: "*", QueryBy: "x"})

			assert.Nil(t, res)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrCollectionNotAllowed))
			assert.Equal(t, int32(0), atomic.LoadInt32(calls), "o Typesense não pode ser consultado")
		})
	}
}

func TestTypesenseService_SearchDocuments_AllowedCollection_CallsTypesense(t *testing.T) {
	for _, collection := range []string{"cursos", "empregos"} {
		t.Run(collection, func(t *testing.T) {
			svc, calls, lastPath := newTypesenseServiceWithFakeServer(t)

			res, err := svc.SearchDocuments(context.Background(), collection, models.SearchParameters{Q: "x", QueryBy: "titulo"})

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, 1, res.Found)
			assert.Equal(t, int32(1), atomic.LoadInt32(calls))
			assert.True(t, strings.Contains(lastPath.Load().(string), "/collections/"+collection+"/"), "caminho: %s", lastPath.Load())
		})
	}
}

func TestTypesenseService_SearchMultiCollection_CollectionNotAllowed_DoesNotCallTypesense(t *testing.T) {
	svc, calls, _ := newTypesenseServiceWithFakeServer(t)

	res, err := svc.SearchMultiCollection(context.Background(), models.MultiCollectionSearchParameters{
		Collections: []string{"cursos", "usuarios_internos"},
		Params:      models.SearchParameters{Q: "x", QueryBy: "titulo"},
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int32(1), atomic.LoadInt32(calls), "só a coleção permitida chega ao Typesense")
	assert.Equal(t, 1, res.Results["cursos"].Found)
	blocked := res.Results["usuarios_internos"]
	assert.Equal(t, 0, blocked.Found)
	assert.Empty(t, blocked.Hits)
	assert.Equal(t, 1, res.TotalFound)
}
