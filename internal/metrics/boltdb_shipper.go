package metrics

import (
	"github.com/prometheus/common/model"
)

const (
	BoltDBShipperReadsName = "BoltDB Shipper successful reads"
	BoltDBShipperWriteName = "BoltDB Shipper successful writes"

	BoltDBReadsOperation  = "Shipper.Query"
	BoltDBWritesOperation = "WRITE"
)

func RequestBoltDBShipperReadsQPS(job string, duration model.Duration) Measurement {
	return requestBoltDBShipperQPS(BoltDBShipperReadsName, job, BoltDBReadsOperation, "2.*", duration)
}

func RequestBoltDBShipperWritesQPS(job string, duration model.Duration) Measurement {
	return requestBoltDBShipperQPS(BoltDBShipperWriteName, job, BoltDBWritesOperation, "2.*", duration)
}
