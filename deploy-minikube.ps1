# for windows powershell

# Function to handle errors
function Handle-Error {
    param($ErrorMessage)
    Write-Host "Error: $ErrorMessage" -ForegroundColor Red
    exit 1
}

Write-Host "Checking if Minikube is running..."
$minikubeStatus = minikube status
if ($minikubeStatus -notmatch "Running") {
    Write-Host "Starting Minikube..."
    minikube start
    if ($LASTEXITCODE -ne 0) {
        Handle-Error "Failed to start Minikube"
    }
}

Write-Host "Enabling metrics-server..."
minikube addons enable metrics-server
if ($LASTEXITCODE -ne 0) {
    Handle-Error "Failed to enable metrics-server"
}

Write-Host "Cleaning up previous deployment..."
kubectl delete -f k8s/deployment.yaml --ignore-not-found
kubectl delete -f k8s/rbac.yaml --ignore-not-found

Write-Host "Cleaning up Docker images..."
docker rmi k-monitor:latest -f 2>$null
minikube image rm k-monitor:latest 2>$null

Write-Host "Building Docker image..."
docker build -t k-monitor:latest . --no-cache --progress=plain
if ($LASTEXITCODE -ne 0) {
    Handle-Error "Failed to build Docker image"
}

Write-Host "Loading image into Minikube..."
minikube image load k-monitor:latest --overwrite=true
if ($LASTEXITCODE -ne 0) {
    Handle-Error "Failed to load image into Minikube"
}

Write-Host "Applying Kubernetes manifests..."
kubectl apply -f k8s/rbac.yaml
if ($LASTEXITCODE -ne 0) {
    Handle-Error "Failed to apply RBAC manifests"
}

kubectl apply -f k8s/deployment.yaml
if ($LASTEXITCODE -ne 0) {
    Handle-Error "Failed to apply deployment manifest"
}

Write-Host "Waiting for deployment to be ready..."
kubectl wait --for=condition=available deployment/k-monitor --timeout=120s
if ($LASTEXITCODE -ne 0) {
    Write-Host "Deployment might have failed. Checking pod status..." -ForegroundColor Yellow
    kubectl get pods -l app=k-monitor
    kubectl describe pods -l app=k-monitor
    Handle-Error "Deployment did not become ready in time"
}

Write-Host "`nDeployment complete! You can check the logs with:"
Write-Host "kubectl logs -f deployment/k-monitor" -ForegroundColor Green

# Enable local registry in minikube
minikube addons enable registry

# Set docker to use minikube's docker daemon
minikube docker-env | Invoke-Expression

# Build the container image
docker build -t k-monitor:latest .

# Apply RBAC roles first
kubectl apply -f k8s/rbac.yaml

# Apply the deployment
kubectl apply -f k8s/deployment.yaml

# Wait for deployment to be ready
kubectl rollout status deployment/k-monitor

# Get the URL to access the service
Write-Host "Getting service URL..."
minikube service k-monitor --url 