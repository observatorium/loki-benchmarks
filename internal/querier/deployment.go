package querier

import (
	"fmt"
	"strings"

	"github.com/observatorium/loki-benchmarks/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultImage = "docker.io/grafana/logcli:2.7.3-amd64"
)

func CreateQueriers(reader *config.Reader, cfg *config.Querier) []client.Object {
	image := DefaultImage
	if cfg.Image != "" {
		image = cfg.Image
	}

	var dpls []client.Object
	for id, query := range reader.Queries {
		dpls = append(dpls, NewLogCLIDeployment(
			fmt.Sprintf("%s-querier", strings.ToLower(id)),
			cfg.Namespace, image, cfg.ServiceAccount, cfg.PullURL, cfg.Tenant, query, reader.QueryRange,
			reader.Replicas,
		),
		)
	}

	return dpls
}

func NewLogCLIDeployment(
	name, namespace, image, serviceAccount, clientURL, tenantID, query, window string,
	replicas int32,
) *appsv1.Deployment {
	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "logcli",
				Image: image,
				Command: []string{
					"/bin/sh",
				},
				Args: []string{
					"-c",
					fmt.Sprintf(`while true; do logcli query '%s' --since=%s; sleep 10; done`, query, window),
				},
				Env: []corev1.EnvVar{
					{
						Name:  "LOKI_ORG_ID",
						Value: tenantID,
					},
					{
						Name:  "LOKI_ADDR",
						Value: clientURL,
					},
				},
			},
		},
	}

	if serviceAccount != "" {
		spec.ServiceAccountName = serviceAccount
		spec.Containers[0].Env = append(spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "LOKI_BEARER_TOKEN_FILE",
				Value: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			},
			corev1.EnvVar{
				Name:  "LOKI_CA_CERT_PATH",
				Value: "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt",
			},
		)
	}

	labels := map[string]string{
		"app": "loki-benchmarks-querier",
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: spec,
			},
		},
	}
}
