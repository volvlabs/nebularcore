#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== WebSocket Load Test ==="
echo "Starting server and k6 load test..."

# Build and start services.
docker compose up --build --abort-on-container-exit --exit-code-from k6

echo ""
echo "=== Results ==="
if [ -f k6/results.json ]; then
    echo "Results saved to tests/load/k6/results.json"
    cat k6/results.json | python3 -m json.tool 2>/dev/null || cat k6/results.json
fi

echo ""
echo "Cleaning up..."
docker compose down --volumes --remove-orphans
