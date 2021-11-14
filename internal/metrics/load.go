package metrics

import (
	"fmt"
	"github.com/prometheus/common/model"
)

func (c *client) LoadNetworkGiPDTotal(_, _ string, duration model.Duration) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(container_network_transmit_bytes_total{pod=~"logger-.*"}[%s])) / %s * 86400`, // in Gi per day
		duration, BytesToGigabytesMultiplier,
	)

	return c.executeScalarQuery(query)
}
