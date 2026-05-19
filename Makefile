CLUSTER_NAME := kwok-scale-poc
KUBECONFIG_PATH := .kwok/kubeconfig
NODES ?= 100
CHURN ?= 30
DURATION ?= 2m
RECONCILE_LOG ?= ./reconcile.jsonl

.PHONY: kwok-up kwok-down kubeconfig controller build driver-build scale report benchmark vet fmt clean

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

report:
	go run ./cmd/report --log $(RECONCILE_LOG)

benchmark: build driver-build kubeconfig
	@rm -f $(RECONCILE_LOG)
	@set -e; \
	trap 'kill $$CTRL_PID 2>/dev/null || true; wait $$CTRL_PID 2>/dev/null || true' EXIT; \
	KUBECONFIG=$(KUBECONFIG_PATH) RECONCILE_LOG=$(RECONCILE_LOG) ./bin/controller & \
	CTRL_PID=$$!; \
	sleep 2; \
	KUBECONFIG=$(KUBECONFIG_PATH) ./bin/driver --nodes=$(NODES) --churn-rate=$(CHURN) --duration=$(DURATION); \
	kill $$CTRL_PID 2>/dev/null || true; \
	wait $$CTRL_PID 2>/dev/null || true; \
	$(MAKE) report RECONCILE_LOG=$(RECONCILE_LOG)

vet:
	go vet ./...

fmt:
	gofmt -w -s .

clean:
	rm -rf bin .kwok
	rm -f *.log *.jsonl
