export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

include .bingo/Variables.mk

export GOROOT=$(shell go env GOROOT)
export GO111MODULE=on

export REPORT_DIR?=$(CURDIR)/reports/$(shell date +%Y-%m-%d-%H-%M-%S)

export KUBECTL=$(shell which kubectl)
export USERNAME=$(USER)

CADVISOR_NAMESPACE := cadvisor
LOKI_NAMESPACE := observatorium-logs-test

LOKI_STORAGE_BUCKET ?= loki-benchmark-$(USERNAME)
LOKI_TEMPLATE_FILE ?= /tmp/observatorium-logs-template.yaml
LOKI_CONFIG_FILE ?= config/loki-parameters.yaml

OCP_PROM_CONFIG_FILE := /tmp/cluster-monitoring-config.yaml

.DEFAULT_GOAL := help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Ingredients
lint: $(GOLANGCI_LINT) ## Lint the code
	@$(GOLANGCI_LINT) run --timeout=4m

$(REPORT_DIR):
	@mkdir -p $(REPORT_DIR)
	@cp reports/README.template $(REPORT_DIR)
	@mv $(REPORT_DIR)/README.template $(REPORT_DIR)/README.md

download-obs-loki-template: ## Download loki observatorium template locally
	wget -nv -O $(LOKI_TEMPLATE_FILE) https://raw.githubusercontent.com/rhobs/configuration/main/resources/services/observatorium-logs-template.yaml
.PHONY: download-obs-loki-template

obs-loki-template-cleanup: ## Cleanup local loki observatorium template
	rm -f $(LOKI_TEMPLATE_FILE)
.PHONY: obs-loki-template-cleanup

ifeq ($(shell test -e $(LOKI_TEMPLATE_FILE) && echo -n yes),yes)
deploy-obs-loki: ## Deploy loki from local observatorium template
	oc create namespace $(LOKI_NAMESPACE)
	hack/deploy-example-secret.sh $(LOKI_NAMESPACE) $(LOKI_STORAGE_BUCKET)
	oc process -f $(LOKI_TEMPLATE_FILE) -p NAMESPACE=$(LOKI_NAMESPACE) -p LOKI_S3_SECRET=test --param-file $(LOKI_CONFIG_FILE) | oc -n $(LOKI_NAMESPACE) apply -f -
else
deploy-obs-loki:
	$(error Can't find Loki observatorium deployment template file: $(LOKI_TEMPLATE_FILE). Use "make download-obs-loki-template" to download )
endif
.PHONY:deploy-obs-loki

obs-loki-cleanup: ## Cleanup loki deployment
	oc --ignore-not-found=true delete namespace $(LOKI_NAMESPACE)
.PHONY: obs-loki-cleanup

##@ Development
deploy-cadvisor: $(KUSTOMIZE) ## Deploy cadvisor (https://github.com/google/cadvisor)
	oc create namespace $(CADVISOR_NAMESPACE)
	$(KUSTOMIZE) build ../cadvisor/deploy/kubernetes/base | oc -n $(CADVISOR_NAMESPACE) apply -f -
	oc -n $(CADVISOR_NAMESPACE) adm policy add-scc-to-user privileged -z $(CADVISOR_NAMESPACE)
	oc -n $(CADVISOR_NAMESPACE) adm policy add-cluster-role-to-user cluster-reader -z $(CADVISOR_NAMESPACE)
.PHONY: deploy-cadvisor

cadvisor-cleanup: ## Cleanup cadvisor
	oc --ignore-not-found=true delete namespace $(CADVISOR_NAMESPACE)
.PHONY: cadvisor-cleanup

bench-dev: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR) ## Execute benchmark
	@TARGET_ENV=development \
	KUBECTL=../observatorium/kubectl \
	OBS_NS=observatorium \
	OBS_LOKI_QF="observatorium-xyz-loki-query-frontend" \
	OBS_LOKI_QR="observatorium-xyz-loki-querier" \
	OBS_LOKI_DST="observatorium-xyz-loki-distributor" \
	OBS_LOKI_ING="observatorium-xyz-loki-ingester" \
	./run.sh
.PHONY: bench-dev

##@ Benchmark Loki on OpenShift
deploy-s3-bucket: ## Deploy s3 bucket
	hack/create-s3-bucket.sh $(LOKI_STORAGE_BUCKET)
.PHONY: deploy-s3-bucket

s3-bucket-cleanup: ## Delete s3 bucket
	hack/delete-s3-bucket.sh $(LOKI_STORAGE_BUCKET)
.PHONY: s3-bucket-cleanup

deploy-ocp-prometheus: ## Deploy prometheus
	oc -n openshift-monitoring apply -f hack/ocp-monitoring/cluster-monitoring-config.yaml
	oc -n openshift-user-workload-monitoring apply -f hack/ocp-monitoring/user-workload-monitoring-config.yaml
.PHONY: deploy-ocp-prometheus

ocp-prometheus-cleanup: ## Cleanup prometheus deployment
	oc --ignore-not-found=true -n openshift-monitoring delete -f hack/ocp-monitoring/cluster-monitoring-config.yaml
	oc --ignore-not-found=true -n openshift-user-workload-monitoring delete  -f hack/ocp-monitoring/user-workload-monitoring-config.yaml
.PHONY: ocp-prometheus-cleanup

deploy-ocp-loki: deploy-obs-loki ## Deploy loki
.PHONY: deploy-ocp-loki

ocp-loki-cleanup: obs-loki-cleanup ## Cleanup loki deployment
.PHONY: ocp-loki-cleanup

ocp-run-benchmarks: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR) ## Run benchmarks
	@TARGET_ENV=ocp-observatorium-test \
	DEPLOY_OCP_PROMETHEUS=true \
	OBS_NS=$(LOKI_NAMESPACE) \
	OBS_LOKI_QF="observatorium-loki-query-frontend" \
	OBS_LOKI_QR="observatorium-loki-querier" \
	OBS_LOKI_DST="observatorium-loki-distributor" \
	OBS_LOKI_ING="observatorium-loki-ingester" \
	./run.sh
.PHONY: ocp-run-benchmarks

ocp-all-cleanup: obs-loki-template-cleanup s3-bucket-cleanup ocp-prometheus-cleanup  ocp-loki-cleanup ## Cleanup all
.PHONY: ocp-all-cleanup

##@ All-in-one
ocp-init-bench: ocp-all-cleanup download-obs-loki-template deploy-s3-bucket ## Setup benchmarks on OpenShift (first deployment only)
.PHONY: ocp-init-bench

ocp-loki-bench: deploy-s3-bucket ocp-prometheus-cleanup deploy-ocp-prometheus ocp-loki-cleanup deploy-ocp-loki ocp-run-benchmarks ## Deploy and Execute loki benchmarks on OCP
.PHONY: ocp-loki-bench
