package querier

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observatorium/loki-benchmarks/internal/config"
)

func Deploy(c client.Client, cfg *config.Querier, scenarioCfg *config.Readers, url, query string) error {
	queryCmd := fmt.Sprintf(
		`while true; do curl -G -s -H 'X-Scope-OrgID: %s' %s --data-urlencode '%s'; sleep 1; done`,
		cfg.TenantID,
		url,
		query,
	)

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
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
							Name:    cfg.Name,
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

func Undeploy(c client.Client, cfg *config.Querier) error {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
	}

	return c.Delete(context.TODO(), obj, &client.DeleteOptions{})
}
