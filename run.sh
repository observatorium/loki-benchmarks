#!/bin/bash

set -eou pipefail

TARGET_ENV="${TARGET_ENV:-development}"

OBS_NS="${OBS_NS:-observatorium}"
CADVISOR_NS="${CADVISOR_NS:-cadvisor}"
OBS_LOKI_QF="${OBS_LOKI_QF:-observatorium-xyz-loki-query-frontend}"
OBS_LOKI_QR="${OBS_LOKI_QR:-observatorium-xyz-loki-querier}"
OBS_LOKI_DST="${OBS_LOKI_DST:-observatorium-xyz-loki-distributor}"
OBS_LOKI_ING="${OBS_LOKI_ING:-observatorium-xyz-loki-ingester}"
port_counter=0

trap 'kill $(jobs -p); tear_down; exit 0' EXIT

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
    ./kind delete cluster
    popd
}

setup_ports() {
    setup_name=$1
    match_label=$2
    prometheus_template_targets=$3
    source_port=$4
    namespace=$5
    echo -e "\nSetup port-forwards to $setup_name pods in namespace $namespace"
    qr_targets=""
    for name in $($KUBECTL -n "$namespace" get pod -l "$match_label" -o name); do
        destination_port=$((source_port+port_counter))
        echo -e "\nSetup port-forward $destination_port:$source_port to $setup_name pod: $name"
        (
            $KUBECTL -n "$namespace" port-forward "$name" $destination_port:"$source_port";
        ) &
        qr_targets="$qr_targets'localhost:$destination_port',"
        ((port_counter=port_counter+1))
    done
    sed -i "s/{{$prometheus_template_targets}}/${qr_targets%%+(,)}/i" config/prometheus/config.yaml
}

forward_ports() {
    shopt -s extglob
    cp config/prometheus/config.template config/prometheus/config.yaml

    echo -e "\nWaiting for available loki query frontend deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_QF" --timeout=300s
    echo -e "\nWaiting for available loki distributor deployment"
    $KUBECTL -n "$OBS_NS" rollout status "deploy/$OBS_LOKI_DST" --timeout=300s
    echo -e "\nWaiting for available loki ingester deployment"
    $KUBECTL -n "$OBS_NS" rollout status "statefulsets/$OBS_LOKI_ING" --timeout=300s
    echo -e "\nWaiting for available querier deployment"
    $KUBECTL -n "$OBS_NS" rollout status "statefulsets/$OBS_LOKI_QR" --timeout=300s

    setup_ports "loki query frontend" app.kubernetes.io/component=query-frontend LOKI_QUERY_FRONTEND_TARGETS 3100 "$OBS_NS"
    setup_ports "loki distributor" app.kubernetes.io/component=distributor LOKI_DISTRIBUTOR_TARGETS 3100 "$OBS_NS"
    setup_ports "loki ingester" app.kubernetes.io/component=ingester LOKI_INGESTER_TARGETS 3100 "$OBS_NS"
    setup_ports "loki querier" app.kubernetes.io/component=querier LOKI_QUERIER_TARGETS 3100 "$OBS_NS"
    setup_ports "cadvisor ingesters" "" CADVISOR_INGESTERS_TARGETS 8080 "$CADVISOR_NS"
}

set_prometheus_relabel_regex() {
    INGESTERS_REGEX=$($KUBECTL get pods -l "app.kubernetes.io/component=ingester" -n "$CADVISOR_NS" -o jsonpath='{range .items[*]}{".*crio-"}{.status.containerStatuses[?(@.name=="observatorium-loki-ingester")].containerID}{".*|"}{end}' | sed -s 's|cri-o://||g')
    INGESTERS_REGEX=${INGESTERS_REGEX%%+(\|)}
    sed -i "s/{{CADVISOR_INGESTERS_TARGETS_PODS}}/$INGESTERS_REGEX/i" config/prometheus/config.yaml
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

    echo -e "\nForward ports to loki deployments"
    forward_ports

    echo -e "\nSet prometheus relabel regex"
    set_prometheus_relabel_regex

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
