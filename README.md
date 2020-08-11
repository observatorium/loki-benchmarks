# loki-benchmarks

[![observatorium](https://circleci.com/gh/observatorium/loki-benchmarks.svg?style=svg)](https://app.circleci.com/pipelines/github/observatorium/loki-benchmarks)

This suite consists of benchmarks tests for the following Loki scenarios:

1. **High Volume Writes**: Maintain p99 write request duration to 300ms where incoming volume is produced by 10 pods each sending 100 messages/s.

## How to run the suite

### Prerequisites

* Checkout a copy of the [observatorium/deployments](https://github.com/observatorium/deployments) repository and place it as a sibling directory to the `loki-benchmarks` repository.

### Run the tests

```
$ make bench-dev
```

Example output:
```
Run benchmarks
Running Suite: Benchmarks Suite
===============================
Random Seed: 1597077883
Will run 1 of 1 specs

• [MEASUREMENT]
High Volume Writes
/home/username/dev/loki-benchmarks/benchmarks/high_volume_writes_test.go:22
  should result in a p99 <= 300ms for all successful requests
  /home/username/dev/loki-benchmarks/benchmarks/high_volume_writes_test.go:36

  Ran 10 samples:

  All distributor 2xx Writes p99:
    Smallest: 0.081
     Largest: 0.090
     Average: 0.087 ± 0.002
------------------------------

```
