package reconciler

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NodeReconciler struct {
	client.Client

	logPath string
	mu      sync.Mutex
	file    *os.File
}

type logEntry struct {
	Timestamp  string  `json:"ts"`
	Node       string  `json:"node"`
	Ready      string  `json:"ready"`
	DurationMS float64 `json:"duration_ms"`
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()
	logger := log.FromContext(ctx)

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	ready := string(corev1.ConditionUnknown)
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			ready = string(c.Status)
			break
		}
	}

	dur := time.Since(start)
	logger.Info("reconciled node", "node", node.Name, "ready", ready, "duration_ms", float64(dur.Microseconds())/1000.0)

	if r.logPath != "" {
		r.writeJSONL(node.Name, ready, float64(dur.Microseconds())/1000.0)
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) writeJSONL(name, ready string, durMS float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		f, err := os.OpenFile(r.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return
		}
		r.file = f
	}

	entry := logEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Node:       name,
		Ready:      ready,
		DurationMS: durMS,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	b = append(b, '\n')
	_, _ = r.file.Write(b)
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.logPath = os.Getenv("RECONCILE_LOG")
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
