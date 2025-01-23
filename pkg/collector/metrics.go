package collector

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsCollector handles the collection of Kubernetes metrics
type MetricsCollector struct {
	kubeClient    *kubernetes.Clientset
	metricsClient *metrics.Clientset
	callback      func(map[string]interface{})
	stopCh        chan struct{}
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector(kubeClient *kubernetes.Clientset, metricsClient *metrics.Clientset) *MetricsCollector {
	return &MetricsCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
		stopCh:        make(chan struct{}),
	}
}

// SetCallback sets the callback function for metrics updates
func (mc *MetricsCollector) SetCallback(callback func(map[string]interface{})) {
	mc.callback = callback
}

// CollectNodeMetrics gathers metrics for all nodes
func (mc *MetricsCollector) CollectNodeMetrics() ([]metricsv1beta1.NodeMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeMetrics, err := mc.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeMetrics.Items, nil
}

// CollectPodMetrics gathers metrics for all pods
func (mc *MetricsCollector) CollectPodMetrics(namespace string) ([]metricsv1beta1.PodMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	podMetrics, err := mc.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podMetrics.Items, nil
}

// StartCollection begins periodic metrics collection
func (mc *MetricsCollector) StartCollection(intervalSeconds int) func() {
	interval := time.Duration(intervalSeconds) * time.Second
	
	// Start collection in a goroutine
	go func() {
		// Collect immediately on start
		if err := mc.collectAndLogMetrics(); err != nil {
			log.Printf("Initial metrics collection failed: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := mc.collectAndLogMetrics(); err != nil {
					log.Printf("Error collecting metrics: %v", err)
				}
			case <-mc.stopCh:
				log.Println("Stopping metrics collection")
				return
			}
		}
	}()

	// Return a function that can be called to stop collection
	return func() {
		close(mc.stopCh)
	}
}

// collectAndLogMetrics collects and logs metrics in one place
func (mc *MetricsCollector) collectAndLogMetrics() error {
	// Collect node metrics
	nodeMetrics, err := mc.CollectNodeMetrics()
	if err != nil {
		return err
	}

	// Collect pod metrics from all namespaces
	podMetrics, err := mc.CollectPodMetrics("")
	if err != nil {
		return err
	}

	// Create metrics map for callback
	metrics := map[string]interface{}{
		"nodes": nodeMetrics,
		"pods":  podMetrics,
	}

	// Call callback if set
	if mc.callback != nil {
		mc.callback(metrics)
	}

	return nil
} 