#!/bin/bash

set -eou pipefail

TARGET_ENV="${TARGET_ENV:-development}"

OBS_NS="${OBS_NS:-observatorium}"
OBS_LOKI_QF="${OBS_LOKI_QF:-observatorium-xyz-loki-query-frontend}"
OBS_LOKI_QR="${OBS_LOKI_QR:-observatorium-xyz-loki-querier}"
OBS_LOKI_DST="${OBS_LOKI_DST:-observatorium-xyz-loki-distributor}"
OBS_LOKI_ING="${OBS_LOKI_ING:-observatorium-xyz-loki-ingester}"

trap 'tear_down;kill $(jobs -p); exit 0' EXIT

tear_down() {
    if [[ "$TARGET_ENV" = "development" ]]; then
        echo -e "\nUndeploying observatorium dev manifests"
        undeploy_observatorium
    fi
}

deploy_observatorium() {
    pushd ../deployments || exit 1
    ./tests/e2e.sh kind
    ./tests/e2e.sh deploy
    popd
}

undeploy_observatorium() {
    pushd ../deployments || exit 1
    echo -e "\nUndeploying observatorium dev manifests"
    ./kind delete cluster
    popd
}

forward_ports() {
    echo -e "\nWaiting for available loki query frontend deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_QF" --timeout=300s

    echo -e "\nSetup port-forward '3100:3100' to loki query frontend"
    (
        $KUBECTL -n "$OBS_NS" port-forward "svc/$OBS_LOKI_QF-http" 3100:3100;
    ) &

    echo -e "\nWaiting for available loki distributor deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_DST" --timeout=300s

    echo -e "\nSetup port-forward '3101:3100' to loki distributor"
    (
        $KUBECTL -n "$OBS_NS" port-forward "svc/$OBS_LOKI_DST-http" 3101:3100;
    ) &

    echo -e "\nWaiting for available loki ingester deployment"
    $KUBECTL -n "$OBS_NS" rollout status "statefulsets/$OBS_LOKI_ING" --timeout=300s || $KUBECTL -n observatorium describe pod "$OBS_LOKI_ING-0"

    echo -e "\nSetup port-forward '3102:3100' to loki ingester"
    (
        $KUBECTL -n "$OBS_NS" port-forward "svc/$OBS_LOKI_ING-http" 3102:3100;
    ) &

    echo -e "\nWaiting for available querier deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_QR" --timeout=300s

    echo -e "\nSetup port-forward '3103:3100' to loki querier"
    (
        $KUBECTL -n "$OBS_NS" port-forward "svc/$OBS_LOKI_QR-http" 3103:3100;
    ) &
}

scrape_loki_metrics() {
    source .bingo/variables.env
    (
        $PROMETHEUS --log.level=warn --config.file=./config/prometheus/config.yaml --storage.tsdb.path="$(mktemp -d)";
    ) &
}

generate_report() {
    source .bingo/variables.env

    for f in $REPORT_DIR/*.gnuplot; do
        gnuplot -e "set term png; set output '$f.png'" "$f"
    done

    cp ./reports/README.template $REPORT_DIR/README.md
    sed -i "s/{{TARGET_ENV}}/$TARGET_ENV/i" $REPORT_DIR/README.md
    $EMBEDMD -w $REPORT_DIR/README.md
}


bench() {
    if [[ "$TARGET_ENV" = "development" ]]; then
        echo "Deploying observatorium dev manifests"
        deploy_observatorium
    fi

    echo -e "\nFoward ports to loki deployments"
    forward_ports

    echo -e "\n Scrape metrics from Loki deployments"
    scrape_loki_metrics

    source .bingo/variables.env

    echo -e "\nRun benchmarks"
    $GINKGO ./benchmarks

    echo -e "\nGenerate benchmark report"
    generate_report
}

bench

exit $?
