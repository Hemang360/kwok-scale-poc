CLUSTER_NAME := kwok-scale-poc
KUBECONFIG_PATH := .kwok/kubeconfig

.PHONY: kwok-up kwok-down kubeconfig controller build vet fmt clean

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

vet:
	go vet ./...

fmt:
	gofmt -w -s .

clean:
	rm -rf bin
	rm -f *.log *.jsonl
