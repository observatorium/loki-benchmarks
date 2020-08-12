package benchmarks_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/observatorium/loki-benchmarks/internal/k8s"
	"github.com/observatorium/loki-benchmarks/internal/logger"
)

var (
	defaultRetry  = 5 * time.Second
	defaulTimeout = 30 * time.Second
)

var _ = Describe("Scenario: High Volume Writes", func() {

	BeforeEach(func() {
		err := logger.Deploy(k8sClient, benchCfg.Logger, benchCfg.Loki.PushURL())
		Expect(err).Should(Succeed(), "Failed to deploy logger")

		err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, benchCfg.Logger.Replicas, defaultRetry, defaulTimeout)
		Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")
	})

	AfterEach(func() {
		Expect(logger.Undeploy(k8sClient, benchCfg.Logger)).Should(Succeed(), "Failed to delete logger deployment")
	})

	Measure("should result in measurements of p99, p50 and avg for all successful write requests to the distributor", func(b Benchmarker) {
		job := benchCfg.Metrics.DistributorJob()

		// Record p99 loki_request_duration_seconds_bucket
		p99, err := metricsClient.RequestDurationOkWritesP99(job, "1m")

		Expect(err).Should(Succeed(), "Failed to read p50 for all distributor writes with status code 2xx")
		Expect(p99).Should(BeNumerically("<", benchCfg.Scenarios.HighVolumeWrites.P99), "p99 should not exceed expectation")

		b.RecordValue("All distributor 2xx Writes p99", p99)

		// Record p50 loki_request_duration_seconds_bucket
		p50, err := metricsClient.RequestDurationOkWritesP50(job, "1m")

		Expect(err).Should(Succeed(), "Failed to read p50 for all distributor writes with status code 2xx")
		Expect(p50).Should(BeNumerically("<", benchCfg.Scenarios.HighVolumeWrites.P50), "p50 should not exceed expectation")

		b.RecordValue("All distributor 2xx Writes p50", p50)

		// Record avg from loki_request_duration_seconds_sum / loki_request_duration_seconds_count
		avg, err := metricsClient.RequestDurationOkWritesAvg(job, "1m")

		Expect(err).Should(Succeed(), "Failed to read average for all distributor writes with status code 2xx")
		Expect(avg).Should(BeNumerically("<", benchCfg.Scenarios.HighVolumeWrites.AVG), "avg should not exceed expectation")

		b.RecordValue("All distributor 2xx Writes avg", avg)

	}, 10)

})
