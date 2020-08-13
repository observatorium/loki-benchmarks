package benchmarks_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/metrics"
	internalreporters "github.com/observatorium/loki-benchmarks/internal/reporters"
)

var (
	benchCfg      *config.Benchmark
	k8sClient     client.Client
	metricsClient metrics.Client

	defaultRetry       = 5 * time.Second
	defaulTimeout      = 30 * time.Second
	defaltLatchTimeout = 2 * time.Minute
)

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)

	reportDir := os.Getenv("REPORT_DIR")
	if reportDir == "" {
		t.FailNow()
	}

	jr := reporters.NewJUnitReporter(fmt.Sprintf("%s/junit.xml", reportDir))
	csv := internalreporters.NewCsvReporter(reportDir)
	gp := internalreporters.NewGnuplotReporter(reportDir)

	RunSpecsWithDefaultAndCustomReporters(t, "Benchmarks Suite", []Reporter{jr, csv, gp})
}

var _ = BeforeSuite(func() {
	// Read target environment
	env := os.Getenv("TARGET_ENV")
	Expect(env).ToNot(BeEmpty(), "No TARGET_ENV environment variable provided")

	// Read config for benchmark tests
	filename := fmt.Sprintf("../config/%s.yaml", env)
	yamlFile, err := ioutil.ReadFile(filename)
	Expect(err).Should(Succeed(), fmt.Sprintf("Failed reading benchmark configuration file: %s", filename))

	benchCfg = &config.Benchmark{}
	err = yaml.Unmarshal(yamlFile, benchCfg)
	Expect(err).Should(Succeed(), "Failed to marshal benchmark configuration file")

	// Create a client to collect metrics
	metricsClient, err = metrics.NewClient(benchCfg.Metrics.URL, 10*time.Second)
	Expect(err).Should(Succeed(), "Failed to create metrics client")

	// Create kubernetes client for deployments
	cfg, err := k8sconfig.GetConfig()
	Expect(err).Should(Succeed(), "Failed to read kubeconfig")

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	Expect(err).Should(Succeed(), "Failed to create new dynamic REST mapper")

	opts := client.Options{Scheme: scheme.Scheme, Mapper: mapper}
	k8sClient, err = client.New(cfg, opts)
	Expect(err).Should(Succeed(), "Failed to create new k8s client")
})
