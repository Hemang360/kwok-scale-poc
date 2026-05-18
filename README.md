# kwok-scale-poc

A proof of concept that runs a controller runtime based kubernetes controller
against a kwok simulated cluster and measures reconcile latency under scale.
The goal is to exercise controller behavior against thousands of simulated
nodes and pods without provisioning real infrastructure.

## Prerequisites

- Go 1.26+
- `kwokctl` and `kwok` on `PATH`
- `kubectl`
- `make`

## Quick start

```
make kwok-up      # start a local kwok cluster
make controller   # build and run the controller against it
make scale        # generate simulated workload
make benchmark    # collect reconcile latency metrics
```

## Layout

- `cmd/` — controller and tooling entry points
- `internal/controller/` — reconciler implementation
- `internal/scale/` — workload generators
- `internal/bench/` — latency measurement and reporting
- `hack/` — scripts for kwok lifecycle and cluster bootstrap
- `config/` — manifests and kwok stage configs
