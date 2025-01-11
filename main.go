package main

import (
	"flag"
	"log"
	"path/filepath"

	"k-monitor/pkg/collector"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	log.Println("Starting k-monitor...")
	
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		log.Println("Failed to get in-cluster config, falling back to kubeconfig:", err)
		
		// Fallback to kubeconfig
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatalf("Error building config: %v", err)
		}
	} else {
		log.Println("Successfully loaded in-cluster config")
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create Metrics clientset
	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating metrics client: %v", err)
	}

	log.Println("Successfully connected to Kubernetes cluster")

	// Initialize metrics collector
	metricsCollector := collector.NewMetricsCollector(clientset, metricsClient)
	
	// Start collecting metrics every 30 seconds
	metricsCollector.StartCollection(30)

	log.Println("Metrics collection started")

	// Keep the application running
	select {}
} 