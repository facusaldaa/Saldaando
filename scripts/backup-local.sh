#!/bin/bash
# Backup local SQLite database

set -e

DB_PATH="./data/bot.db"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/bot.db.backup.${TIMESTAMP}"

if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file not found at $DB_PATH"
    exit 1
fi

mkdir -p "$BACKUP_DIR"

echo "Creating backup..."
cp "$DB_PATH" "$BACKUP_FILE"

# Also create a compressed backup
gzip -c "$BACKUP_FILE" > "${BACKUP_FILE}.gz"

echo "Backup created: $BACKUP_FILE"
echo "Compressed backup: ${BACKUP_FILE}.gz"
echo "Database size: $(du -h "$DB_PATH" | cut -f1)"
echo "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"
