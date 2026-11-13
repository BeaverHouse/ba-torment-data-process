#!/bin/sh
# A failing step ends the script with its exit code; set -e stops before the next step.
set -e

echo "============================================"
echo "BA Torment Data Processing"
echo "============================================"
echo ""

echo "[1/3] Updating student data from SchaleDB..."
echo "--------------------------------------------"
/app/bin/batorment update-from-schaledb
echo "✓ SchaleDB update completed successfully"

echo ""
echo "[2/3] Processing raid data..."
echo "--------------------------------------------"
/app/bin/batorment process-raid
echo "✓ Raid processing completed successfully"

echo ""
echo "[3/3] Running total analysis..."
echo "--------------------------------------------"
/app/bin/batorment total-analysis
echo "✓ Total analysis completed successfully"

echo ""
echo "============================================"
echo "All tasks completed successfully!"
echo "============================================"
