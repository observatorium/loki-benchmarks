package loadclient

import (
	"fmt"

	"github.com/observatorium/loki-benchmarks/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateGenerator(scenarioCfg *config.Writers, cfg *config.Generator) client.Object {
	args := []string{
		"generate",
		fmt.Sprintf("--%s=%s", "url", cfg.PushURL),
		fmt.Sprintf("--%s=%s", "tenant", cfg.Tenant),
		"--destination=loki",
	}

	for k, v := range scenarioCfg.Args {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}

	return NewLoadClientDeployment("generator", cfg.Namespace, cfg.Image, cfg.ServiceAccount, args, scenarioCfg.Replicas)
}

func NewLoadClientDeployment(
	name, namespace, image, serviceAccount string,
	args []string,
	replicas int32,
) *appsv1.Deployment {
	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "loadclient",
				Image: image,
				Args:  args,
			},
		},
	}

	if serviceAccount != "" {
		spec.ServiceAccountName = serviceAccount
	}

	labels := map[string]string{
		"app": "loki-benchmarks-generator",
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(replicas),
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
