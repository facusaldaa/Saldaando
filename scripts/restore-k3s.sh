#!/bin/bash
# Restore SQLite database to k3s pod

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <backup-file> [pod-name]"
    echo "Example: $0 ./backups/bot.db.k3s.backup.20240101_120000"
    exit 1
fi

BACKUP_FILE="$1"
NAMESPACE="${NAMESPACE:-default}"
RELEASE_NAME="${RELEASE_NAME:-botgastospareja}"
POD_NAME="${2:-}"
DB_PATH="/data/bot.db"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    # Try with .gz extension
    if [ -f "${BACKUP_FILE}.gz" ]; then
        echo "Decompressing ${BACKUP_FILE}.gz..."
        gunzip -c "${BACKUP_FILE}.gz" > "${BACKUP_FILE}.tmp"
        BACKUP_FILE="${BACKUP_FILE}.tmp"
    else
        echo "Error: Backup file not found: $BACKUP_FILE"
        exit 1
    fi
fi

# Get pod name if not provided
if [ -z "$POD_NAME" ]; then
    POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=${RELEASE_NAME}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [ -z "$POD_NAME" ]; then
        echo "Error: Could not find pod. Try:"
        echo "  kubectl get pods -n $NAMESPACE"
        exit 1
    fi
fi

echo "WARNING: This will overwrite the database in pod: $POD_NAME"
echo "Database will be replaced with: $BACKUP_FILE"
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Restore cancelled."
    [ -f "${BACKUP_FILE}.tmp" ] && rm -f "${BACKUP_FILE}.tmp"
    exit 0
fi

echo "Stopping pod to ensure clean restore..."
kubectl scale deployment -n "$NAMESPACE" "${RELEASE_NAME}" --replicas=0

echo "Waiting for pod to terminate..."
kubectl wait --for=delete pod -n "$NAMESPACE" "$POD_NAME" --timeout=60s || true

echo "Copying backup to pod..."
# We need to copy to a temp location first, then move it
TEMP_BACKUP="/tmp/bot.db.restore"
kubectl cp "$BACKUP_FILE" "${NAMESPACE}/${POD_NAME}:${TEMP_BACKUP}" || {
    echo "Error: Could not copy backup. Starting pod again..."
    kubectl scale deployment -n "$NAMESPACE" "${RELEASE_NAME}" --replicas=1
    [ -f "${BACKUP_FILE}.tmp" ] && rm -f "${BACKUP_FILE}.tmp"
    exit 1
}

echo "Moving backup to database location..."
kubectl exec -n "$NAMESPACE" "$POD_NAME" -- sh -c "mv $TEMP_BACKUP $DB_PATH && chmod 644 $DB_PATH"

echo "Starting pod..."
kubectl scale deployment -n "$NAMESPACE" "${RELEASE_NAME}" --replicas=1

echo "Restore complete!"
[ -f "${BACKUP_FILE}.tmp" ] && rm -f "${BACKUP_FILE}.tmp"
