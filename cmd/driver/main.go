package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")
	nodes := flag.Int("nodes", 100, "number of nodes to create")
	churnRate := flag.Float64("churn-rate", 0, "Ready/NotReady flips per node per minute")
	duration := flag.Duration("duration", 60*time.Second, "total run time")
	prefix := flag.String("prefix", "kwok-node-", "node name prefix")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runCtx, cancel := context.WithTimeout(ctx, *duration)
	defer cancel()

	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if *kubeconfig != "" {
		loading.ExplicitPath = *kubeconfig
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatalf("load kubeconfig: %v", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("clientset: %v", err)
	}

	created := make([]string, 0, *nodes)
	ready := make(map[string]bool, *nodes)
	for i := 0; i < *nodes; i++ {
		name := fmt.Sprintf("%s%d", *prefix, i)
		if err := createNode(runCtx, cs, name); err != nil {
			log.Printf("create %s: %v", name, err)
			continue
		}
		created = append(created, name)
		ready[name] = true
	}
	log.Printf("created=%d", len(created))

	defer cleanup(created, cs)

	var flap <-chan time.Time
	if *churnRate > 0 && len(created) > 0 {
		interval := time.Duration(float64(time.Minute) / (*churnRate * float64(len(created))))
		if interval < time.Millisecond {
			interval = time.Millisecond
		}
		t := time.NewTicker(interval)
		defer t.Stop()
		flap = t.C
	}

	summary := time.NewTicker(5 * time.Second)
	defer summary.Stop()
	start := time.Now()
	flips := 0

	for {
		select {
		case <-runCtx.Done():
			log.Printf("done created=%d flips=%d elapsed=%s", len(created), flips, time.Since(start).Round(time.Second))
			return
		case <-flap:
			name := created[rand.Intn(len(created))]
			ready[name] = !ready[name]
			if err := patchReady(runCtx, cs, name, ready[name]); err != nil {
				log.Printf("patch %s: %v", name, err)
				continue
			}
			flips++
		case <-summary.C:
			log.Printf("created=%d flips=%d elapsed=%s", len(created), flips, time.Since(start).Round(time.Second))
		}
	}
}

func createNode(ctx context.Context, cs *kubernetes.Clientset, name string) error {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"type":                  "kwok",
				"beta.kubernetes.io/os": "linux",
			},
			Annotations: map[string]string{
				"kwok.x-k8s.io/node": "fake",
			},
		},
	}
	_, err := cs.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func patchReady(ctx context.Context, cs *kubernetes.Clientset, name string, ready bool) error {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	body := fmt.Sprintf(
		`{"status":{"conditions":[{"type":"Ready","status":"%s","lastTransitionTime":"%s"}]}}`,
		status, time.Now().UTC().Format(time.RFC3339),
	)
	_, err := cs.CoreV1().Nodes().Patch(ctx, name, types.StrategicMergePatchType, []byte(body), metav1.PatchOptions{}, "status")
	return err
}

func cleanup(names []string, cs *kubernetes.Clientset) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, n := range names {
		if err := cs.CoreV1().Nodes().Delete(ctx, n, metav1.DeleteOptions{}); err != nil {
			log.Printf("delete %s: %v", n, err)
		}
	}
	log.Printf("cleaned up %d nodes", len(names))
}
