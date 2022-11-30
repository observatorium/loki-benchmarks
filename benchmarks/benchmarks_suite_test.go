package benchmarks_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/observatorium/loki-benchmarks/internal/config"
	"github.com/observatorium/loki-benchmarks/internal/metrics"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	benchCfg      *config.Benchmark
	k8sClient     client.Client
	metricsClient *metrics.Client

	defaultRetry   = 5 * time.Second
	defaultRange   = 5 * time.Minute
	defaultTimeout = 1 * time.Minute
)

func init() {
	// Read target environment
	configDir := os.Getenv("BENCHMARKING_CONFIGURATION_DIRECTORY")
	if configDir == "" {
		panic("Missing BENCHMARKING_CONFIGURATION_DIRECTORY env variable")
	}

	filename := fmt.Sprintf("../config/benchmarks/%s/benchmark.yaml", configDir)
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed reading benchmark configuration file: %s", filename))
	}

	benchCfg = &config.Benchmark{}
	err = yaml.Unmarshal(yamlFile, benchCfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal benchmark configuration file %s with errors %v", filename, err))
	}

	// Create K8s Client
	cfg := k8sconfig.GetConfigOrDie()
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		panic("Failed to create new dynamic REST mapper")
	}

	opts := client.Options{Scheme: scheme.Scheme, Mapper: mapper}
	k8sClient, err = client.New(cfg, opts)
	if err != nil {
		panic("Failed to create new k8s client")
	}

	// Create Metrics Client
	promToken := os.Getenv("PROMETHEUS_TOKEN")
	metricsClient, err = metrics.NewClient(benchCfg.Metrics.URL, promToken, 30*time.Second)
	if err != nil {
		panic("Failed to create metrics client")
	}

	fmt.Printf("\nUsing benchmark configuration:\n===============================\n%s\n", yamlFile)
}

func TestBenchmarks(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Benchmarks Suite")
}
