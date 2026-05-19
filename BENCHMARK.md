# BENCHMARK

Measured reconcile latency for the controller in `cmd/controller` running against
a kwok-simulated cluster driven by `cmd/driver`. All numbers below come from the
JSONL log written by the reconciler (`RECONCILE_LOG`) and summarised by
`cmd/report`.

## Environment

```
go version go1.26.3 linux/amd64
kwokctl version v0.7.0 go1.24.0 (linux/amd64)
Linux hemang-Vivobook-ASUSLaptop-K3405VC-K3405VCB 6.17.0-19-generic #19~24.04.2-Ubuntu SMP PREEMPT_DYNAMIC Fri Mar  6 23:08:46 UTC 2 x86_64 x86_64 x86_64 GNU/Linux
model name      : 13th Gen Intel(R) Core(TM) i5-13500H
MemTotal:       16033760 kB
```

Single-host setup: kwok runs the simulated apiserver/etcd in `--runtime=binary`
mode, the controller and the driver run as plain Go processes on the same
machine.

## Workload

The driver in `cmd/driver` performs three things in sequence:

1. Creates N kwok Node objects via the apiserver.
2. Flips a random node's Ready condition between `True` and `False` at a fixed
   rate, measured in flips per node per minute.
3. On exit, deletes every node it created.

The reconcile loop lives in `internal/reconciler/node_reconciler.go`. Each
reconcile fetches the Node, reads the Ready condition, logs one structured
line, and returns. Duration is measured with `time.Since` taken at the start
of `Reconcile` and reported in milliseconds.

Parameters used for both scenarios below (Makefile defaults):

- churn rate: 30 flips per node per minute
- duration: 2m

The only variable across scenarios is `NODES`.

## Results

### 100 nodes

```
count   1749
mean      0.06 ms
p50       0.05 ms
p90       0.09 ms
p95       0.11 ms
p99       0.19 ms
max       1.52 ms
```

### 500 nodes

```
count   2429
mean      0.06 ms
p50       0.05 ms
p90       0.08 ms
p95       0.10 ms
p99       0.18 ms
max       4.09 ms
```

## Observations

- Reconcile work itself is trivial — a single `Get`, a condition scan, a log
  line — so per-reconcile cost is well under a millisecond at both sizes. The
  numbers reflect the harness, not a real controller's reconcile cost.
- Mean and p50 are flat (0.06 ms / 0.05 ms) across 100 and 500 nodes. The
  reconcile path does no per-cluster-size work, so this is expected.
- Tail moves with scale: max grows from 1.52 ms at 100 nodes to 4.09 ms at
  500 nodes, while p99 stays roughly the same (0.19 ms vs 0.18 ms). Most of
  the tail growth shows up only at the extreme — likely GC pauses, scheduler
  jitter, or the occasional client cache miss rather than reconcile work.
- Reconcile counts (1749 and 2429 over 2 minutes) are dominated by the
  initial Node create events plus the configured churn rate; they are not a
  throughput ceiling.

## Reproduce

```
make kwok-up
make benchmark NODES=100 CHURN=30 DURATION=2m
make benchmark NODES=500 CHURN=30 DURATION=2m
make kwok-down
```

`make benchmark` starts the controller with `RECONCILE_LOG=./reconcile.jsonl`,
runs the driver, kills the controller, and then runs `make report` against the
log file.

## Caveats

This is a PoC harness, not a benchmark of any real controller. The reconcile
loop here does one apiserver `Get` and a log line; a real controller would
spend most of its time in business logic, downstream API calls, and status
updates, none of which are exercised here.

Latencies above are dominated by controller-runtime workqueue handoff and the
local apiserver round-trip (kwok speaks the real Kubernetes API but runs in
the same process group), not by reconcile logic. Numbers should not be
compared to production controller SLOs.

Both scenarios ran on the same idle laptop with no resource isolation, so
absolute numbers are sensitive to background load. The shape of the
distribution (flat median, growing tail with scale) is the interesting signal,
not the millisecond values themselves.
