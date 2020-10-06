package config

import (
	"fmt"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/prometheus/common/model"
	"k8s.io/client-go/kubernetes/scheme"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client interface {
	Name() string
}

type K8sClient struct {
	Client k8sclient.Client
}

type LocalClient struct {
	Client *dockerclient.Client
}

func NewClient(name string) (Client, error) {
	switch name {
	case "k8s":
		cfg, err := k8sconfig.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("Failed to get kubeconfig")
		}
		mapper, err := apiutil.NewDynamicRESTMapper(cfg)
		if err != nil {
			return nil, fmt.Errorf("Failed to create new dynamic REST mapper")
		}
		opts := k8sclient.Options{Scheme: scheme.Scheme, Mapper: mapper}
		cli, err := k8sclient.New(cfg, opts)
		if err != nil {
			return nil, fmt.Errorf("Failed to create new k8s client")
		}
		return &K8sClient{
			Client: cli,
		}, nil
	case "docker":
		cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("Failed to create docker client: %v", err)
		}
		return &LocalClient{
			Client: cli,
		}, nil
	default:
		return nil, fmt.Errorf("Unsupported client type %s", name)
	}
}

func (*K8sClient) Name() string {
	return "k8s"
}

func (lc *LocalClient) Name() string {
	return "docker"
}

type Logger struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Image     string `yaml:"image"`
	TenantID  string `yaml:"tenantId"`
	NetworkID string `yaml:"networkId"`
	// TODO: ID is used only for removing docker container
	ID string
}

type Querier struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Image     string `yaml:"image"`
	TenantID  string `yaml:"tenantId"`
	NetworkID string `yaml:"networkId"`
	ID        string
}

type Metrics struct {
	URL  string            `yaml:"url"`
	Jobs map[string]string `yaml:"jobs"`
}

func (m *Metrics) DistributorJob() string {
	job, ok := m.Jobs["distributor"]
	if !ok {
		return ""
	}
	return job
}

func (m *Metrics) IngesterJob() string {
	job, ok := m.Jobs["ingester"]
	if !ok {
		return ""
	}
	return job
}

func (m *Metrics) QuerierJob() string {
	job, ok := m.Jobs["querier"]
	if !ok {
		return ""
	}
	return job
}

func (m *Metrics) QueryFrontendJob() string {
	job, ok := m.Jobs["queryFrontend"]
	if !ok {
		return ""
	}
	return job
}

type Loki struct {
	Distributor   string `yaml:"distributor"`
	QueryFrontend string `yaml:"queryFrontend"`
}

func (lc *Loki) PushURL() string {
	return fmt.Sprintf("%s/loki/api/v1/push", lc.Distributor)
}

func (lc *Loki) QueryURL() string {
	return fmt.Sprintf("%s/loki/api/v1/query_range", lc.QueryFrontend)
}

type Writers struct {
	Replicas   int32 `yaml:"replicas"`
	Throughput int32 `yaml:"throughput"`
}

type Readers struct {
	Replicas       int32   `yaml:"replicas"`
	Query          string  `yaml:"query"`
	StartThreshold float64 `yaml:"startThreshold"`
}

type Samples struct {
	Interval time.Duration  `yaml:"interval"`
	Range    model.Duration `yaml:"range"`
	Total    int            `yaml:"total"`
}

type HighVolumeWrites struct {
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type HighVolumeReads struct {
	Samples Samples  `yaml:"samples"`
	Readers *Readers `yaml:"readers,omitempty"`
	Writers *Writers `yaml:"writers,omitempty"`
}

type Scenarios struct {
	HighVolumeWrites HighVolumeWrites `yaml:"highVolumeWrites"`
	HighVolumeReads  HighVolumeReads  `yaml:"highVolumeReads"`
}

type Benchmark struct {
	Logger    *Logger    `yaml:"logger"`
	Querier   *Querier   `yaml:"querier"`
	Metrics   *Metrics   `yaml:"metrics"`
	Loki      *Loki      `yaml:"loki"`
	Scenarios *Scenarios `yaml:"scenarios"`
}
