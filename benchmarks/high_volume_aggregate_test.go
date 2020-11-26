package benchmarks_test

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/k8s"
	"github.com/observatorium/loki-benchmarks/internal/latch"
	"github.com/observatorium/loki-benchmarks/internal/logger"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/observatorium/loki-benchmarks/internal/querier"
)

var _ = Describe("Scenario: High Volume Aggregate", func() {
	var (
		beforeOnce  sync.Once
		afterOnce   sync.Once
		scenarioCfg config.HighVolumeAggregate
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.HighVolumeAggregate
		if !scenarioCfg.Enabled {
			Skip("High Volume Aggregate Benchmark not enabled!")

			return
		}

		beforeOnce.Do(func() {
			writerCfg := scenarioCfg.Writers
			readerCfg := scenarioCfg.Readers

			// Deploy the logger to ingest logs
			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.Loki.PushURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaulTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			// Wait until we ingested enough logs based on startThreshold
			err = latch.WaitUntilGreaterOrEqual(metricsClient, metrics.DistributorBytesReceivedTotal, readerCfg.StartThreshold, defaultLatchTimeout)
			Expect(err).Should(Succeed(), "Failed to wait until latch activated")

			// Undeploy logger to assert only read traffic
			err = logger.Undeploy(k8sClient, benchCfg.Logger)
			Expect(err).Should(Succeed(), "Failed to delete logger deployment")

			// Deploy the query clients
			for id, query := range readerCfg.Queries {
				err = querier.Deploy(k8sClient, benchCfg.Querier, readerCfg, benchCfg.Loki.QueryURL(), id, query)
				Expect(err).Should(Succeed(), "Failed to deploy querier")
			}

			for id := range readerCfg.Queries {
				name := querier.DeploymentName(benchCfg.Querier, id)

				err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Querier.Namespace, name, readerCfg.Replicas, defaultRetry, defaulTimeout)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", name))
			}
		})

		time.Sleep(scenarioCfg.Samples.Interval)
	})

	AfterEach(func() {
		afterOnce.Do(func() {
			readerCfg := scenarioCfg.Readers
			for id := range readerCfg.Queries {
				Expect(querier.Undeploy(k8sClient, benchCfg.Querier, id)).Should(Succeed(), "Failed to delete querier deployment")
			}
		})
	})

	Measure("should result in measurements of p99, p50 and avg for all successful aggregate requests to the query frontend", func(b Benchmarker) {
		defaultRange := scenarioCfg.Samples.Range

		//
		// Collect measurements for the query frontend
		//
		job := benchCfg.Metrics.QueryFrontendJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err := metricsClient.RequestDurationOkQueryP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p99 for all query frontend aggregate reads with status code 2xx")
		b.RecordValue("All query frontend 2xx aggregate reads p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err := metricsClient.RequestDurationOkQueryP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all query frontend aggregate reads with status code 2xx")
		b.RecordValue("All query frontend 2xx aggregate reads p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err := metricsClient.RequestDurationOkQueryAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all query frontend aggregate reads with status code 2xx")
		b.RecordValue("All query frontend 2xx aggregate reads avg", avg)

		//
		// Collect measurements for the querier
		//
		job = benchCfg.Metrics.QuerierJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err = metricsClient.RequestDurationOkQueryP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p99 for all querier query with status code 2xx")
		b.RecordValue("All querier 2xx query aggregate p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err = metricsClient.RequestDurationOkQueryP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all querier query with status code 2xx")
		b.RecordValue("All querier 2xx query aggregate p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err = metricsClient.RequestDurationOkQueryAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all querier query with status code 2xx")
		b.RecordValue("All querier 2xx query aggregate avg", avg)

		//
		// Collect measurements for the ingester
		//
		job = benchCfg.Metrics.IngesterJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err = metricsClient.RequestDurationOkGrpcQuerySampleP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all ingester query sample with status code 2xx")
		b.RecordValue("All ingester successful query sample aggregate p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err = metricsClient.RequestDurationOkGrpcQuerySampleP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all ingester reads with status code 2xx")
		b.RecordValue("All ingester successful query sample aggregate p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err = metricsClient.RequestDurationOkGrpcQuerySampleAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all ingester reads with status code 2xx")
		b.RecordValue("All ingester successful query sample aggregate avg", avg)
	}, benchCfg.Scenarios.HighVolumeAggregate.Samples.Total)
})
