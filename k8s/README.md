# Kubernetes — Production Infrastructure

All manifests are deployed to the `link-generator` namespace.

## Directory Structure

```
k8s/
├── namespace.yaml          # Namespace definition
├── secrets/                # Credentials (created from .env)
├── statefulsets/           # Databases and queues
├── deployments/            # Application workloads
├── services/               # Internal cluster networking
├── ingress/                # External access + TLS
├── hpa/                    # Horizontal Pod Autoscaling
├── monitoring/             # Prometheus, Grafana, exporters
├── deploy.sh               # One-command deployment script
└── README.md
```

---

## CI/CD

Every push to `master` triggers the GitHub Actions pipeline ([.github/workflows/deploy.yml](../.github/workflows/deploy.yml)):

```
push to master
    │
    ├─► test       — go test ./...
    │
    ├─► build      — Docker build + push to ghcr.io (5 images in parallel)
    │
    └─► deploy     — kubectl apply → rollout status
```

### Required GitHub Secrets

Go to **Settings → Secrets → Actions** and add:

| Secret | Description |
|---|---|
| `KUBE_CONFIG` | `cat ~/.kube/config \| base64` — kubeconfig for the production cluster |

`GITHUB_TOKEN` is provided automatically by GitHub Actions.

### Docker Images

Images are published to GitHub Container Registry (`ghcr.io`):

| Image | Source |
|---|---|
| `link-generator-server` | `docker/Dockerfile.server` |
| `link-generator-consumer-payment` | `docker/Dockerfile.consumer` (`CONSUMER=paymentIntent`) |
| `link-generator-consumer-subscription` | `docker/Dockerfile.consumer` (`CONSUMER=subscription`) |
| `link-generator-consumer-invoice` | `docker/Dockerfile.consumer` (`CONSUMER=invoice`) |
| `link-generator-consumer-stats` | `docker/Dockerfile.consumer` (`CONSUMER=stats`) |

Each image is tagged with both `latest` and the commit SHA.

---

## One-Command Deploy

For manual deploys or first-time setup:

```bash
cd k8s
./deploy.sh            # uses .env from project root
./deploy.sh /path/to/.env  # or specify a custom path
```

The script runs 7 steps sequentially and waits for each rollout to complete before proceeding.

### Prerequisites

```bash
# nginx ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml

# cert-manager for automatic TLS
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

---

## Infrastructure Overview

### StatefulSets — persistent data

| Service    | Image                    | Storage | CPU        | RAM         |
|------------|--------------------------|---------|------------|-------------|
| postgres   | postgres:16              | 20Gi    | 250m–1000m | 512Mi–2Gi   |
| redis      | redis:latest             | 5Gi     | 100m–500m  | 256Mi–512Mi |
| rabbitmq   | rabbitmq:3-management    | 5Gi     | 100m–500m  | 256Mi–512Mi |
| clickhouse | clickhouse-server:latest | 50Gi    | 250m–2000m | 512Mi–4Gi   |

All StatefulSets use headless services (`clusterIP: None`) for stable pod DNS.

### Deployments — application workloads

| Service               | Replicas | CPU       | RAM          |
|-----------------------|----------|-----------|--------------|
| server                | 2        | 100m–500m | 128Mi–512Mi  |
| consumer-payment      | 2        | 50m–200m  | 64Mi–256Mi   |
| consumer-subscription | 2        | 50m–200m  | 64Mi–256Mi   |
| consumer-invoice      | 2        | 50m–200m  | 64Mi–256Mi   |
| consumer-stats        | 2        | 100m–500m | 128Mi–512Mi  |

Consumers process RabbitMQ tasks. Each consumer type is a separate Deployment for independent scaling.

### Autoscaling (HPA)

| Target         | Min | Max | CPU  | RAM |
|----------------|-----|-----|------|-----|
| server         | 2   | 10  | 70%  | 80% |
| consumer-stats | 2   | 8   | 70%  | —   |

### Networking

- Databases and queues are **cluster-internal only** (headless ClusterIP)
- Server is exposed externally via **Ingress** on your domain
- TLS is managed automatically by **cert-manager + Let's Encrypt**
- Grafana is available at `grafana.yourdomain.com`

Update your domain in [ingress/ingress.yaml](ingress/ingress.yaml) before deploying.

### Monitoring

| Component           | Port | Purpose                               |
|---------------------|------|---------------------------------------|
| prometheus          | 9090 | Metrics collection                    |
| grafana             | 3000 | Dashboards                            |
| postgres-exporter   | 9187 | PostgreSQL metrics                    |
| redis-exporter      | 9121 | Redis metrics                         |
| rabbitmq-exporter   | 9419 | RabbitMQ queue depth, message rates   |
| clickhouse-exporter | 9116 | ClickHouse query and insert metrics   |

---

## Secrets

All credentials are stored in a single Kubernetes Secret `app-secrets`, created from your `.env` file:

```bash
kubectl create secret generic app-secrets \
  --from-env-file=.env \
  --namespace=link-generator
```

> Never commit real secrets to git. Use [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) or [Vault](https://www.vaultproject.io/) in production.

Copy `.env.example` to `.env` and fill in all values before deploying.

---

## Useful Commands

```bash
# Pod status
kubectl get pods -n link-generator

# Stream server logs
kubectl logs -n link-generator deploy/server -f

# Stream consumer logs
kubectl logs -n link-generator deploy/consumer-stats -f

# HPA status
kubectl get hpa -n link-generator

# Restart a deployment after image update
kubectl rollout restart deploy/server -n link-generator

# Rollout status
kubectl rollout status deploy/server -n link-generator

# Open postgres shell
kubectl exec -it -n link-generator statefulset/postgres -- psql -U postgres -d link

# Open redis CLI
kubectl exec -it -n link-generator statefulset/redis -- redis-cli
```

---

## Known Issues

- [statefulsets/postgres.yaml](statefulsets/postgres.yaml) — `POSTGRES_PASSWORD` incorrectly references the `REDIS_USER_PASSWORD` secret key. Fix before deploying to production.
- [monitoring/grafana.yaml](monitoring/grafana.yaml) — Grafana admin password is set to `admin`. Change it before deploying.
