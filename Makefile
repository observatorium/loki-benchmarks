export GOBIN=$(CURDIR)/bin
export PATH:=$(GOBIN):$(PATH)

include .bingo/Variables.mk

export GOROOT=$(shell go env GOROOT)
export GOFLAGS=-mod=vendor
export GO111MODULE=on

export REPORT_DIR?=$(CURDIR)/reports/$(shell date +%Y-%m-%d-%H-%M-%S)

export KUBECTL=$(shell command -v kubectl)

CADVISOR_NAMESPACE := cadvisor
LOKI_NAMESPACE := observatorium-logs-test

LOKI_STORAGE_BUCKET ?= loki-benchmark-storage
LOKI_TEMPLATE_FILE ?= /tmp/observatorium-logs-template.yaml
LOKI_CONFIG_FILE ?= config/loki-parameters.yaml

all: lint bench-dev

lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

$(REPORT_DIR):
	@mkdir -p $(REPORT_DIR)

download-observatorium-loki-template:
	wget -nv -O $(LOKI_TEMPLATE_FILE) https://raw.githubusercontent.com/rhobs/configuration/main/resources/services/observatorium-logs-template.yaml
.PHONY: download-observatorium-loki-template

deploy-cadvisor: $(KUSTOMIZE)
	oc create namespace $(CADVISOR_NAMESPACE)
	oc -n $(CADVISOR_NAMESPACE) adm policy add-scc-to-user privileged -z $(CADVISOR_NAMESPACE)
	oc -n $(CADVISOR_NAMESPACE) adm policy add-cluster-role-to-user cluster-reader -z $(CADVISOR_NAMESPACE)
	$(KUSTOMIZE) build ../cadvisor/deploy/kubernetes/base | oc -n $(CADVISOR_NAMESPACE) apply -f -
.PHONY: deploy-cadvisor

deploy-observatorium-loki: download-observatorium-loki-template
	oc create namespace $(LOKI_NAMESPACE)
	hack/deploy-example-secret.sh $(LOKI_NAMESPACE) $(LOKI_STORAGE_BUCKET)
	oc process -f $(LOKI_TEMPLATE_FILE) -p NAMESPACE=$(LOKI_NAMESPACE) -p LOKI_S3_SECRET=test --param-file $(LOKI_CONFIG_FILE) | oc -n $(LOKI_NAMESPACE) apply -f -
.PHONY:deploy-observatorium-loki

bench-dev: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR)
	@TARGET_ENV=development \
	KUBECTL=../observatorium/kubectl \
	OBS_NS=observatorium \
	OBS_LOKI_QF="observatorium-xyz-loki-query-frontend" \
	OBS_LOKI_QR="observatorium-xyz-loki-querier" \
	OBS_LOKI_DST="observatorium-xyz-loki-distributor" \
	OBS_LOKI_ING="observatorium-xyz-loki-ingester" \
	./run.sh
.PHONY: bench-dev

bench-obs-logs-test: $(GINKGO) $(PROMETHEUS) $(EMBEDMD) $(REPORT_DIR)
	@TARGET_ENV=ocp-observatorium-test \
	OBS_NS=$(LOKI_NAMESPACE) \
	OBS_LOKI_QF="observatorium-loki-query-frontend" \
	OBS_LOKI_QR="observatorium-loki-querier" \
	OBS_LOKI_DST="observatorium-loki-distributor" \
	OBS_LOKI_ING="observatorium-loki-ingester" \
	./run.sh
.PHONY: bench-obs-logs-test

cadvisor-cleanup:
	oc delete namespace $(CADVISOR_NAMESPACE)
.PHONY: cadvisor-cleanup

observatorium-loki-cleanup:
	oc delete namespace $(LOKI_NAMESPACE)
.PHONY: observatorium-loki-cleanup
