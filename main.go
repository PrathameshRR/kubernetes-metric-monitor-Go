package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"k-monitor/pkg/collector"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsServer struct {
	collector *collector.MetricsCollector
	mu        sync.RWMutex
	metrics   map[string]interface{}
	ready     bool
}

func NewMetricsServer(collector *collector.MetricsCollector) *MetricsServer {
	return &MetricsServer{
		collector: collector,
		metrics:   make(map[string]interface{}),
		ready:     false,
	}
}

func (s *MetricsServer) UpdateMetrics(metrics map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = metrics
	s.ready = true
}

func (s *MetricsServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !s.ready {
		http.Error(w, "Service not ready", http.StatusServiceUnavailable)
		return
	}

	if err := json.NewEncoder(w).Encode(s.metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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
	metricsServer := NewMetricsServer(metricsCollector)

	// Set up metrics collection with callback
	metricsCollector.SetCallback(func(metrics map[string]interface{}) {
		metricsServer.UpdateMetrics(metrics)
	})
	
	// Start collecting metrics
	stopCollection := metricsCollector.StartCollection(30)

	// Set up HTTP server with timeouts
	server := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Register handlers
	http.HandleFunc("/api/metrics", metricsServer.handleMetrics)
	
	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)
	
	// Start the service listening for requests.
	go func() {
		log.Println("Starting HTTP server on :8081")
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown... signal: %v\n", sig)
		
		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop metrics collection
		stopCollection()

		// Asking listener to shut down and shed load.
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v : %v", 30*time.Second, err)
			if err := server.Close(); err != nil {
				log.Printf("Error killing server : %v", err)
			}
		}
	}
} 