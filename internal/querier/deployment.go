package querier

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observatorium/loki-benchmarks/internal/config"
)

func DeploymentName(cfg *config.Querier, id string) string {
	return cfg.Name + "-" + strings.ToLower(id)
}

func Deploy(c client.Client, cfg *config.Querier, scenarioCfg *config.Readers, url, id, query string) error {
	queryCmd := fmt.Sprintf(
		`while :; do curl -G -o /dev/null -w '%%{response_code}' -H 'X-Scope-OrgID: %s' %s --data-urlencode '%s'; done`,
		cfg.TenantID,
		url,
		query,
	)

	name := DeploymentName(cfg, id)

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cfg.Namespace,
			Labels: map[string]string{
				"app": "loki-benchmarks-querier",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &scenarioCfg.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "loki-benchmarks-querier",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "loki-benchmarks-querier",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    name,
							Image:   cfg.Image,
							Command: []string{"/bin/sh"},
							Args: []string{
								"-c",
								queryCmd,
							},
						},
					},
				},
			},
		},
	}

	return c.Create(context.TODO(), obj, &client.CreateOptions{})
}

func Undeploy(c client.Client, cfg *config.Querier, id string) error {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName(cfg, id),
			Namespace: cfg.Namespace,
		},
	}

	return c.Delete(context.TODO(), obj, &client.DeleteOptions{})
}
