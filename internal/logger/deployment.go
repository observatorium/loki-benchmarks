package logger

import (
	"context"
	"fmt"

	"github.com/observatorium/loki-benchmarks/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Deploy(c client.Client, cfg *config.Logger, scenarioCfg *config.Writers, pushURL string) error {
	var args []string

	if scenarioCfg.Command != "" {
		args = append(args, scenarioCfg.Command)
	}

	args = append(args, fmt.Sprintf("--%s=%s", "url", pushURL))
	args = append(args, fmt.Sprintf("--%s=%s", "tenant", cfg.TenantID))

	for k, v := range scenarioCfg.Args {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}

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
							Args:  args,
						},
					},
					ServiceAccountName: "lokistack-dev-benchmarks-logger",
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
