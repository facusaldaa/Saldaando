#!/bin/bash
# Backup SQLite database from k3s pod

set -e

# Default values (can be overridden via env vars or args)
NAMESPACE="${NAMESPACE:-default}"
RELEASE_NAME="${RELEASE_NAME:-botgastospareja}"
POD_NAME="${POD_NAME:-}"
DB_PATH="/data/bot.db"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/bot.db.k3s.backup.${TIMESTAMP}"

# Get pod name if not provided
if [ -z "$POD_NAME" ]; then
    POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=${RELEASE_NAME}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [ -z "$POD_NAME" ]; then
        echo "Error: Could not find pod. Try:"
        echo "  kubectl get pods -n $NAMESPACE"
        echo "Or set POD_NAME env var: POD_NAME=your-pod-name $0"
        exit 1
    fi
fi

echo "Using pod: $POD_NAME (namespace: $NAMESPACE)"
echo "Database path: $DB_PATH"

mkdir -p "$BACKUP_DIR"

# Check if database exists in pod
if ! kubectl exec -n "$NAMESPACE" "$POD_NAME" -- test -f "$DB_PATH" 2>/dev/null; then
    echo "Error: Database file not found at $DB_PATH in pod"
    exit 1
fi

echo "Creating backup from k3s pod..."
kubectl cp "${NAMESPACE}/${POD_NAME}:${DB_PATH}" "$BACKUP_FILE"

# Also create a compressed backup
gzip -c "$BACKUP_FILE" > "${BACKUP_FILE}.gz"

echo "Backup created: $BACKUP_FILE"
echo "Compressed backup: ${BACKUP_FILE}.gz"
echo "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"

# Show backup info
echo ""
echo "To restore this backup:"
echo "  ./scripts/restore-k3s.sh $BACKUP_FILE"
