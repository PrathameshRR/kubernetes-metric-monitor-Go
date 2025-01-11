#!/bin/bash

echo "Checking if Minikube is running..."
if ! minikube status | grep -q "Running"; then
    echo "Starting Minikube..."
    minikube start
fi

echo "Enabling metrics-server..."
minikube addons enable metrics-server

echo "Building Docker image..."
docker build -t k-monitor:latest .

echo "Loading image into Minikube..."
minikube image load k-monitor:latest

echo "Applying Kubernetes manifests..."
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/deployment.yaml

echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available deployment/k-monitor --timeout=60s

echo "Deployment complete! You can check the logs with:"
echo "kubectl logs -f deployment/k-monitor"