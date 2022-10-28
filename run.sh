#!/bin/bash

set -eou pipefail

source .bingo/variables.env

IS_TESTING="${IS_TESTING:-false}"
IS_OPENSHIFT="${IS_OPENSHIFT:-false}"
USE_CADVISOR="${USE_CADVISOR:-false}"

BENCHMARK_NAMESPACE="${BENCHMARK_NAMESPACE:-observatorium}"
LOKI_COMPONENT_PREFIX="${LOKI_COMPONENT_PREFIX:-observatorium-xyz-loki}"

SCENARIO_CONFIGURATION_FILE="${SCENARIO_CONFIGURATION_FILE:-benchmarks.yaml}"
BENCHMARKING_CONFIGURATION_DIRECTORY="${BENCHMARKING_CONFIGURATION_DIRECTORY:-observatorium}"

benchmarking_configuration_path="config/benchmarks/$BENCHMARKING_CONFIGURATION_DIRECTORY"
benchmarking_configuration_file="config/benchmarks/$BENCHMARKING_CONFIGURATION_DIRECTORY/benchmark.yaml"
ocp_prometheus_config_path="config/openshift"

port_counter=0

# Deploy Loki with the Observatorium configuration
# Intended to work on a Kind cluster for quick development
observatorium() {
    output_directory=$1
    
    create_benchmarking_environment
    
    pushd ../observatorium || exit 1
    KUBECTL=$(which kubectl) ./configuration/tests/e2e.sh deploy
    popd
    
    wait_for_ready_components
    configure_prometheus
    run_benchmark_suite $output_directory
    
    # Clean up
    destroy_benchmarking_environment
}

# Deploy Loki with the Red Hat Observability Service configuration
# Intended to work on a OpenShift cluster for benchmarking
rhobs() {
    output_directory=$1
    rhobs_loki_deployment_file=$2
    storage_bucket=$3

    create_benchmarking_environment
    create_s3_storage $storage_bucket

    kubectl -n $BENCHMARK_NAMESPACE apply -f $rhobs_loki_deployment_file
    ./hack/scripts/deploy-example-secret.sh $BENCHMARK_NAMESPACE $storage_bucket

    wait_for_ready_components
    configure_prometheus
    run_benchmark_suite $output_directory

    # Clean Up
    echo -e "\nRemoving RHOBS configuration file"
    rm $rhobs_loki_deployment_file

    destroy_benchmarking_environment
    destory_s3_storage $storage_bucket
}

# Deploy Loki with the Red Hat Loki Operator
# Intended to work on a OpenShift cluster for benchmarking
operator() {
    output_directory=$1
    operator_registry=$2
    storage_bucket=$3

    if $IS_OPENSHIFT; then
        echo -e "\nOverwriting BENCHMARK_NAMESPACE"
        BENCHMARK_NAMESPACE="openshift-logging"
    fi

    # Create namespaces and storage
    create_benchmarking_environment
    create_s3_storage $storage_bucket
    
    # Deploy operator and Lokistack
    pushd ../loki/operator || exit 1
    if $IS_OPENSHIFT; then
        kubectl create namespace openshift-operators-redhat
        kubectl label ns/$BENCHMARK_NAMESPACE openshift.io/cluster-monitoring=true --overwrite 

        make olm-deploy REGISTRY_ORG=$operator_registry VERSION=v0.0.1
        ./hack/deploy-aws-storage-secret.sh $storage_bucket
        kubectl -n $BENCHMARK_NAMESPACE apply -f hack/lokistack_gateway_ocp.yaml
    else
        make deploy REGISTRY_ORG=$operator_registry VERSION=v0.0.1
        kubectl -n $BENCHMARK_NAMESPACE apply -f hack/lokistack_gateway_dev.yaml
    fi
    popd

    kubectl -n $BENCHMARK_NAMESPACE apply -f hack/loadclient-rbac.yaml

    wait_for_ready_components
    configure_prometheus
    run_benchmark_suite $output_directory

    # Clean Up
    kubectl -n $BENCHMARK_NAMESPACE delete -f hack/loadclient-rbac.yaml

    if $IS_OPENSHIFT; then
        kubectl delete namespace openshift-operators-redhat --ignore-not-found=true
    fi

    destroy_benchmarking_environment
    destory_s3_storage $storage_bucket
}

create_benchmarking_environment() {
    echo -e "\nCreating benchmarking file"

    export BENCHMARKING_CONFIGURATION_DIRECTORY

    cat $benchmarking_configuration_path/generator.yaml > $benchmarking_configuration_file
    cat $benchmarking_configuration_path/querier.yaml >> $benchmarking_configuration_file
    cat $benchmarking_configuration_path/metrics.yaml >> $benchmarking_configuration_file
    cat config/benchmarks/scenarios/$SCENARIO_CONFIGURATION_FILE >> $benchmarking_configuration_file

    echo -e "\nCreating benchmarking environment"

    if $IS_TESTING; then
        $KIND create cluster
    fi

    kubectl create namespace $BENCHMARK_NAMESPACE

    if $USE_CADVISOR; then
         pushd ../cadvisor || exit 1
         kubectl $KUSTOMIZE deploy/kubernetes/base | kubectl apply -f -
         popd
    fi
}

destroy_benchmarking_environment() {
    echo -e "\nRemoving benchmarking file"

    rm $benchmarking_configuration_file

    echo -e "\nDestroying benchmarking environment"

    if $USE_CADVISOR; then
         pushd ../cadvisor || exit 1
         kubectl $KUSTOMIZE deploy/kubernetes/base | kubectl delete -f -
         popd
    fi

    if $IS_OPENSHIFT; then
        disable_ocp_prometheus_monitoring
    fi

    if $IS_TESTING; then
        $KIND delete cluster
    else
        kubectl delete namespace $BENCHMARK_NAMESPACE --ignore-not-found=true
    fi
}

create_s3_storage() {
    bucket_names=$1

    if $IS_OPENSHIFT; then
        echo -e "\nCreating AWS S3 storage"
        ./hack/scripts/create-s3-bucket.sh $bucket_names
    fi
}

destory_s3_storage() {
    bucket_names=$1

    if $IS_OPENSHIFT; then
        echo -e "\nDestroying AWS S3 storage"
        ./hack/scripts/delete-s3-bucket.sh $bucket_names
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
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-querier" --timeout=600s

    echo -e "\nWaiting for available loki query frontend deployment"
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-query-frontend" --timeout=600s

    echo -e "\nWaiting for available loki distributor deployment"
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "deploy/$LOKI_COMPONENT_PREFIX-distributor" --timeout=600s

    echo -e "\nWaiting for available loki ingester statefulset"
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-ingester" --timeout=600s

    echo -e "\nWaiting for available loki index gateway statefulset"
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-index-gateway" --timeout=600s

    echo -e "\nWaiting for available loki compactor statefulset"
    kubectl -n "$BENCHMARK_NAMESPACE" rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-compactor" --timeout=600s
}

run_benchmark_suite() {
    output_directory=$1
    
    $GINKGO --json-report=report.json -output-dir=$output_directory ./benchmarks
}

enable_ocp_prometheus_monitoring() {
    echo -e "\nAdding user workload monitoring configuration"

    kubectl -n openshift-monitoring apply -f $ocp_prometheus_config_path/cluster-monitoring-config.yaml
	kubectl -n openshift-user-workload-monitoring apply -f $ocp_prometheus_config_path/user-workload-monitoring-config.yaml
}

disable_ocp_prometheus_monitoring() {
    echo -e "\nRemoving user workload monitoring configuration"

    kubectl -n openshift-monitoring delete -f $ocp_prometheus_config_path/cluster-monitoring-config.yaml --ignore-not-found=true
	kubectl -n openshift-user-workload-monitoring delete  -f $ocp_prometheus_config_path/user-workload-monitoring-config.yaml --ignore-not-found=true
}

export_ocp_prometheus_settings() {
    echo -e "\nRetrieving Prometheus URL and bearer token"

    secret=$(kubectl -n openshift-user-workload-monitoring get secret | grep prometheus-user-workload-token | head -n 1 | awk '{print $1 }')
    export PROMETHEUS_URL="https://$(kubectl -n openshift-monitoring get route thanos-querier -o json | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["spec"]["host"])')"
    export PROMETHEUS_TOKEN=$(kubectl -n openshift-user-workload-monitoring get secret $secret -o json | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["data"]["token"])' | base64 -d)
}

forward_ports() {
    shopt -s extglob

    cp config/prometheus/config.template config/prometheus/config.yaml

    setup_ports "loki query frontend" app.kubernetes.io/component=query-frontend LOKI_QUERY_FRONTEND_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki distributor" app.kubernetes.io/component=distributor LOKI_DISTRIBUTOR_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki ingester" app.kubernetes.io/component=ingester LOKI_INGESTER_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "loki querier" app.kubernetes.io/component=querier LOKI_QUERIER_TARGETS 3100 "$BENCHMARK_NAMESPACE"
    setup_ports "cadvisor ingesters" "" CADVISOR_INGESTERS_TARGETS 8080 "$BENCHMARK_NAMESPACE"
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
    INGESTERS_REGEX=$(kubectl get pods -l "app.kubernetes.io/component=ingester" -n "$BENCHMARK_NAMESPACE" -o jsonpath='{range .items[*]}{".*crio-"}{.status.containerStatuses[?(@.name=="observatorium-loki-ingester")].containerID}{".*|"}{end}' | sed -s 's|cri-o://||g')
    INGESTERS_REGEX=${INGESTERS_REGEX%%+(\|)}
    sed -i "s/{{CADVISOR_INGESTERS_TARGETS_PODS}}/$INGESTERS_REGEX/i" config/prometheus/config.yaml
}

scrape_loki_metrics() {
    (
        $PROMETHEUS --log.level=warn --config.file=./config/prometheus/config.yaml --storage.tsdb.path="$(mktemp -d)";
    ) &
}

case $1 in
observatorium)
    observatorium $2
    ;;

rhobs)
    rhobs $2 $3 $4
    ;;

operator)
    operator $2 $3 $4
    ;;

help)
    echo "usage: $(basename "$0") { observatorium | rhobs | operator }"
    ;;

*)
    observatorium
    rhobs
    loki_operator
    ;;
esac