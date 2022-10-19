package benchmarks_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/utils"
)

var _ = Describe("Scenario: High Volume Writes", func() {

	scenarioCfgs := benchCfg.Scenarios.HighVolumeWrites

	samplingRange := scenarioCfgs.Samples.Range
	samplingCfg := gmeasure.SamplingConfig{
		N:                   scenarioCfgs.Samples.Total,
		Duration:            scenarioCfgs.Samples.Interval * time.Duration(scenarioCfgs.Samples.Total+1),
		MinSamplingInterval: scenarioCfgs.Samples.Interval,
	}

	willQueryForBoltDBShipperMetrics := samplingCfg.Duration >= (15 * time.Minute)

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Writes Benchmark not enabled!")
		}
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg
		generatorCfg := loadclient.GeneratorConfig(scenarioCfg.Writers, benchCfg.Generator)

		Describe(fmt.Sprintf("Configuration: %s", scenarioCfg.Description), func() {
			BeforeEach(func() {
				err := loadclient.CreateDeployment(k8sClient, generatorCfg)
				Expect(err).Should(Succeed(), "Failed to deploy logger")

				err = utils.WaitForReadyDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace, generatorCfg.Replicas, defaultRetry, defaultTimeout)
				Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

				DeferCleanup(func() {
					err := loadclient.DeleteDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace)
					Expect(err).Should(Succeed(), "Failed to delete logger deployment")
				})
			})

			It("should measure metrics", func() {
				e := gmeasure.NewExperiment(scenarioCfg.Description)
				AddReportEntry(e.Name, e)

				e.Sample(func(idx int) {
					// Distributor
					job := benchCfg.Metrics.Jobs.Distributor

					err := metricsClient.Measure(e, metricsClient.RequestWritesQPS, "2xx push (Req/s)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushP99, "2xx push p99", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushP50, "2xx push p50", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkPushAvg, "2xx push avg", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.DistributorGiPDReceivedTotal, "Received Total (Gi/Day)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.DistributorGiPDDiscardedTotal, "Discarded Total (Gi/Day)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Logger
					job = generatorCfg.Name

					err = metricsClient.Measure(e, metricsClient.LoadNetworkTotal, "Load Total (MB/s)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.LoadNetworkGiPDTotal, "Load Total (Gi/Day)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Ingesters
					job = benchCfg.Metrics.Jobs.Ingester

					err = metricsClient.Measure(e, metricsClient.ProcessCPU, "Processes CPU (Mi/Core)", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestWritesGrpcQPS, "successful GRPC push QPS", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushP99, "successful GRPC push p99", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushP50, "successful GRPC push p50", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metricsClient.RequestDurationOkGrpcPushAvg, "successful GRPC push avg", job, samplingRange)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if benchCfg.Metrics.EnableCadvisorMetrics {
						err = metricsClient.Measure(e, metricsClient.ContainerUserCPU, "Containers User CPU (Mi/Core)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
						err = metricsClient.Measure(e, metricsClient.ContainerWorkingSetMEM, "Containers WorkingSet memory (MB)", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					if willQueryForBoltDBShipperMetrics {
						err = metricsClient.Measure(e, metricsClient.RequestBoltDBShipperWritesQPS, "Boltdb shipper successful writes QPS", job, samplingRange)
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}
				}, samplingCfg)
			})
		})
	}
})
