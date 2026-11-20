package router

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type k8sObject struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name        string            `yaml:"name"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		Template struct {
			Metadata struct {
				Annotations map[string]string `yaml:"annotations"`
			} `yaml:"metadata"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

func loadK8sObjects(t *testing.T, path string) []k8sObject {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var objects []k8sObject
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var obj k8sObject
		err := dec.Decode(&obj)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		if obj.Kind != "" {
			objects = append(objects, obj)
		}
	}
	return objects
}

// The Argo CD Application `go` ignores /spec/template/metadata/annotations of
// the Rollout (RespectIgnoreDifferences), so scrape annotations there never
// reach the pods. The scrape config lives on the stable/active Service `go`,
// read by the Prometheus job kubernetes-service-endpoints.
func TestK8sManifests_ScrapeAnnotations_OnServiceNotPodTemplate(t *testing.T) {
	for _, env := range []string{"staging", "prod"} {
		t.Run(env, func(t *testing.T) {
			objects := loadK8sObjects(t, "../../k8s/"+env+"/resources.yaml")

			var service, rollout *k8sObject
			for i := range objects {
				switch {
				case objects[i].Kind == "Service" && objects[i].Metadata.Name == "go":
					service = &objects[i]
				case objects[i].Kind == "Rollout" && objects[i].Metadata.Name == "go":
					rollout = &objects[i]
				}
			}
			require.NotNil(t, service, "Service go")
			require.NotNil(t, rollout, "Rollout go")

			assert.Equal(t, map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "8080",
				"prometheus.io/path":   "/metrics",
			}, service.Metadata.Annotations, "Service go carries the scrape annotations")

			for key := range rollout.Spec.Template.Metadata.Annotations {
				assert.False(t, strings.HasPrefix(key, "prometheus.io/"),
					"pod template annotation %s is ignored by Argo CD and would duplicate the scrape", key)
			}

			for _, obj := range objects {
				if obj.Kind == "Service" && obj.Metadata.Name != "go" {
					for key := range obj.Metadata.Annotations {
						assert.False(t, strings.HasPrefix(key, "prometheus.io/"),
							"Service %s must not be scraped too (it selects the same pods)", obj.Metadata.Name)
					}
				}
			}
		})
	}
}
