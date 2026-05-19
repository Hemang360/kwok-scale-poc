# kwok-scale-poc

A small harness that runs a controller-runtime-based Kubernetes controller
against a kwok-simulated cluster and measures reconcile latency under scale.
The goal is to exercise controller behavior against many simulated nodes
without provisioning real infrastructure.

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
make benchmark    # full run: controller + driver + latency report
```

## Running the benchmark

`make benchmark` orchestrates an end-to-end run. It builds the controller and
driver, starts the controller in the background with
`RECONCILE_LOG=./reconcile.jsonl`, runs the driver against the kwok cluster,
kills the controller, and prints latency statistics via `make report`.

Variables (overridable on the command line):

- `NODES` (default `100`): number of simulated nodes
- `CHURN` (default `30`): Ready/NotReady flips per node per minute
- `DURATION` (default `2m`): total run time
- `RECONCILE_LOG` (default `./reconcile.jsonl`): path the reconciler writes to

Example:

```
make kwok-up
make benchmark NODES=500 CHURN=30 DURATION=2m
make kwok-down
```

See [BENCHMARK.md](BENCHMARK.md) for measured numbers from this repo.

## Layout

- `cmd/controller/`: controller entrypoint (controller-runtime manager)
- `cmd/driver/`: scale driver that creates nodes and flips Ready conditions
- `cmd/report/`: JSONL reconcile-log parser and latency stats printer
- `internal/reconciler/`: Node reconciler implementation
- `hack/`: scripts for kwok lifecycle and cluster bootstrap

## What this is / what this isn't

This repo is a PoC harness for measuring reconcile latency against a kwok
cluster. The reconciler itself does one `Get` and writes a log line, which is
the entire reconcile path. Latencies measured here reflect controller-runtime
workqueue handoff and the local apiserver round-trip, not real controller
logic.

It is not a benchmark of any real controller, and the numbers in
[BENCHMARK.md](BENCHMARK.md) should not be compared to production controller
SLOs. Use this as a starting point to plug in a real reconciler and measure
its behavior under simulated scale, not as a fixed result.
