package latch

import (
	"fmt"
	"time"

	"github.com/observatorium/loki-benchmarks/internal/metrics"
)

func WaitUntilGreaterOrEqual(m metrics.Client, lm metrics.MetricType, threshold float64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().UnixNano() == deadline.UnixNano() {
			return fmt.Errorf("deadline exceeded waiting for latch activation: %s", string(lm))
		}

		var (
			sample float64
			err    error
		)

		switch lm {
		case metrics.DistributorBytesReceivedTotal:
			sample, err = m.DistributorBytesReceivedTotal()
			if err != nil {
				continue
			}
		default:
			return fmt.Errorf("unsupported latch metric: %s", string(lm))
		}

		if sample >= threshold {
			break
		}
	}

	return nil
}
