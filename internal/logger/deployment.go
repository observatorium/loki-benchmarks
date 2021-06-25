package logger

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/observatorium/loki-benchmarks/internal/config"
)

func Deploy(c client.Client, cfg *config.Logger, scenarioCfg *config.Writers, pushURL string) error {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels: map[string]string{
				"app": "loki-benchmarks-logger",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &scenarioCfg.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "loki-benchmarks-logger",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "loki-benchmarks-logger",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cfg.Name,
							Image: cfg.Image,
							Args: []string{
								fmt.Sprintf("--url=%s", pushURL),
								fmt.Sprintf("--logps=%d", scenarioCfg.Throughput),
								fmt.Sprintf("--tenant=%s", cfg.TenantID),
								fmt.Sprintf("--message-size=%d", scenarioCfg.MessageSize),
							},
						},
					},
				},
			},
		},
	}

	return c.Create(context.TODO(), obj, &client.CreateOptions{})
}

func Undeploy(c client.Client, cfg *config.Logger) error {
	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
	}

	return c.Delete(context.TODO(), obj, &client.DeleteOptions{})
}
