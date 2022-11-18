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

func RequestBoltDBShipperReadsAvg(job string, duration model.Duration) Measurement {
	return requestBoltDBShipperAvg(BoltDBShipperReadsName, job, BoltDBReadsOperation, "2.*", duration)
}

func RequestBoltDBShipperWritesQPS(job string, duration model.Duration) Measurement {
	return requestBoltDBShipperQPS(BoltDBShipperWriteName, job, BoltDBWritesOperation, "2.*", duration)
}

func RequestBoltDBShipperWritesAvg(job string, duration model.Duration) Measurement {
	return requestBoltDBShipperAvg(BoltDBShipperWriteName, job, BoltDBWritesOperation, "2.*", duration)
}
