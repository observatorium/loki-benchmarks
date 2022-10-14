#!/bin/bash

set -eou pipefail

source .bingo/variables.env

IS_TESTING="${IS_TESTING:-false}"
IS_OPENSHIFT="${IS_OPENSHIFT:-false}"

CADVISOR_NAMESPACE="${CADVISOR_NAMESPACE:-cadvisor}"
BENCHMARK_NAMESPACE="${BENCHMARK_NAMESPACE:-observatorium}"
LOKI_COMPONENT_PREFIX="${$LOKI_COMPONENT_PREFIX:-observatorium-xyz-loki}"

ocp_prometheus_config_directory=config/ocp-monitoring

port_counter=0

# Deploy Loki with the Observatorium configuration
# Intended to work on a Kind cluster for quick development
deploy_loki_with_observatorium_configuration() {
    REPORT_DIR=$1

    create_benchmarking_environment

    pushd ../observatorium || exit 1
    ./configuration/tests/e2e.sh deploy
    popd

    wait_for_ready_components

    configure_prometheus

    # Run benchmarks
    $GINKGO --json-report=report.json -output-dir=$REPORT_DIR ./benchmarks
    
    # Clean up
    destroy_benchmarking_environment
}

# Deploy Loki with the Red Hat Observability Service configuration
# Intended to work on a OpenShift cluster for benchmarking
deploy_loki_with_rhobs_configuration() {
    REPORT_DIR=$1
    RHOBS_LOKI_FILE=$2
    LOKI_STORAGE_BUCKET=$3

    create_benchmarking_environment

    kubectl -n $BENCHMARK_NAMESPACE apply -f $RHOBS_LOKI_FILE
    ./hack/deploy-example-secret.sh $BENCHMARK_NAMESPACE $LOKI_STORAGE_BUCKET

    wait_for_ready_components

    configure_prometheus

    # Run benchmarks
    $GINKGO --json-report=report.json -output-dir=$REPORT_DIR ./benchmarks

    # Clean Up
    destroy_benchmarking_environment

    echo -e "\nRemoving RHOBS configuration file"
    rm $RHOBS_LOKI_FILE
}

# Deploy Loki with the Red Hat Loki Operator
# Intended to work on a OpenShift cluster for benchmarking
deploy_loki_with_red_hat_operator() {
    REPORT_DIR=$1
    LOKI_OPERATOR_REGISTRY=$2
    LOKI_STORAGE_BUCKET=$3

    # Create namespaces
    create_benchmarking_environment
    kubectl create namespace openshift-operators-redhat
    kubectl ns/$BENCHMARK_NAMESPACE openshift.io/cluster-monitoring=true --overwrite

    # Deploy operator and Lokistack
    pushd ../loki/operator || exit 1
    make olm-deploy REGISTRY_ORG=$LOKI_OPERATOR_REGISTRY VERSION=v0.0.1
    ./hack/deploy-aws-storage-secret.sh $LOKI_STORAGE_BUCKET
    kubectl -n $BENCHMARK_NAMESPACE apply -f hack/lokistack_gateway_ocp.yaml
    popd

    wait_for_ready_components

    configure_prometheus

    # Run benchmarks
    $GINKGO --json-report=report.json -output-dir=$REPORT_DIR ./benchmarks

    # Clean Up
    kubectl delete namespace openshift-operators-redhat --ignore-not-found=true
    destroy_benchmarking_environment
}

create_benchmarking_environment() {
    echo -e "\nCreating benchmarking environment"

    if $IS_TESTING; then
        $KIND create cluster
    fi

    kubectl create namespace $BENCHMARK_NAMESPACE
}

destroy_benchmarking_environment() {
    echo -e "\nDestroying benchmarking environment"

    if $IS_TESTING; then
        $KIND delete cluster
    else
        kubectl delete namespace $BENCHMARK_NAMESPACE --ignore-not-found=true
    fi

    if $IS_OPENSHIFT; then
        disable_ocp_prometheus_monitoring
    fi
}

configure_prometheus() {
    echo -e "\nConfiguring Prometheus"

    if $IS_OPENSHIFT; then
        enable_ocp_prometheus_monitoring
        export_ocp_prometheus_settings
    else
        echo -e "\nForward ports to loki components"
        forward_ports

        echo -e "\nSet prometheus relabel regex"
        set_prometheus_relabel_regex

        echo -e "\nScrape metrics from Loki components"
        scrape_loki_metrics
    fi
}

wait_for_ready_components() {
    echo -e "\nWaiting for available querier deployment"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-querier" --timeout=600s

    echo -e "\nWaiting for available loki query frontend deployment"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-query-frontend" --timeout=600s

    echo -e "\nWaiting for available loki distributor deployment"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-distributor" --timeout=600s

    echo -e "\nWaiting for available loki ingester statefulset"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-ingester" --timeout=600s

    echo -e "\nWaiting for available loki index gateway statefulset"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-index-gateway" --timeout=600s

    echo -e "\nWaiting for available loki compactor statefulset"
    $KUBECTL -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-compactor" --timeout=600s
}

enable_ocp_prometheus_monitoring() {
    echo -e "\nAdding user workload monitoring configuration"

    kubectl -n openshift-monitoring apply -f $ocp_prometheus_config_directory/cluster-monitoring-config.yaml
	kubectl -n openshift-user-workload-monitoring apply -f $ocp_prometheus_config_directory/user-workload-monitoring-config.yaml
}

disable_ocp_prometheus_monitoring() {
    echo -e "\nRemoving user workload monitoring configuration"

    kubectl -n openshift-monitoring delete -f $ocp_prometheus_config_directory/cluster-monitoring-config.yaml --ignore-not-found=true
	kubectl -n openshift-user-workload-monitoring delete  -f $ocp_prometheus_config_directory/user-workload-monitoring-config.yaml --ignore-not-found=true
}

export_ocp_prometheus_settings() {
    echo -e "\nRetrieving Prometheus URL and bearer token"

    secret=$(kubectl -n openshift-user-workload-monitoring get secret | grep prometheus-user-workload-token | head -n 1 | awk '{print $1 }')
    export PROMETHEUS_URL="https://$(kubectl -n openshift-monitoring get route thanos-querier -o json | jq -r '.spec.host')"
    export PROMETHEUS_TOKEN=$(kubectl -n openshift-user-workload-monitoring get secret $secret -o json | jq -r '.data.token' | base64 -d)
}

forward_ports() {
    shopt -s extglob

    cp config/prometheus/config.template config/prometheus/config.yaml

    setup_ports "loki query frontend" app.kubernetes.io/component=query-frontend LOKI_QUERY_FRONTEND_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki distributor" app.kubernetes.io/component=distributor LOKI_DISTRIBUTOR_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki ingester" app.kubernetes.io/component=ingester LOKI_INGESTER_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki querier" app.kubernetes.io/component=querier LOKI_QUERIER_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "cadvisor ingesters" "" CADVISOR_INGESTERS_TARGETS 8080 "$CADVISOR_NAMESPACE"
}

setup_ports() {
    setup_name=$1
    match_label=$2
    prometheus_template_targets=$3
    source_port=$4
    namespace=$5

    qr_targets=""

    echo -e "\nSetup port-forwards to $setup_name pods in namespace $namespace"
    
    for name in $(kubectl -n "$namespace" get pod -l "$match_label" -o name); do
        destination_port=$((source_port+port_counter))
        echo -e "\nSetup port-forward $destination_port:$source_port to $setup_name pod: $name"
        (
            kubectl -n "$namespace" port-forward "$name" $destination_port:"$source_port";
        ) &
        qr_targets="$qr_targets'localhost:$destination_port',"
        ((port_counter=port_counter+1))
    done
    
    sed -i "s/{{$prometheus_template_targets}}/${qr_targets%%+(,)}/i" config/prometheus/config.yaml
}

set_prometheus_relabel_regex() {
    INGESTERS_REGEX=$($KUBECTL get pods -l "app.kubernetes.io/component=ingester" -n "$BENCHMARK_NAMESPACE" -o jsonpath='{range .items[*]}{".*crio-"}{.status.containerStatuses[?(@.name=="observatorium-loki-ingester")].containerID}{".*|"}{end}' | sed -s 's|cri-o://||g')
    INGESTERS_REGEX=${INGESTERS_REGEX%%+(\|)}
    sed -i "s/{{CADVISOR_INGESTERS_TARGETS_PODS}}/$INGESTERS_REGEX/i" config/prometheus/config.yaml
}

scrape_loki_metrics() {
    (
        $PROMETHEUS --log.level=warn --config.file=./config/prometheus/config.yaml --storage.tsdb.path="$(mktemp -d)";
    ) &
}