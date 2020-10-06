package querier

import (
	"context"
	"fmt"
	"io"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/observatorium/loki-benchmarks/internal/config"
)

func Deploy(c config.Client, cfg *config.Querier, scenarioCfg *config.Readers, url, query string) error {

	queryCmd := fmt.Sprintf(
		`while true; do curl -G -s -H 'X-Scope-OrgID: %s' %s --data-urlencode '%s'; sleep 1; done`,
		cfg.TenantID,
		url,
		query,
	)

	switch cli := c.(type) {
	case *config.K8sClient:
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
		cli.Client.Create(context.TODO(), obj, &client.CreateOptions{})
	case *config.LocalClient:
		fmt.Println("Deploying docker querier")
		ctx := context.Background()

		reader, err := cli.Client.ImagePull(ctx, cfg.Image, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		io.Copy(os.Stdout, reader)

		fmt.Printf("network: %s\n", cfg.NetworkID)
		ncfg := &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"eth0": &network.EndpointSettings{
					NetworkID: cfg.NetworkID,
				},
			},
		}
		resp, err := cli.Client.ContainerCreate(ctx, &container.Config{
			Image: cfg.Image,
			Cmd:   []string{"/bin/sh", "-c", queryCmd},
		}, nil, ncfg, nil, cfg.Name)
		if err != nil {
			return err
		}

		if err := cli.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}

		cfg.ID = resp.ID
	default:
		return fmt.Errorf("unknown type client")
	}
	return nil
}

func Undeploy(c config.Client, cfg *config.Querier) error {
	switch cli := c.(type) {
	case *config.K8sClient:
		obj := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.Name,
				Namespace: cfg.Namespace,
			},
		}
		return cli.Client.Delete(context.TODO(), obj, &client.DeleteOptions{})
	case *config.LocalClient:
		return cli.Client.ContainerRemove(context.Background(), cfg.ID, types.ContainerRemoveOptions{Force: true})

	default:
		return fmt.Errorf("Undeploy unknown type client")
	}
}
