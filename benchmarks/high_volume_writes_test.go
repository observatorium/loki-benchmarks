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

var _ = Describe("High Volume Writes", func() {

	BeforeEach(func() {
		err := logger.Deploy(k8sClient, benchCfg.Logger, benchCfg.Loki.PushURL())
		Expect(err).Should(Succeed(), "Failed to deploy logger")

		err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, benchCfg.Logger.Replicas, defaultRetry, defaulTimeout)
		Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")
	})

	AfterEach(func() {
		Expect(logger.Undeploy(k8sClient, benchCfg.Logger)).Should(Succeed(), "Failed to delete logger deployment")
	})

	Measure("should result in a p99 <= 300ms for all successful requests", func(b Benchmarker) {
		val, err := metricsClient.RequestDurationOkWritesP99(benchCfg.Metrics.DistributorJob(), "1m")

		Expect(err).Should(Succeed(), "Failed to read P99 for all distributor writes with status code 2xx")
		Expect(val).Should(BeNumerically("<", 0.3), "Request Duration should not take too long.")

		b.RecordValue("All distributor 2xx Writes p99", val, "")
	}, 10)

})
