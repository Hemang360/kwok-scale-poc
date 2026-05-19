package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

type entry struct {
	Timestamp  string  `json:"ts"`
	Node       string  `json:"node"`
	Ready      string  `json:"ready"`
	DurationMS float64 `json:"duration_ms"`
}

func main() {
	logPath := flag.String("log", "", "path to JSONL reconcile log")
	flag.Parse()

	path := *logPath
	if path == "" && flag.NArg() > 0 {
		path = flag.Arg(0)
	}
	if path == "" {
		fmt.Fprintln(os.Stderr, "usage: report --log <path>")
		os.Exit(2)
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var durs []float64
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	line := 0
	for scanner.Scan() {
		line++
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		var e entry
		if err := json.Unmarshal(raw, &e); err != nil {
			fmt.Fprintf(os.Stderr, "line %d: %v\n", line, err)
			continue
		}
		durs = append(durs, e.DurationMS)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scan: %v\n", err)
		os.Exit(1)
	}

	n := len(durs)
	if n == 0 {
		fmt.Println("  count   0")
		return
	}

	sort.Float64s(durs)

	var sum float64
	for _, d := range durs {
		sum += d
	}
	mean := sum / float64(n)

	fmt.Printf("  %-7s %d\n", "count", n)
	fmt.Printf("  %-7s %6.2f ms\n", "mean", mean)
	fmt.Printf("  %-7s %6.2f ms\n", "p50", pct(durs, 50))
	fmt.Printf("  %-7s %6.2f ms\n", "p90", pct(durs, 90))
	fmt.Printf("  %-7s %6.2f ms\n", "p95", pct(durs, 95))
	fmt.Printf("  %-7s %6.2f ms\n", "p99", pct(durs, 99))
	fmt.Printf("  %-7s %6.2f ms\n", "max", durs[n-1])
}

// nearest-rank percentile on sorted slice
func pct(sorted []float64, p float64) float64 {
	n := len(sorted)
	idx := int((p/100.0)*float64(n-1) + 0.5)
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return sorted[idx]
}
