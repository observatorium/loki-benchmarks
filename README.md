# loki-benchmarks

[![observatorium](https://circleci.com/gh/observatorium/loki-benchmarks.svg?style=svg)](https://app.circleci.com/pipelines/github/observatorium/loki-benchmarks)

This project is a simple golang testing project which uses [Ginkgo](https://github.com/onsi/ginkgo) to create a benchmarking suite for [Loki](https://github.com/grafana/loki). These tests are designed to record and report resource and network metrics from the write (distributor, ingester, etc) and read (querier, query-frontend, ingester, etc) paths.

These benchmarks can be executed on a vanilla Kubernetes or OpenShift cluster. It supports three deployment methods of Loki: Observatorium, Red Hat Observability Service, and the Loki Operator.

## Prerequisites

* `kubectl`, `aws`
* Repositories:
  * Observatorium Deployments Only: [Observatorium](https://github.com/observatorium/observatorium)
  * Operator Deployments Only: [Loki Operator](https://github.com/grafana/loki/tree/main/operator)
  * Non-OpenShift Deployments Only; Optional: [Cadvisor](https://github.com/google/cadvisor)

* Notes
  * Clone git repositories into sibling directories to the `loki-benchmarks` one.
  * Recommended cluster size: `m4.16xlarge`

## Configuring Tests

To change the testing configuration, see the files in the [config](./config) directory.

Different scenarios can be customized under `config/benchmarks/scenarios/benchmarks`. Current the benchmarks support two testing scenarios:

* Ingestion path scenarios: [suppored configuration](./internal/config/config.go#L77), this test will generate X amount of logs throughout a 30 minute window that's supposed to represent a full day of log ingestion.
* Query path scenarios: [supported configuration](./internal/config/config.go#L103), the theory behind this test is to generate the amount of data that would be queried before it starts running the queries.

## Running Benchmarks

### Prerequisites

The `run-operator-benchmarks` expects the following two env vars to be set `LOKI_OPERATOR_REGISTRY_BASE` `LOKI_STORAGE_BUCKET`.
E.g

```shell
export LOKI_OPERATOR_REGISTRY_BASE=quay.io/myorg
export LOKI_STORAGE_BUCKET=myname-loki-benchmark-storage
```

### Steps

1. Use the `make run-rhobs-benchmarks` or `make run-operator-benchmarks` to execute the benchmark program with the RHOBS or operator deployment styles on OpenShift respectively.
Both commands will run all the scenarios under `config/benchmarks/scenarios/benchmarks`.
2. Upon successful completion of each scenario, a JSON and XML file will be created in the `reports/date+time/scenario_name` directory with the results of the tests.
3. Once all scenarios have been run we can run `python3 hack/scripts/generate_report.py $PATH_TO_SCENARIO_1 $PATH_TO_SCENARIO_2 $PATH_TO_SCENARIO_...` to compile a report that helps compare the different scenarios.
4. To share the report on gDoc you can run `python3 hack/scripts/create-gdoc.py $PATH_TO_THE_REPORT` this will generate a docx file that can then be shared.

## Troubleshooting

During benchmark execution, use [hack/scripts/ocp-deploy-grafana.sh](hack/scripts/ocp-deploy-grafana.sh) to deploy Grafana and connect to Loki as a datasource:

* Use a web browser to access grafana UI. The URL, username and password are printed by the script
* In the UI, under settings -> data-sources hit `Save & test` to verify that Loki data-source is connected and that there are no errors
* In explore tab change the data-source to `Loki` and use `{client="promtail"}` query to visualize log lines
* Use additional queries such as `rate({client="promtail"}[1m])` to verify the behaviour of Loki and the benchmark
