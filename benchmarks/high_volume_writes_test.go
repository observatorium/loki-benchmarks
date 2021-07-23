package benchmarks_test

import (
	"fmt"
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
		scenarioCfg        config.HighVolumeWrites
		configurationCount int
		mu                 sync.Mutex // Guard
		totalSamples       int
	)

	BeforeEach(func() {
		scenarioCfg = benchCfg.Scenarios.HighVolumeWrites
		if !scenarioCfg.Enabled {
			Skip("High Volumes Writes Benchmark not enabled!")
			return
		}

		if totalSamples == 0 {
			totalSamples = scenarioCfg.Configurations[configurationCount].Samples.Total
			writerCfg := scenarioCfg.Configurations[configurationCount].Writers

			// delete previous logger deployment from earlier. executions (if exist)
			_ = logger.Undeploy(k8sClient, benchCfg.Logger)

			// deploy loggers
			err := logger.Deploy(k8sClient, benchCfg.Logger, writerCfg, benchCfg.Loki.PushURL())
			Expect(err).Should(Succeed(), "Failed to deploy logger")

			// wait for loggers to be ready
			err = k8s.WaitForReadyDeployment(k8sClient, benchCfg.Logger.Namespace, benchCfg.Logger.Name, writerCfg.Replicas, defaultRetry, defaultTimeout)
			Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")
		}

		time.Sleep(scenarioCfg.Configurations[configurationCount].Samples.Interval)
	})

	AfterEach(func() {
		mu.Lock()
		defer mu.Unlock()

		if totalSamples == 0 {
			Expect(logger.Undeploy(k8sClient, benchCfg.Logger)).Should(Succeed(), "Failed to delete logger deployment")
			configurationCount++
		}
	})

	for _, configuration := range benchCfg.Scenarios.HighVolumeWrites.Configurations {
		c := configuration // Make a local copy to avoid the "Using the variable on range scope `verb` in function literal"
		Measure("should result in measurements - configuration: "+c.Description, func(b Benchmarker) {
			defaultRange := scenarioCfg.Configurations[configurationCount].Samples.Range

			// Collect measurements for distributors
			job := benchCfg.Metrics.DistributorJob()
			err := metricsClient.Measure(b, metricsClient.RequestWritesQPS, "2xx push QPS", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkPushP99, "2xx push p99", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkPushP50, "2xx push p50", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkPushAvg, "2xx push avg", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

			// Collect measurements for Ingesters
			job = benchCfg.Metrics.IngesterJob()
			cadvisorJob := benchCfg.Metrics.CadvisorIngesterJob()
			err = metricsClient.Measure(b, metricsClient.ProcessCPU, "Processes CPU", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			if benchCfg.Metrics.EnableCadvisorMetrics {
				err = metricsClient.Measure(b, metricsClient.ContainerUserCPU, "Containers User CPU", cadvisorJob.QueryLabel, cadvisorJob.Job, c.Description, defaultRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				err = metricsClient.Measure(b, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory", cadvisorJob.QueryLabel, cadvisorJob.Job, c.Description, defaultRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}
			err = metricsClient.Measure(b, metricsClient.RequestWritesGrpcQPS, "successful GRPC push QPS", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			if c.Samples.Interval > 15*time.Minute {
				err = metricsClient.Measure(b, metricsClient.RequestBoltDBShipperWritesQPS, "Boltdb shipper successful writes QPS", job.QueryLabel, job.Job, c.Description, defaultRange)
				Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			}
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcPushP99, "successful GRPC push p99", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcPushP50, "successful GRPC push p50", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
			err = metricsClient.Measure(b, metricsClient.RequestDurationOkGrpcPushAvg, "successful GRPC push avg", job.QueryLabel, job.Job, c.Description, defaultRange)
			Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

			mu.Lock()
			defer mu.Unlock()
			totalSamples -= 1

		}, c.Samples.Total)
	}
})
