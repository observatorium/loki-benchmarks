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

* Checkout a copy of the [observatorium/observatorium](https://github.com/observatorium/observatorium) repository and place it as a sibling directory to the `loki-benchmarks` repository.  
* Working [cadvisor](https://github.com/google/cadvisor) is required to accurately measure CPU and Memory. Deployment script available at [hack/deploy_cadvisor.sh](hack/deploy_cadvisor.sh)  
* gnuplot is required to create graph reports. To install execute: `sudo yum install gnuplot`

## Deployment and benchmark on OCP cluster 

* Connect to OCP cluster and make sure `oc` command is working  
* From within `hack` folder execute `deploy_cadvisor.sh` to deploy cadvisor
* From within `hack` folder execute `deploy_loki_using_observatorium_yaml.sh` to deploy Loki  
* Start the benchmark using `bench-obs-logs-test` configuration `make bench-obs-logs-test`
* Upon benchmark execution completion, results are available in the `reports/date+time` folder 

## How to add new benchmarks to this suite

### Developing

* Declare a new scenario with expected measurement values for each profile in the [config](./config) directory.
* Extend the golang `Scenarios` struct in [internal/config/config.go](./internal/config/config.go) with the new scenario.
* Add a new `_test.go` file in the [benchmarks](./benchmarks) directory.

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
