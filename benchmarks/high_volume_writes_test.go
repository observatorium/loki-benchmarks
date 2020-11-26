package benchmarks_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/k8s"
	"github.com/observatorium/loki-benchmarks/internal/logger"
)

var _ = Describe("Scenario: High Volume Writes", func() {
	var (
		beforeOnce  sync.Once
		afterOnce   sync.Once
		scenarioCfg config.HighVolumeWrites
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.HighVolumeWrites
		if !scenarioCfg.Enabled {
			Skip("High Volumes Writes Benchmark not enabled!")

			return
		}

		beforeOnce.Do(func() {
			writerCfg := scenarioCfg.Writers

			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.Loki.PushURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaulTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")
		})

		time.Sleep(scenarioCfg.Samples.Interval)
	})

	AfterEach(func() {
		afterOnce.Do(func() {
			Expect(logger.Undeploy(k8sClient, benchCfg.Logger)).Should(Succeed(), "Failed to delete logger deployment")
		})
	})

	Measure("should result in measurements of p99, p50 and avg for all successful write requests to the distributor", func(b Benchmarker) {
		defaultRange := scenarioCfg.Samples.Range

		//
		// Collect measurements for the distributor
		//
		job := benchCfg.Metrics.DistributorJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err := metricsClient.RequestDurationOkPushP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p99 for all distributor push requests with status code 2xx")
		b.RecordValue("All distributor 2xx push p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err := metricsClient.RequestDurationOkPushP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all distributor push requests with status code 2xx")
		b.RecordValue("All distributor 2xx push p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err := metricsClient.RequestDurationOkPushAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all distributor push requests with status code 2xx")
		b.RecordValue("All distributor 2xx push avg", avg)

		//
		// Collect measurements for the ingester
		//
		job = benchCfg.Metrics.IngesterJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err = metricsClient.RequestDurationOkGrpcPushP99(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p99 for all ingester GRPC push with status code 2xx")
		b.RecordValue("All ingester successful GRPC push p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err = metricsClient.RequestDurationOkGrpcPushP50(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read p50 for all ingester GRPC push with status code 2xx")
		b.RecordValue("All ingester successful GRPC push p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err = metricsClient.RequestDurationOkGrpcPushAvg(job, defaultRange)
		Expect(err).Should(Succeed(), "Failed to read average for all ingester GRPC push with status code 2xx")
		b.RecordValue("All ingester successful GRPC push avg", avg)
	}, benchCfg.Scenarios.HighVolumeWrites.Samples.Total)
})
