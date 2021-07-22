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

var _ = Describe("Scenario: High Volume Reads", func() {

	var (
		scenarioCfg        config.HighVolumeReads
		configurationCount int
		mu                 sync.Mutex // Guard
		totalSamples       int
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.HighVolumeReads
		if !scenarioCfg.Enabled {
			Skip("High Volumes Reads Benchmark not enabled!")
			return
		}

		if totalSamples == 0 {
			totalSamples = scenarioCfg.Configurations[configurationCount].Samples.Total
			interval := scenarioCfg.Configurations[configurationCount].Samples.Interval
			readerDuration := time.Duration(int64(totalSamples) * int64(interval))
			writerCfg := scenarioCfg.Configurations[configurationCount].Writers
			readerCfg := scenarioCfg.Configurations[configurationCount].Readers

			// delete previous logger deployment from earlier. executions (if exist)
			_ = logger.Undeploy(k8sClient, benchCfg.Logger)

			// deploy loggers
			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.Loki.PushURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			// wait for loggers to be ready
			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaultTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

			// Wait until we ingested enough logs based on startThreshold
			err = latch.WaitUntilGreaterOrEqual(metricsClient, metrics.DistributorBytesReceivedTotal, readerCfg.StartThreshold, defaultLatchTimeout)
			Expect(err).Should(Succeed(), "Failed to wait until latch activated")

			// Undeploy logger to assert only read traffic
			err = logger.Undeploy(k8sClient, benchCfg.Logger)
			Expect(err).Should(Succeed(), "Failed to delete logger deployment")

			// delete previous query clients deployment from earlier. executions (if exist)
			for id := range readerCfg.Queries {
				_ = querier.Undeploy(k8sClient, benchCfg.Querier, id)
			}

			// Deploy the query clients
			for id, query := range readerCfg.Queries {
				err = querier.Deploy(k8sClient, benchCfg.Querier, readerCfg, benchCfg.Loki.QueryFrontend, id, query, readerDuration)
				Expect(err).Should(Succeed(), "Failed to deploy querier")
			}

			for id := range readerCfg.Queries {
				name := querier.DeploymentName(benchCfg.Querier, id)

				err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Querier.Namespace, name, readerCfg.Replicas, defaultRetry, defaultTimeout)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", name))
			}
		}

		time.Sleep(scenarioCfg.Configurations[configurationCount].Samples.Interval)
	})

	AfterEach(func() {
		mu.Lock()
		defer mu.Unlock()

		if totalSamples == 0 {
			readerCfg := scenarioCfg.Configurations[configurationCount].Readers
			for id := range readerCfg.Queries {
				Expect(querier.Undeploy(k8sClient, benchCfg.Querier, id)).Should(Succeed(), "Failed to delete querier deployment")
			}
			configurationCount++
		}
	})

	for _, configuration := range benchCfg.Scenarios.HighVolumeReads.Configurations {
		c := configuration // Make a local copy to avoid the "Using the variable on range scope `verb` in function literal"
		Measure("should result in measurements - configuration: "+c.Description, func(b Benchmarker) {
			defaultRange := scenarioCfg.Configurations[configurationCount].Samples.Range

			// Collect measurements for query frontend
			job := benchCfg.Metrics.QueryFrontendJob()
			err := metricsClient.Measure(b, metricsClient.RequestReadsQPS, "2xx reads QPS", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

			// Collect measurements for querier
			job = benchCfg.Metrics.QuerierJob()
			err = metricsClient.Measure(b, metricsClient.RequestReadsQPS, "2xx reads QPS", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeP99, "2xx reads p99", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeP50, "2xx reads p50", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkQueryRangeAvg, "2xx reads avg", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

			// Collect measurements for ingester
			job = benchCfg.Metrics.IngesterJob()
			if c.Samples.Interval > 15*time.Minute {
				err = metricsClient.Measure(b, metricsClient.RequestBoltDBShipperReadsQPS, "Boltdb shipper reads QPS", job, c.Description, defaultRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcQuerySampleP99, "successful GRPC query p99", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcQuerySampleP50, "successful GRPC query p50", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcQuerySampleAvg, "successful GRPC query avg", job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

			mu.Lock()
			defer mu.Unlock()
			totalSamples -= 1

		}, c.Samples.Total)
	}
})
