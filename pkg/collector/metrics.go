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
}

// NewMetricsCollector creates a new metrics collector instance
func NewMetricsCollector(kubeClient *kubernetes.Clientset, metricsClient *metrics.Clientset) *MetricsCollector {
	return &MetricsCollector{
		kubeClient:    kubeClient,
		metricsClient: metricsClient,
	}
}

// CollectNodeMetrics gathers metrics for all nodes
func (mc *MetricsCollector) CollectNodeMetrics() ([]metricsv1beta1.NodeMetrics, error) {
	nodeMetrics, err := mc.metricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeMetrics.Items, nil
}

// CollectPodMetrics gathers metrics for all pods
func (mc *MetricsCollector) CollectPodMetrics(namespace string) ([]metricsv1beta1.PodMetrics, error) {
	podMetrics, err := mc.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podMetrics.Items, nil
}

// StartCollection begins periodic metrics collection
func (mc *MetricsCollector) StartCollection(intervalSeconds int) {
	interval := time.Duration(intervalSeconds) * time.Second
	go func() {
		// Collect immediately on start
		mc.collectAndLogMetrics()

		ticker := time.NewTicker(interval)
		for range ticker.C {
			mc.collectAndLogMetrics()
		}
	}()
}

// collectAndLogMetrics collects and logs metrics in one place
func (mc *MetricsCollector) collectAndLogMetrics() {
	// Collect node metrics
	nodeMetrics, err := mc.CollectNodeMetrics()
	if err != nil {
		log.Printf("Error collecting node metrics: %v", err)
		return
	}
	log.Printf("Collected metrics from %d nodes", len(nodeMetrics))

	// Log some details about node metrics
	for _, node := range nodeMetrics {
		log.Printf("Node: %s", node.Name)
		log.Printf("  CPU: %v", node.Usage.Cpu())
		log.Printf("  Memory: %v", node.Usage.Memory())
	}

	// Collect pod metrics from all namespaces
	podMetrics, err := mc.CollectPodMetrics("")
	if err != nil {
		log.Printf("Error collecting pod metrics: %v", err)
		return
	}
	log.Printf("Collected metrics from %d pods", len(podMetrics))

	// Log some details about pod metrics
	for _, pod := range podMetrics {
		log.Printf("Pod: %s/%s", pod.Namespace, pod.Name)
		for _, container := range pod.Containers {
			log.Printf("  Container: %s", container.Name)
			log.Printf("    CPU: %v", container.Usage.Cpu())
			log.Printf("    Memory: %v", container.Usage.Memory())
		}
	}
} 