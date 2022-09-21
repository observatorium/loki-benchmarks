package utils

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WaitForReadyDeployment(c client.Client, ns, name string, replicas int32, retry, timeout time.Duration) error {
	return wait.Poll(retry, timeout, func() (done bool, err error) {
		dpl := &appsv1.Deployment{}
		key := client.ObjectKey{Name: name, Namespace: ns}

		err = c.Get(context.TODO(), key, dpl)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if dpl.Status.ReadyReplicas >= replicas {
			return true, nil
		}
		return false, nil
	})
}

func WaitUntilReceivedBytes(m metrics.Client, threshold float64, duration string, retry, timeout time.Duration) error {
	dur, err := model.ParseDuration(duration)
	if err != nil {
		return err
	}

	return wait.Poll(retry, timeout, func() (done bool, err error) {
		sample, err := m.DistributorBytesReceivedTotal(dur)
		if err != nil {
			return false, err
		}

		if sample >= threshold {
			return true, nil
		}
		return false, nil
	})
}
