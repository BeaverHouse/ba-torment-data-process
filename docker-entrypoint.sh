#!/bin/sh
set -e

echo "============================================"
echo "BA Torment Data Processing"
echo "============================================"
echo ""

# Step 1: Update SchaleDB data
echo "[1/3] Updating student data from SchaleDB..."
echo "--------------------------------------------"
/app/bin/update_from_schaledb
if [ $? -eq 0 ]; then
    echo "✓ SchaleDB update completed successfully"
else
    echo "✗ SchaleDB update failed with exit code $?"
    exit 1
fi

echo ""
echo "[2/3] Processing raid data..."
echo "--------------------------------------------"
/app/bin/process_raid
if [ $? -eq 0 ]; then
    echo "✓ Raid processing completed successfully"
else
    echo "✗ Raid processing failed with exit code $?"
    exit 1
fi

echo ""
echo "[3/3] Running total analysis..."
echo "--------------------------------------------"
/app/bin/total_analysis
if [ $? -eq 0 ]; then
    echo "✓ Total analysis completed successfully"
else
    echo "✗ Total analysis failed with exit code $?"
    exit 1
fi

echo ""
echo "============================================"
echo "All tasks completed successfully!"
echo "============================================"
