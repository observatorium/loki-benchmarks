# loki-benchmarks

[![observatorium](https://circleci.com/gh/observatorium/loki-benchmarks.svg?style=svg)](https://app.circleci.com/pipelines/github/observatorium/loki-benchmarks)

This suite consists of Loki benchmarks tests for multiple scenarios. Each scenario asserts recorded measurements against a selected profile from the [config](./config) directory:

1. **Write benchmarks**:
   - High Volume Writes: Measure `CPU`, `MEM` and `QPS`, `p99`, `p50` `avg` request duration for all 2xx write requests to all Loki distributor and ingester pods.

2. **Read benchmarks**:
   - High Volume Reads: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to all Loki query-frontend, querier and ingester pods.
   - High Volume Aggregate: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to all Loki query-frontend, querier and ingester pods.
   - High Volume Aggregate: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to all Loki query-frontend, querier and ingester pods.
   - Dashboard queries: Measure `QPS`, `p99`, `p50` and `avg` request duration for all 2xx read requests to all Loki query-frontend, querier and ingester pods.

## Prerequisites

* Software: `gnuplot`
  
Note: Install on Linux environment, e.g. on Fedora using: `sudo dnf install gnuplot`

### Kubernetes

* Required software: `kubectl`
* Repositories:
  * [Observatorium](https://github.com/observatorium/observatorium)
  * Optional: [Cadvisor](https://github.com/google/cadvisor)

Note: Clone git repositories into sibling directories to the `loki-benchmarks` one.   

Note: Cadvisor is only required if measuring CPU and memory of the container. In addition, change the value of the `enableCadvisorMetrics` key in the configuration to be `true`. It is `false` by default.

#### Deployment

1. Configure the parameters (`config/loki-parameters`) and deploy Loki & configure Prometheus: `make deploy-obs-loki`
2. Run the benchmarks: `make bench-dev`

### OCP AWS Cluster

* Required software: `oc`, `aws`
* Cluster Size: `m4.16xlarge`

#### Deployment

1. Configure benchmark parameters `config/loki-parameters`
1. Create S3 bucket: `make deploy-s3-bucket`
1. Deploy prometheus `make deploy-ocp-prometheus`
1. Download loki observatorium template locally `make download-obs-loki-template`
1. Deploy Loki `make deploy-ocp-loki`
1. Run the benchmarks: `make ocp-run-benchmarks`

Note: For additional details and all-in-one commands use: `make help`

Upon benchmark execution completion, results are available in the `reports/date+time` folder.

Uninstall using: `make ocp-all-cleanup`.


## How to add new benchmarks to this suite

### Developing

* Declare a new scenario with expected measurement values for each profile in the [config](./config) directory.
* Extend the golang `Scenarios` struct in [internal/config/config.go](./internal/config/config.go) with the new scenario.
* Add a new `_test.go` file in the [benchmarks](./benchmarks) directory.
* When using [`cluster-logging-load-client`](quay.io/openshift-logging/cluster-logging-load-client:latest) as logger,
  the `command` configuration parameter is either **generate** or **query** and  
  all other `args` configuration parameters are described in [https://github.com/ViaQ/cluster-logging-load-client](https://github.com/ViaQ/cluster-logging-load-client)
* Overriding `url` and `tenant` requires that the logger implementation provides such named CLI flags

### Run the tests

```
$ make bench-dev
```

Example output:
```
Running Suite: Benchmarks Suite
===============================
Random Seed: 1597237201
Will run 1 of 1 specs

• [MEASUREMENT]
Scenario: High Volume Writes
/home/username/dev/loki-benchmarks/benchmarks/high_volume_writes_test.go:18
  should result in measurements of p99, p50 and avg for all successful write requests to the distributor
  /home/username/dev/loki-benchmarks/benchmarks/high_volume_writes_test.go:32

  Ran 10 samples:
  All distributor 2xx Writes p99:
    Smallest: 0.087
     Largest: 0.096
     Average: 0.092 ± 0.003
  All distributor 2xx Writes p50:
    Smallest: 0.003
     Largest: 0.003
     Average: 0.003 ± 0.000
  All distributor 2xx Writes avg:
    Smallest: 0.370
     Largest: 0.594
     Average: 0.498 ± 0.085
------------------------------
```

### Inspecting the benchmark report

On each run a new time-based report directory is created under the [reports](./reports) directory. Each report includes:
* Summary `README.md` with all benchmark measurements.
* A CSV file for each specific measurement.
* A GNUPlot file for each specific measurement to transform the data into a PNG graph.

Example output:
```
reports
├── 2020-08-12-10-33-31
   ├── All-distributor-2xx-Writes-avg.csv
   ├── All-distributor-2xx-Writes-avg.gnuplot
   ├── All-distributor-2xx-Writes-avg.gnuplot.png
   ├── All-distributor-2xx-Writes-p50.csv
   ├── All-distributor-2xx-Writes-p50.gnuplot
   ├── All-distributor-2xx-Writes-p50.gnuplot.png
   ├── All-distributor-2xx-Writes-p99.csv
   ├── All-distributor-2xx-Writes-p99.gnuplot
   ├── All-distributor-2xx-Writes-p99.gnuplot.png
   ├── junit.xml
   └── README.md
```
