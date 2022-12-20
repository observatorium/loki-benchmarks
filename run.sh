#!/bin/bash

set -eou pipefail

source .bingo/variables.env

IS_TESTING="${IS_TESTING:-false}"
IS_OPENSHIFT="${IS_OPENSHIFT:-false}"
USE_CADVISOR="${USE_CADVISOR:-false}"

BENCHMARK_NAMESPACE="${BENCHMARK_NAMESPACE:-observatorium}"
LOKI_COMPONENT_PREFIX="${LOKI_COMPONENT_PREFIX:-observatorium-xyz-loki}"

OUTPUT_DIRECTORY="${OUTPUT_DIRECTORY:-reports/$(date +%Y-%m-%d-%H-%M-%S)}"
SCENARIO_CONFIGURATION_DIRECTORY="${SCENARIO_CONFIGURATION_DIRECTORY:-benchmarks}"
BENCHMARKING_CONFIGURATION_DIRECTORY="${BENCHMARKING_CONFIGURATION_DIRECTORY:-observatorium}"

PROMETHEUS_CLIENT_PROTOCOL="http"
PROMETHEUS_CLIENT_URL="${PROMETHEUS_CLIENT_URL:-127.0.0.1:9090}"

ocp_prometheus_config_path="config/openshift"
scenario_configuration_path="config/benchmarks/scenarios/$SCENARIO_CONFIGURATION_DIRECTORY"
benchmarking_configuration_path="config/benchmarks/$BENCHMARKING_CONFIGURATION_DIRECTORY"
benchmarking_configuration_file="config/benchmarks/$BENCHMARKING_CONFIGURATION_DIRECTORY/benchmark.yaml"
ocp_prometheus_config_path="config/openshift"
scripts_path="hack/scripts"

port_counter=0

# Deploy Loki with the Observatorium configuration
# Intended to work on a Kind cluster for quick development
observatorium() {
    for scenario in $scenario_configuration_path/*.yaml; do
        create_benchmarking_environment
        
        pushd ../observatorium || exit 1
        KUBECTL=$(which kubectl) ./configuration/tests/e2e.sh deploy
        popd
        
        wait_for_ready_loki_components
        wait_for_ready_query_scheduler
        configure_prometheus
        run_benchmark_suite $scenario
        
        # Clean up
        destroy_benchmarking_environment
    done
}

# Deploy Loki with the Red Hat Observability Service configuration
# Intended to work on a OpenShift cluster for benchmarking
rhobs() {
    rhobs_loki_deployment_file=$1
    storage_bucket=$2

    for scenario in $scenario_configuration_path/*.yaml; do
        create_benchmarking_environment
        create_s3_storage $storage_bucket

        kubectl -n $BENCHMARK_NAMESPACE apply -f $rhobs_loki_deployment_file
        $scripts_path/deploy-example-secret.sh $BENCHMARK_NAMESPACE $storage_bucket

        wait_for_ready_loki_components
        wait_for_ready_query_scheduler
        configure_prometheus
        run_benchmark_suite $scenario

        # Clean Up
        echo -e "\nRemoving RHOBS configuration file"
        rm $rhobs_loki_deployment_file

        destroy_benchmarking_environment
        destroy_s3_storage $storage_bucket
    done
}

# Deploy Loki with the Red Hat Loki Operator
# Intended to work on a OpenShift cluster for benchmarking
operator() {
    operator_registry=$1
    storage_bucket=$2

    if $IS_OPENSHIFT; then
        echo -e "\nOverwriting BENCHMARK_NAMESPACE"
        BENCHMARK_NAMESPACE="openshift-logging"
    fi

    for scenario in $scenario_configuration_path/*.yaml; do
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

        # There is a small - sometimes noticeable - time gap, between the Lokistack CR 
        # being applied and the deployment and statefulsets being created.
        sleep 10

        wait_for_ready_loki_components
        configure_prometheus
        run_benchmark_suite $scenario

        # Clean Up
        destroy_benchmarking_environment
        destroy_s3_storage $storage_bucket
    done
}

create_benchmarking_environment() {
    echo -e "\nCreating output directory"
    mkdir -p $OUTPUT_DIRECTORY

    echo -e "\nExporting benchmark directory name"
    export BENCHMARKING_CONFIGURATION_DIRECTORY

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
    echo -e "\nDestroying benchmarking environment"

    if $USE_CADVISOR; then
         pushd ../cadvisor || exit 1
         kubectl $KUSTOMIZE deploy/kubernetes/base | kubectl delete -f -
         popd
    fi

    if $IS_OPENSHIFT; then
        disable_ocp_user_workload_monitoring
    fi

    if $IS_TESTING; then
        $KIND delete cluster
    else
        kubectl -n $BENCHMARK_NAMESPACE delete -f hack/loadclient-rbac.yaml --ignore-not-found=true
        kubectl delete namespace openshift-operators-redhat --ignore-not-found=true
        kubectl delete namespace $BENCHMARK_NAMESPACE --ignore-not-found=true
    fi
}

create_s3_storage() {
    bucket_names=$1

    if $IS_OPENSHIFT; then
        echo -e "\nCreating AWS S3 storage"
        $scripts_path/create-s3-bucket.sh $bucket_names
    fi
}

destroy_s3_storage() {
    bucket_names=$1

    if $IS_OPENSHIFT; then
        echo -e "\nDestroying AWS S3 storage"
        $scripts_path/delete-s3-bucket.sh $bucket_names
    fi
}

configure_prometheus() {
    echo -e "\nConfiguring Prometheus"

    if $IS_OPENSHIFT; then
        # There is a small time period between activating workload monitoring
        # and the resulting changes being applied.

        enable_ocp_user_workload_monitoring
        sleep 10
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

wait_for_ready_loki_components() {
    echo -e "\nWaiting for available querier deployment"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "deployments/$LOKI_COMPONENT_PREFIX-querier" --timeout=600s

    echo -e "\nWaiting for available query frontend deployment"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "deployments/$LOKI_COMPONENT_PREFIX-query-frontend" --timeout=600s

    echo -e "\nWaiting for available distributor deployment"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "deployments/$LOKI_COMPONENT_PREFIX-distributor" --timeout=600s

    echo -e "\nWaiting for available ingester statefulset"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-ingester" --timeout=600s

    echo -e "\nWaiting for available index gateway statefulset"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "statefulsets/$LOKI_COMPONENT_PREFIX-index-gateway" --timeout=600s
}

wait_for_ready_query_scheduler() {
    echo -e "\nWaiting for available query scheduler deployment"
    kubectl -n $BENCHMARK_NAMESPACE rollout status "deployments/$LOKI_COMPONENT_PREFIX-query-scheduler" --timeout=600s
}

run_benchmark_suite() {
    scenario_file=$1
    report_directory="$OUTPUT_DIRECTORY/$(basename $scenario_file .yaml)"

    mkdir -p $report_directory
    create_benchmarking_file $scenario_file

    echo -e "\nRunning benchmark suite"
    $GINKGO --output-dir=$report_directory --json-report="report.json" --timeout=4h ./benchmarks

    echo -e "\nMoving configuration file to report directory"
    mv $benchmarking_configuration_file $report_directory

    echo -e "\nProcessing JSON report"
    python3 $scripts_path/post_processing.py $report_directory
}

create_benchmarking_file() {
    scenario_file=$1

    echo -e "\nCreating benchmarking file"

    echo -e "\nCopying generator and querier configuration"
    cat $benchmarking_configuration_path/generator.yaml > $benchmarking_configuration_file
    cat $benchmarking_configuration_path/querier.yaml >> $benchmarking_configuration_file

    echo -e "\nCreating metrics configuration"
    cat <<-EOF >> $benchmarking_configuration_file
metrics:
  url: $PROMETHEUS_CLIENT_PROTOCOL://$PROMETHEUS_CLIENT_URL
  enableCadvisorMetrics: $IS_OPENSHIFT
  jobs:
    distributor: $LOKI_COMPONENT_PREFIX-distributor
    ingester: $LOKI_COMPONENT_PREFIX-ingester
    querier: $LOKI_COMPONENT_PREFIX-querier
    queryFrontend: $LOKI_COMPONENT_PREFIX-query-frontend
EOF

    echo -e "\nCopying scenario configuration"
    cat $scenario_file >> $benchmarking_configuration_file
}

enable_ocp_user_workload_monitoring() {
    echo -e "\nAdding user workload monitoring configuration"

    kubectl -n openshift-monitoring apply -f $ocp_prometheus_config_path/cluster-monitoring-config.yaml
	kubectl -n openshift-user-workload-monitoring apply -f $ocp_prometheus_config_path/user-workload-monitoring-config.yaml
}

disable_ocp_user_workload_monitoring() {
    echo -e "\nRemoving user workload monitoring configuration"

    kubectl -n openshift-monitoring delete -f $ocp_prometheus_config_path/cluster-monitoring-config.yaml --ignore-not-found=true
	kubectl -n openshift-user-workload-monitoring delete  -f $ocp_prometheus_config_path/user-workload-monitoring-config.yaml --ignore-not-found=true
}

export_ocp_prometheus_settings() {
    echo -e "\nRetrieving Prometheus URL and bearer token"

    PROMETHEUS_CLIENT_PROTOCOL="https"
    PROMETHEUS_CLIENT_URL="$(kubectl -n openshift-monitoring get route thanos-querier -o json | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["spec"]["host"])')"

    secret=$(kubectl -n openshift-user-workload-monitoring get secret | grep prometheus-user-workload-token | head -n 1 | awk '{print $1 }')
    export PROMETHEUS_TOKEN=$(kubectl -n openshift-user-workload-monitoring get secret $secret -o json | python3 -c 'import json,sys;obj=json.load(sys.stdin);print(obj["data"]["token"])' | base64 -d)
}

forward_ports() {
    shopt -s extglob

    cp config/prometheus/config.template config/prometheus/config.yaml

    setup_ports "loki query frontend" app.kubernetes.io/component=query-frontend LOKI_QUERY_FRONTEND_TARGETS 3100 $BENCHMARK_NAMESPACE
    setup_ports "loki distributor" app.kubernetes.io/component=distributor LOKI_DISTRIBUTOR_TARGETS 3100 $BENCHMARK_NAMESPACE
    setup_ports "loki ingester" app.kubernetes.io/component=ingester LOKI_INGESTER_TARGETS 3100 $BENCHMARK_NAMESPACE
    setup_ports "loki querier" app.kubernetes.io/component=querier LOKI_QUERIER_TARGETS 3100 $BENCHMARK_NAMESPACE
    setup_ports "cadvisor ingesters" "" CADVISOR_INGESTERS_TARGETS 8080 $BENCHMARK_NAMESPACE
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
    observatorium
    ;;

rhobs)
    rhobs $2 $3
    ;;

operator)
    operator $2 $3
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