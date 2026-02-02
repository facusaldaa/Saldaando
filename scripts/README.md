# Database Backup Scripts

## Quick Backup (k3s)

**Before your big change, run:**

```bash
./scripts/backup-k3s.sh
```

This will:
- Find your pod automatically
- Copy the database from `/data/bot.db` in the pod
- Save it to `./backups/bot.db.k3s.backup.TIMESTAMP`
- Create a compressed `.gz` version

## Local Backup

If you have the database file locally:

```bash
./scripts/backup-local.sh
```

## Restore from Backup

```bash
./scripts/restore-k3s.sh ./backups/bot.db.k3s.backup.20240101_120000
```

**Warning:** This will stop the pod, replace the database, and restart it.

## Manual k3s Backup (one-liner)

If you prefer a quick manual backup:

```bash
# Get pod name
POD=$(kubectl get pods -l app.kubernetes.io/name=botgastospareja -o jsonpath='{.items[0].metadata.name}')

# Copy database
kubectl cp default/$POD:/data/bot.db ./backups/bot.db.$(date +%Y%m%d_%H%M%S)
```

## Environment Variables

You can override defaults:
- `NAMESPACE` (default: `default`)
- `RELEASE_NAME` (default: `botgastospareja`)
- `POD_NAME` (auto-detected if not set)

Example:
```bash
NAMESPACE=production RELEASE_NAME=mybot ./scripts/backup-k3s.sh
```
