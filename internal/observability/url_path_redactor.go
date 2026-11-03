package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// urlPathKey é o atributo semântico com o caminho cru da requisição. Nas
	// rotas com parâmetros (ex.: /api/v1/enrollments/user/:cpf) ele carrega o
	// valor do parâmetro, inclusive CPF.
	urlPathKey = attribute.Key("url.path")
	// httpRouteKey é o template da rota casada (ex.: /api/v1/enrollments/user/:cpf).
	httpRouteKey = attribute.Key("http.route")
	// unmatchedRoutePath substitui url.path quando nenhuma rota casou (404/405),
	// porque aí não há template e o caminho cru pode conter dado pessoal.
	unmatchedRoutePath = "<unmatched>"
)

// urlPathRedactor é um SpanProcessor que impede que o caminho cru da
// requisição (url.path) chegue ao exportador: no início de cada span que tem
// url.path, o valor é trocado pelo template da rota (http.route) ou, sem rota,
// por unmatchedRoutePath. Spans sem url.path não são alterados.
type urlPathRedactor struct{}

func newURLPathRedactor() sdktrace.SpanProcessor {
	return urlPathRedactor{}
}

// OnStart roda depois que os atributos de início do span foram gravados e
// antes de qualquer exportação; SetAttributes sobrescreve a chave existente.
func (urlPathRedactor) OnStart(_ context.Context, s sdktrace.ReadWriteSpan) {
	var (
		hasURLPath bool
		route      string
	)
	for _, kv := range s.Attributes() {
		switch kv.Key {
		case urlPathKey:
			hasURLPath = true
		case httpRouteKey:
			route = kv.Value.AsString()
		}
	}
	if !hasURLPath {
		return
	}
	if route == "" {
		route = unmatchedRoutePath
	}
	s.SetAttributes(urlPathKey.String(route))
}

func (urlPathRedactor) OnEnd(sdktrace.ReadOnlySpan) {}

func (urlPathRedactor) Shutdown(context.Context) error { return nil }

func (urlPathRedactor) ForceFlush(context.Context) error { return nil }
