#!/bin/sh
set -e

echo "============================================"
echo "BA Torment Data Processing"
echo "============================================"
echo ""

# Step 1: Update SchaleDB data
echo "[1/2] Updating student data from SchaleDB..."
echo "--------------------------------------------"
/app/bin/update_from_schaledb
if [ $? -eq 0 ]; then
    echo "✓ SchaleDB update completed successfully"
else
    echo "✗ SchaleDB update failed with exit code $?"
    exit 1
fi

echo ""
echo "[2/2] Processing raid data..."
echo "--------------------------------------------"
/app/bin/process_raid
if [ $? -eq 0 ]; then
    echo "✓ Raid processing completed successfully"
else
    echo "✗ Raid processing failed with exit code $?"
    exit 1
fi

echo ""
echo "============================================"
echo "All tasks completed successfully!"
echo "============================================"
