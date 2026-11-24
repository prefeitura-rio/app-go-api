package router

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readmeEndpointLine casa uma linha de lista do README no formato
// "- `MÉTODO /caminho` - descrição".
var readmeEndpointLine = regexp.MustCompile("^- `(GET|POST|PUT|PATCH|DELETE) (/[^`\\s]*)`")

// readmeEndpoints devolve os pares "MÉTODO /caminho" listados na seção
// "## Endpoints da API" do README da raiz do repositório.
func readmeEndpoints(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var endpoints []string
	inSection := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "## ") {
			inSection = line == "## Endpoints da API"
			continue
		}
		if !inSection {
			continue
		}
		if m := readmeEndpointLine.FindStringSubmatch(line); m != nil {
			endpoints = append(endpoints, m[1]+" "+m[2])
		}
	}
	require.NoError(t, scanner.Err())
	return endpoints
}

// O README documenta rotas para quem integra com a API: cada rota listada em
// "Endpoints da API" precisa existir no router, com o mesmo método e o mesmo
// caminho (DEV-788: o README listava /api/v1/cursos, que nunca existiu).
func TestREADME_EndpointsDaAPI_RotasRegistradas(t *testing.T) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiPublic := router.Group("/api/public")
	registerCoreRoutes(apiV1, apiPublic, createMockApplicationContainer(), newTestCoreConfig())

	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	documented := readmeEndpoints(t, "../../README.md")
	require.NotEmpty(t, documented, "a seção \"Endpoints da API\" do README deve listar rotas")

	for _, endpoint := range documented {
		assert.True(t, registered[endpoint], "README documenta uma rota que o router não registra: %s", endpoint)
	}
}
