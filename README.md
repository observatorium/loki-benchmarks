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

Use the `scenarios/benchmarks.yaml` file to add, modify, or remove configurations. Modify the `generator.yaml`, `metrics.yaml`, or `querier.yaml` in the prefered deployment method directory to change these soruces.

## Running Benchmarks

Use the `make run-rhobs-benchmarks` or `make run-operator-benchmarks` to execute the benchmark program with the RHOBS or operator deployment styles on OpenShift respectively. Upon successful completion, a JSON and XML file will be created in the `reports/date+time` directory with the results of the tests.

## Troubleshooting

During benchmark execution, use [hack/scripts/ocp-deploy-grafana.sh](hack/scripts/ocp-deploy-grafana.sh) to deploy grafna and connect to Loki as a datasource: 
- Use a web browser to access grafana UI. The URL, username and password are printed by the script 
- In the UI, under settings -> data-sources hit `Save & test` to verify that Loki data-source is connected and that there are no errors
- In explore tab change the data-source to `Loki` and use `{client="promtail"}` query to visualize log lines
- Use additional queries such as `rate({client="promtail"}[1m])` to verify the behaviour of Loki and the benchmark


