package loadclient

import (
	"context"
	"fmt"

	"github.com/observatorium/loki-benchmarks/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentConfig struct {
	Name           string
	Namespace      string
	ServiceAccount string
	Labels         map[string]string
	Args           []string
	Image          string
	Replicas       int32
}

func GeneratorConfig(scenarioCfg *config.Writers, cfg *config.Generator) DeploymentConfig {
	config := defaultConfig("generator", cfg.Namespace, cfg.Image, cfg.ServiceAccount, scenarioCfg.Replicas)
	config.Labels = map[string]string{
		"app": "loki-benchmarks-generator",
	}

	args := []string{
		"generate",
	}

	args = append(args, fmt.Sprintf("--%s=%s", "url", cfg.PushURL))
	args = append(args, fmt.Sprintf("--%s=%s", "tenant", cfg.Tenant))

	for k, v := range scenarioCfg.Args {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}

	config.Args = args

	return config
}

func QuerierConfig(scenarioCfg *config.Readers, cfg *config.Querier, url, query, id string) DeploymentConfig {
	querierName := fmt.Sprintf("%s-%s", cfg.Name, strings.ToLower(id))

	config := defaultConfig(querierName, cfg.Namespace, cfg.Image, cfg.ServiceAccount, scenarioCfg.Replicas)
	config.Labels = map[string]string{
		"app": "loki-benchmarks-querier",
	}

	args := []string{
		"query",
	}

	args = append(args, fmt.Sprintf("--%s=%s", "url", url))
	args = append(args, fmt.Sprintf("--%s=%s", "tenant", cfg.TenantID))
	args = append(args, fmt.Sprintf("--%s=%s", "queries", query))

	for k, v := range scenarioCfg.Args {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}

	config.Args = args

	return config
}

func CreateDeployment(c client.Client, cfg DeploymentConfig) error {
	dpl := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(cfg.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: cfg.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cfg.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cfg.Name,
							Image: cfg.Image,
							Args:  cfg.Args,
						},
					},
				},
			},
		},
	}

	if cfg.ServiceAccount != "" {
		dpl.Spec.Template.Spec.ServiceAccountName = cfg.ServiceAccount
	}

	return c.Create(context.TODO(), dpl, &client.CreateOptions{})
}

func DeleteDeployment(c client.Client, name, namespace string) error {
	dpl := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	return c.Delete(context.TODO(), dpl, &client.DeleteOptions{})
}

func defaultConfig(name, namespace, image, serviceAccount string, replicas int32) DeploymentConfig {
	return DeploymentConfig{
		Name:           name,
		Namespace:      namespace,
		Image:          image,
		Replicas:       replicas,
		ServiceAccount: serviceAccount,
	}
}
