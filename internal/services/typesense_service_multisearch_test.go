package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typesense/typesense-go/v3/typesense"
)

// internalHost imita o nome interno do Typesense no cluster, que nunca pode
// chegar ao chamador da API (DEV-144).
const internalHost = "typesense.go.svc.cluster.local"

// newFakeTypesense sobe um Typesense falso: as coleções em healthy respondem
// uma busca válida; qualquer outra responde 503 com um corpo que cita o host
// interno, como faria um proxy na frente de um Typesense indisponível.
func newFakeTypesense(t *testing.T, healthy ...string) *httptest.Server {
	t.Helper()
	ok := make(map[string]bool, len(healthy))
	for _, name := range healthy {
		ok[name] = true
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 4 || parts[0] != "collections" || parts[2] != "documents" || parts[3] != "search" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if !ok[parts[1]] {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"message":"dial tcp: lookup ` + internalHost + ` on 10.20.0.10:53: no such host"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"found": 2,
			"page":  1,
			"hits": []map[string]interface{}{
				{"document": map[string]interface{}{"id": "1", "titulo": "Curso A"}, "highlights": []interface{}{}},
				{"document": map[string]interface{}{"id": "2", "titulo": "Curso B"}, "highlights": []interface{}{}},
			},
			"request_params": map[string]interface{}{"collection_name": parts[1], "per_page": 10, "q": "*"},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newServiceFor(url string) *TypesenseService {
	return NewMockTypesenseService(typesense.NewClient(
		typesense.WithServer(url),
		typesense.WithAPIKey("test-key"),
		typesense.WithNumRetries(0),
	))
}

func multiSearchParams(collections ...string) models.MultiCollectionSearchParameters {
	return models.MultiCollectionSearchParameters{
		Collections: collections,
		Params:      models.SearchParameters{Q: "*", QueryBy: "titulo"},
	}
}

// assertNoLeak garante que nada do erro interno aparece no que vai ao chamador.
func assertNoLeak(t *testing.T, payload string, extra ...string) {
	t.Helper()
	for _, s := range append([]string{internalHost, "10.20.0.10", "no such host", "dial tcp", "status code"}, extra...) {
		assert.NotContains(t, payload, s)
	}
}

func TestTypesenseService_SearchMultiCollection_AllCollectionsFail_ReturnsMaskedError(t *testing.T) {
	srv := newFakeTypesense(t)
	svc := newServiceFor(srv.URL)

	result, err := svc.SearchMultiCollection(context.Background(), multiSearchParams("cursos", "empregos"))

	require.Error(t, err)
	assert.Nil(t, result)
	assertNoLeak(t, err.Error())
}

func TestTypesenseService_SearchMultiCollection_UnreachableBackend_ReturnsMaskedError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := strings.TrimPrefix(srv.URL, "http://")
	srv.Close() // nada escuta mais nesse endereço: falha de conexão real
	svc := newServiceFor("http://" + addr)

	result, err := svc.SearchMultiCollection(context.Background(), multiSearchParams("cursos"))

	require.Error(t, err)
	assert.Nil(t, result)
	assertNoLeak(t, err.Error(), addr, "connection refused")
}

func TestTypesenseService_SearchMultiCollection_PartialFailure_MasksFailedCollection(t *testing.T) {
	srv := newFakeTypesense(t, "cursos")
	svc := newServiceFor(srv.URL)

	result, err := svc.SearchMultiCollection(context.Background(), multiSearchParams("cursos", "empregos"))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalFound)
	assert.Equal(t, 2, result.Results["cursos"].Found)
	assert.NotContains(t, result.Results["cursos"].RequestParams, "error")

	failed, ok := result.Results["empregos"]
	require.True(t, ok)
	assert.Equal(t, 0, failed.Found)
	assert.Empty(t, failed.Hits)
	assert.Equal(t, map[string]interface{}{"error": multiSearchCollectionErrorMessage}, failed.RequestParams)

	body, mErr := json.Marshal(result)
	require.NoError(t, mErr)
	assertNoLeak(t, string(body), strings.TrimPrefix(srv.URL, "http://"))
}

func TestTypesenseService_SearchMultiCollection_AllCollectionsSucceed(t *testing.T) {
	srv := newFakeTypesense(t, "cursos", "empregos")
	svc := newServiceFor(srv.URL)

	result, err := svc.SearchMultiCollection(context.Background(), multiSearchParams("cursos", "empregos"))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 4, result.TotalFound)
	assert.Len(t, result.Results, 2)
}
