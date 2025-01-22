# K-Monitor: Kubernetes Resource Monitoring Tool

A lightweight, efficient Kubernetes monitoring solution that provides real-time insights into cluster resource utilization.

## Features

- Real-time monitoring of Kubernetes cluster resources
- Node-level metrics collection (CPU, Memory)
- Pod-level metrics collection (CPU, Memory per container)
- Automatic metrics collection every 30 seconds
- In-cluster deployment using Kubernetes DaemonSet
- RBAC-compliant with minimal required permissions

## Prerequisites

- Kubernetes cluster (v1.28+)
- Metrics Server installed and running
- Docker (for building the image)
- Go 1.21+ (for development)
- kubectl configured with cluster access

## Installation

### 1. Using Minikube

```bash
# Start Minikube with metrics-server enabled
minikube start
minikube addons enable metrics-server

# Deploy K-Monitor
./deploy-minikube.ps1  # For Windows
# OR
./deploy-minikube.sh   # For Linux/Mac
```

### 2. Manual Installation

1. Build the Docker image:
```bash
docker build -t k-monitor:latest .
```

2. Apply the RBAC configuration:
```bash
kubectl apply -f k8s/rbac.yaml
```

3. Deploy the application:
```bash
kubectl apply -f k8s/deployment.yaml
```

## Architecture

K-Monitor consists of the following components:

1. **Metrics Collector**: Core component that interfaces with the Kubernetes Metrics API
2. **RBAC Configuration**: ServiceAccount, ClusterRole, and ClusterRoleBinding for secure access
3. **Deployment**: Kubernetes deployment configuration for running the monitor

### Project Structure
```
k-monitor/
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── pkg/
│   └── collector/
│       └── metrics.go
└── k8s/
    ├── deployment.yaml
    └── rbac.yaml
```

## Configuration

The application can be configured through the following methods:

1. **In-Cluster Configuration**: Automatically used when deployed inside Kubernetes
2. **Kubeconfig**: Falls back to local kubeconfig when running outside the cluster

## Metrics Collection

K-Monitor collects the following metrics:

### Node Metrics
- CPU Usage
- Memory Usage

### Pod Metrics
- Per-container CPU Usage
- Per-container Memory Usage

## Development

1. Clone the repository:
```bash
git clone https://github.com/prathameshrr/k-monitor.git
cd k-monitor
```

2. Install dependencies:
```bash
go mod download
```

3. Run locally:
```bash
go run main.go
```

## Building

Build the Docker image:
```bash
docker build -t k-monitor:latest .
```

## Deployment

1. Apply RBAC configuration:
```bash
kubectl apply -f k8s/rbac.yaml
```

2. Deploy the application:
```bash
kubectl apply -f k8s/deployment.yaml
```

## Monitoring

View the metrics:
```bash
kubectl logs -f deployment/k-monitor
```

## Troubleshooting

1. **Pod in CrashLoopBackOff**:
   - Check if metrics-server is running
   - Verify RBAC permissions
   - Check pod logs for errors

2. **No Metrics Data**:
   - Ensure metrics-server is properly configured
   - Check network policies
   - Verify API server access

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.

## Acknowledgments

- Kubernetes Metrics API
- client-go library
- metrics-server project
