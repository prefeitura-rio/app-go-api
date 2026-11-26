package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
)

// alertingRulesPath is the Prometheus rule file the Infra lane installs
// (docs/observability.md, "Production alerts").
const alertingRulesPath = "../../monitoring/prometheus/alerting-rules.yaml"

type ruleFile struct {
	Groups []struct {
		Name     string `yaml:"name"`
		Interval string `yaml:"interval"`
		Rules    []struct {
			Alert       string            `yaml:"alert"`
			Record      string            `yaml:"record"`
			Expr        string            `yaml:"expr"`
			For         string            `yaml:"for"`
			Labels      map[string]string `yaml:"labels"`
			Annotations map[string]string `yaml:"annotations"`
		} `yaml:"rules"`
	} `yaml:"groups"`
}

type alertRule struct {
	expr        string
	forDuration string
	labels      map[string]string
	annotations map[string]string
}

func loadAlertRules(t *testing.T) map[string]alertRule {
	t.Helper()
	raw, err := os.ReadFile(alertingRulesPath)
	require.NoError(t, err)

	var file ruleFile
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&file), "the rule file must follow the Prometheus rule file format")
	require.NotEmpty(t, file.Groups)

	alerts := map[string]alertRule{}
	for _, g := range file.Groups {
		require.NotEmpty(t, g.Name, "every group has a name")
		for _, r := range g.Rules {
			require.NotEmpty(t, r.Alert, "only alerting rules live in this file")
			require.Empty(t, r.Record)
			require.NotEmpty(t, strings.TrimSpace(r.Expr), r.Alert)
			_, dup := alerts[r.Alert]
			require.False(t, dup, "duplicate alert %s", r.Alert)
			alerts[r.Alert] = alertRule{expr: r.Expr, forDuration: r.For, labels: r.Labels, annotations: r.Annotations}
		}
	}
	return alerts
}

// compact removes whitespace so a PromQL fragment matches however the
// expression is wrapped in the YAML.
func compact(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func TestAlertingRules_CoverTheMinimumProductionAlerts(t *testing.T) {
	alerts := loadAlertRules(t)

	tests := []struct {
		alert     string
		forDur    string
		severity  string
		fragments []string
	}{
		{
			alert:    "GoApiHighErrorRate",
			forDur:   "5m",
			severity: "critical",
			fragments: []string{
				`istio_requests_total{reporter="destination",destination_workload_namespace="go",destination_workload="go",response_code=~"5.."}[5m]`,
				`>0.02`,
			},
		},
		{
			alert:    "GoApiHighLatencyP95",
			forDur:   "10m",
			severity: "warning",
			fragments: []string{
				`histogram_quantile(0.95,sumby(le,route)(rate(goapi_http_request_duration_seconds_bucket{namespace="go",route=~"/api/(v1|public)/.*"}[5m])))>1`,
			},
		},
		{
			alert:    "GoApiPodRestarting",
			forDur:   "",
			severity: "critical",
			fragments: []string{
				`increase(kube_pod_container_status_restarts_total{namespace="go",container="go"}[15m])>3`,
			},
		},
		{
			alert:    "GoApiRMICircuitBreakerOpen",
			forDur:   "5m",
			severity: "warning",
			fragments: []string{
				`rmi_client_circuit_breaker_state{namespace="go"}>0`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.alert, func(t *testing.T) {
			rule, ok := alerts[tt.alert]
			require.True(t, ok, "alert %s is missing", tt.alert)
			assert.Equal(t, tt.forDur, rule.forDuration)
			assert.Equal(t, tt.severity, rule.labels["severity"])
			assert.Equal(t, "app-go-api", rule.labels["service"])
			for _, fragment := range tt.fragments {
				assert.Contains(t, compact(rule.expr), fragment)
			}
			assert.NotEmpty(t, rule.annotations["summary"])
			assert.NotEmpty(t, rule.annotations["description"])
			assert.True(t, strings.HasPrefix(rule.annotations["runbook_url"],
				"https://github.com/prefeitura-rio/app-go-api/blob/main/docs/runbook.md#"), rule.annotations["runbook_url"])
		})
	}
}

// appMetricPattern matches the metrics this service exports itself; istio_* and
// kube_* come from the sidecar and from kube-state-metrics.
var appMetricPattern = regexp.MustCompile(`\b(?:goapi|rmi_client)_[a-z0-9_]+`)

// An alert on a metric name the service does not export would never fire: every
// goapi_*/rmi_client_* metric a rule reads must be served at GET /metrics.
func TestAlertingRules_AppMetricsAreExposed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousOut := gin.DefaultWriter
	gin.DefaultWriter = &bytes.Buffer{}
	t.Cleanup(func() { gin.DefaultWriter = previousOut })

	r := newEngine()
	r.GET("/api/public/alerting-rules-test", func(c *gin.Context) { c.Status(http.StatusOK) })
	registerMetricsRoute(r, nil)
	_ = clients.NewRMIClient("http://rmi.invalid", time.Second, clients.WithName("alerting-rules-test"))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/public/alerting-rules-test", nil))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	seen := 0
	for name, rule := range loadAlertRules(t) {
		for _, metric := range appMetricPattern.FindAllString(rule.expr, -1) {
			seen++
			family := metric
			for _, suffix := range []string{"_bucket", "_count", "_sum"} {
				if base, ok := strings.CutSuffix(metric, suffix); ok && strings.Contains(body, "# TYPE "+base+" histogram") {
					family = base
				}
			}
			assert.Contains(t, body, "# TYPE "+family+" ", "alert %s reads %s, which /metrics does not export", name, metric)
		}
	}
	assert.NotZero(t, seen, "the rules read at least one metric of the service")
}
