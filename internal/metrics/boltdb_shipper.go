package metrics

import "github.com/prometheus/common/model"

func (c *client) RequestBoltDBShipperReadsQPS(label, job string, duration model.Duration) (float64, error) {
	return c.requestBoltDBShipperQPS(label, job, "QUERY", "success", duration)
}

func (c *client) RequestBoltDBShipperWritesQPS(label, job string, duration model.Duration) (float64, error) {
	return c.requestBoltDBShipperQPS(label, job, "WRITE", "success", duration)
}
