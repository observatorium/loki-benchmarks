package benchmarks_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/prometheus/common/model"

	"github.com/observatorium/loki-benchmarks/internal/loadclient"
	"github.com/observatorium/loki-benchmarks/internal/querier"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	"github.com/observatorium/loki-benchmarks/internal/utils"
)

var _ = Describe("Scenario: High Volume Reads", func() {

	scenarioCfgs := benchCfg.Scenarios.HighVolumeReads

	samplingRange := model.Duration(scenarioCfgs.Samples.Interval)
	samplingCfg := gmeasure.SamplingConfig{
		N:                   scenarioCfgs.Samples.Total,
		Duration:            scenarioCfgs.Samples.Interval * time.Duration(scenarioCfgs.Samples.Total+1),
		MinSamplingInterval: scenarioCfgs.Samples.Interval,
	}

	BeforeEach(func() {
		if !scenarioCfgs.Enabled {
			Skip("High Volumes Reads Benchmark not enabled!")
		}

		generatorCfg := loadclient.GeneratorConfig(scenarioCfgs.Generator, benchCfg.Generator)

		err := loadclient.CreateDeployment(k8sClient, generatorCfg)
		Expect(err).Should(Succeed(), "Failed to deploy logger")

		err = utils.WaitForReadyDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace, generatorCfg.Replicas, defaultRetry, defaultTimeout)
		Expect(err).Should(Succeed(), "Failed to wait for ready logger deployment")

		err = utils.WaitUntilReceivedBytes(metricsClient, scenarioCfgs.StartThreshold, defaultRange, defaultRetry, defaultTimeout)
		Expect(err).Should(Succeed(), "Failed to wait until latch activated")

		err = loadclient.DeleteDeployment(k8sClient, generatorCfg.Name, generatorCfg.Namespace)
		Expect(err).Should(Succeed(), "Failed to delete logger deployment")
	})

	for _, scenarioCfg := range scenarioCfgs.Configurations {
		scenarioCfg := scenarioCfg
		querierDpls := querier.CreateQueriers(scenarioCfg.Readers, querierCfg, benchCfg.Loki.QueryFrontend, scenarioCfg.Readers.Queries)

		Describe(fmt.Sprintf("Configuration: %s", scenarioCfg.Description), func() {
			BeforeEach(func() {
				for _, dpl := range querierDpls {
					err := k8sClient.Create(context.TODO(), dpl, &client.CreateOptions{})
					Expect(err).Should(Succeed(), "Failed to deploy querier")

					err = utils.WaitForReadyDeployment(k8sClient, querierCfg.Namespace, dpl.GetName(), defaultRetry, defaultTimeout)
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed to wait for ready querier deployment: %s", dpl.GetName()))
				}

				// Sleeping for the first interval so that the data is accurate for the new workload.
				time.Sleep(scenarioCfgs.Samples.Interval)

				DeferCleanup(func() {
					for _, dpl := range querierDpls {
						err := k8sClient.Delete(context.TODO(), dpl, &client.DeleteOptions{})
						Expect(err).Should(Succeed(), "Failed to delete querier deployment")
					}
				})
			})

			It("should measure metrics", func() {
				e := gmeasure.NewExperiment(scenarioCfg.Description)
				AddReportEntry(e.Name, e)

				e.Sample(func(idx int) {
					// Query Frontend
					job := benchCfg.Metrics.Jobs.QueryFrontend
					annotation := metrics.QueryFrontendAnnotation

					err := metricsClient.Measure(e, metrics.RequestReadsQPS(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangeAvg(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangePercentile(99, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangePercentile(50, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestQueryRangeThroughput(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Querier
					job = benchCfg.Metrics.Jobs.Querier
					annotation = metrics.QuerierAnnotation

					err = metricsClient.Measure(e, metrics.ContainerCPU(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if benchCfg.Metrics.EnableCadvisorMetrics {
						err = metricsClient.Measure(e, metrics.ContainerMemoryWorkingSetBytes(job, samplingRange, annotation))
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					err = metricsClient.Measure(e, metrics.RequestReadsQPS(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangeAvg(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangePercentile(99, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkQueryRangePercentile(50, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestQueryRangeThroughput(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					// Ingesters
					job = benchCfg.Metrics.Jobs.Ingester
					annotation = metrics.IngesterAnnotation

					err = metricsClient.Measure(e, metrics.ContainerCPU(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					if benchCfg.Metrics.EnableCadvisorMetrics {
						err = metricsClient.Measure(e, metrics.ContainerMemoryWorkingSetBytes(job, samplingRange, annotation))
						Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					}

					err = metricsClient.Measure(e, metrics.RequestReadsGrpcQPS(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkGrpcQuerySampleAvg(job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkGrpcQuerySamplePercentile(99, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestDurationOkGrpcQuerySamplePercentile(50, job, samplingRange, annotation))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))

					err = metricsClient.Measure(e, metrics.RequestBoltDBShipperReadsQPS(job, samplingRange))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
					err = metricsClient.Measure(e, metrics.RequestBoltDBShipperReadsAvg(job, samplingRange))
					Expect(err).Should(Succeed(), fmt.Sprintf("Failed - %v", err))
				}, samplingCfg)
			})
		})
	}
})
