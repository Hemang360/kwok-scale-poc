CLUSTER_NAME := kwok-scale-poc
KUBECONFIG_PATH := .kwok/kubeconfig
NODES ?= 100
CHURN ?= 0
DURATION ?= 60s

.PHONY: kwok-up kwok-down kubeconfig controller build driver-build scale vet fmt clean

kwok-up:
	hack/setup-kwok.sh

kwok-down:
	kwokctl delete cluster --name=$(CLUSTER_NAME)

kubeconfig:
	@mkdir -p $(dir $(KUBECONFIG_PATH))
	@kwokctl get kubeconfig --name=$(CLUSTER_NAME) > $(KUBECONFIG_PATH)
	@echo $(KUBECONFIG_PATH)

controller: build kubeconfig
	KUBECONFIG=$(KUBECONFIG_PATH) ./bin/controller

build:
	go build -o bin/controller ./cmd/controller

driver-build:
	go build -o bin/driver ./cmd/driver

scale: driver-build kubeconfig
	KUBECONFIG=$(KUBECONFIG_PATH) ./bin/driver --nodes=$(NODES) --churn-rate=$(CHURN) --duration=$(DURATION)

vet:
	go vet ./...

fmt:
	gofmt -w -s .

clean:
	rm -rf bin .kwok
	rm -f *.log *.jsonl
