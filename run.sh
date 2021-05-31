#!/bin/bash

set -eou pipefail

TARGET_ENV="${TARGET_ENV:-development}"

OBS_NS="${OBS_NS:-observatorium}"
CADVISOR_NS="${OBS_NS:-cadvisor}"
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
    pushd ../observatorium || exit 1
    ./configuration/tests/e2e.sh kind
    ./configuration/tests/e2e.sh deploy
    popd
}

undeploy_observatorium() {
    pushd ../observatorium || exit 1
    echo -e "\nUndeploying observatorium dev manifests"
    ./configuration/kind delete cluster
    popd
}

forward_ports() {
    shopt -s extglob

    i=0
    j=0
    cp config/prometheus/config.template config/prometheus/config.yaml

    echo -e "\nWaiting for available loki query frontend deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_QF" --timeout=300s

    echo -e "\nSetup port-forwards to loki query frontend pods"
    qf_targets=""
    for name in $($KUBECTL -n "$OBS_NS" get pod -l "app.kubernetes.io/component=query-frontend" -o name); do
        echo -e "\nSetup port-forward 310$i:3100 to loki query frontend pod: $name"
        (
            $KUBECTL -n "$OBS_NS" port-forward "$name" 310$i:3100;
        ) &
        qf_targets="$qf_targets'localhost:310$i',"
        ((i=i+1))
    done
    sed -i "s/{{LOKI_QUERY_FRONTEND_TARGETS}}/${qf_targets%%+(,)}/i" config/prometheus/config.yaml

    echo -e "\nWaiting for available loki distributor deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_DST" --timeout=300s

    echo -e "\nSetup port-forwards to loki distributor pods"
    ds_targets=""
    for name in $($KUBECTL -n "$OBS_NS" get pod -l "app.kubernetes.io/component=distributor" -o name); do
        echo -e "\nSetup port-forward 310$i:3100 to loki distributor pod: $name"
        (

            $KUBECTL -n "$OBS_NS" port-forward "$name" 310$i:3100;
        ) &
        ds_targets="$ds_targets'localhost:310$i',"
        ((i=i+1))
    done
    sed -i "s/{{LOKI_DISTRIBUTOR_TARGETS}}/${ds_targets%%+(,)}/i" config/prometheus/config.yaml

    echo -e "\nWaiting for available loki ingester deployment"
    $KUBECTL -n "$OBS_NS" rollout status "statefulsets/$OBS_LOKI_ING" --timeout=300s

    echo -e "\nSetup port-forwards to loki ingester pods"
    in_targets=""
    for name in $($KUBECTL -n "$OBS_NS" get pod -l "app.kubernetes.io/component=ingester" -o name); do
        echo -e "\nSetup port-forward 310$i:3100 to loki ingester pod: $name"
        (
            $KUBECTL -n "$OBS_NS" port-forward "$name" 310$i:3100;
        ) &
        in_targets="$in_targets'localhost:310$i',"
        ((i=i+1))
    done
    sed -i "s/{{LOKI_INGESTER_TARGETS}}/${in_targets%%+(,)}/i" config/prometheus/config.yaml

    echo -e "\nWaiting for available querier deployment"
    $KUBECTL -n "$OBS_NS" rollout status "statefulsets/$OBS_LOKI_QR" --timeout=300s

    echo -e "\nSetup port-forwards to loki querier pods"
    qr_targets=""
    for name in $($KUBECTL -n "$OBS_NS" get pod -l "app.kubernetes.io/component=querier" -o name); do
        echo -e "\nSetup port-forward 310$i:3100 to loki querier pod: $name"
        (
            $KUBECTL -n "$OBS_NS" port-forward "$name" 310$i:3100;
        ) &
        qr_targets="$qr_targets'localhost:310$i',"
        ((i=i+1))
    done
    sed -i "s/{{LOKI_QUERIER_TARGETS}}/${qr_targets%%+(,)}/i" config/prometheus/config.yaml

    PROJECT_CADVISOR=$($KUBECTL get ns | grep "$CADVISOR_NS")
    qr_targets=""
    if [ -n "$PROJECT_CADVISOR" ]; then
      echo -e "\nSetup port-forwards to cadvisor pods"
      for name in $($KUBECTL -n "$CADVISOR_NS" get pod -o name); do
          echo -e "\nSetup port-forward 808$j:8080 to cadvisor pod: $name"
          (
              $KUBECTL -n "$CADVISOR_NS" port-forward "$name" 808$j:8080;
          ) &
          qr_targets="$qr_targets'localhost:808$j',"
          ((j=j+1))
      done
    fi
    sed -i "s/{{CADVISOR_TARGETS}}/${qr_targets%%+(,)}/i" config/prometheus/config.yaml
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

    echo -e "\nScrape metrics from Loki deployments"
    scrape_loki_metrics

    source .bingo/variables.env

    echo -e "\nRun benchmarks"
    $GINKGO -v ./benchmarks

    echo -e "\nGenerate benchmark report"
    generate_report
}

bench

exit $?
