# Bot Gastos Pareja Helm Chart

## Prerequisites

- Kubernetes cluster (k3s)
- kubectl configured
- helm 3.x
- Docker (for building image)

## Building and Deploying

### 1. Build Docker Image

```bash
cd /home/fuser/development/botGastosPareja
docker build -t botgastospareja:latest .
```

### 2. Load Image into k3s

Since k3s uses containerd, you need to import the image:

```bash
# Save the image
docker save botgastospareja:latest -o botgastospareja.tar

# Import into k3s
sudo k3s ctr images import botgastospareja.tar

# Or use k3d/k3s image import if available
```

Alternatively, if you have a registry:
```bash
docker tag botgastospareja:latest localhost:5000/botgastospareja:latest
docker push localhost:5000/botgastospareja:latest
```

### 3. Update values.yaml

Edit `values.yaml` and set your Telegram bot token:

```yaml
secrets:
  telegramBotToken: "YOUR_TELEGRAM_BOT_TOKEN_HERE"
```

### 4. Deploy with Helm

```bash
cd helm/botgastospareja

# Create namespace (optional)
kubectl create namespace botgastospareja

# Install
helm install botgastospareja . --namespace botgastospareja

# Or upgrade if already installed
helm upgrade botgastospareja . --namespace botgastospareja
```

### 5. Check Status

```bash
kubectl get pods -n botgastospareja
kubectl logs -f deployment/botgastospareja -n botgastospareja
```

## Configuration

Key values in `values.yaml`:

- `image.repository`: Docker image name
- `image.tag`: Image tag (default: latest)
- `secrets.telegramBotToken`: Your Telegram bot token
- `persistence.enabled`: Enable persistent storage for SQLite DB
- `persistence.size`: Storage size (default: 1Gi)
- `resources`: CPU/memory limits and requests

## Uninstall

```bash
helm uninstall botgastospareja --namespace botgastospareja
```

