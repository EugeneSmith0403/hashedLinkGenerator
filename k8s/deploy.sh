#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="link-generator"
ENV_FILE="${1:-.env}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()    { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ── Prerequisites ────────────────────────────────────────────────────────────
command -v kubectl &>/dev/null || error "kubectl not found"

if [[ ! -f "$ENV_FILE" ]]; then
  error ".env file not found at '$ENV_FILE'. Usage: ./deploy.sh [path/to/.env]"
fi

info "Starting deployment to namespace '$NAMESPACE'"

# ── 1. Namespace ─────────────────────────────────────────────────────────────
info "Step 1/7 — Namespace"
kubectl apply -f namespace.yaml

# ── 2. Secrets ───────────────────────────────────────────────────────────────
info "Step 2/7 — Secrets"
if kubectl get secret app-secrets -n "$NAMESPACE" &>/dev/null; then
  warn "Secret 'app-secrets' already exists — deleting and recreating"
  kubectl delete secret app-secrets -n "$NAMESPACE"
fi
kubectl create secret generic app-secrets \
  --from-env-file="$ENV_FILE" \
  --namespace="$NAMESPACE"

# ── 3. StatefulSets ──────────────────────────────────────────────────────────
info "Step 3/7 — StatefulSets (postgres, redis, rabbitmq, clickhouse)"
kubectl apply -f statefulsets/

info "Waiting for StatefulSets to be ready..."
for ss in postgres redis rabbitmq clickhouse; do
  kubectl rollout status statefulset/"$ss" -n "$NAMESPACE" --timeout=120s
done

# ── 4. Services ──────────────────────────────────────────────────────────────
info "Step 4/7 — Services"
kubectl apply -f services/

# ── 5. Deployments ───────────────────────────────────────────────────────────
info "Step 5/7 — Deployments (server + consumers)"
kubectl apply -f deployments/

info "Waiting for Deployments to be ready..."
for deploy in server consumer-payment consumer-subscription consumer-invoice consumer-stats; do
  kubectl rollout status deployment/"$deploy" -n "$NAMESPACE" --timeout=120s
done

# ── 6. Ingress & HPA ─────────────────────────────────────────────────────────
info "Step 6/7 — Ingress & HPA"
kubectl apply -f ingress/
kubectl apply -f hpa/

# ── 7. Monitoring ────────────────────────────────────────────────────────────
info "Step 7/7 — Monitoring (Prometheus, Grafana, exporters)"
kubectl apply -f monitoring/

# ── Done ─────────────────────────────────────────────────────────────────────
echo ""
info "Deployment complete!"
echo ""
echo "  Pods:"
kubectl get pods -n "$NAMESPACE"
echo ""
echo "  Services:"
kubectl get svc -n "$NAMESPACE"
