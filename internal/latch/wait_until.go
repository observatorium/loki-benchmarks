package latch

import (
	"fmt"
	"time"

	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/prometheus/common/model"
)

func WaitUntilGreaterOrEqual(m metrics.Client, lm metrics.MetricType, threshold float64, duration string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	dur, parseErr := model.ParseDuration(duration)

	if parseErr != nil {
		return fmt.Errorf("failed to parse duration: %w", parseErr)
	}

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
			sample, err = m.DistributorBytesReceivedTotal(dur)
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
