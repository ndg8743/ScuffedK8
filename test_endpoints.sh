#!/bin/bash

echo "=========================================="
echo "ScuffedK8 Endpoint Tests"
echo "=========================================="
echo ""

# Test 1: Root endpoint
echo "1. Testing Root Endpoint"
echo "   GET /"
curl -s http://localhost:8080/
echo ""
echo ""

# Test 2: Health check
echo "2. Testing Health Endpoint"
echo "   GET /health"
curl -s http://localhost:8080/health | jq .
echo ""

# Test 3: Nodes status
echo "3. Testing Nodes Status Endpoint"
echo "   GET /nodes"
curl -s http://localhost:8080/nodes | jq .
echo ""

# Test 4: CPU Workload - Light
echo "4. Testing CPU Workload (Light - 1M iterations)"
echo "   GET /workload?iterations=1000000"
START=$(date +%s)
curl -s "http://localhost:8080/workload?iterations=1000000" | jq .
END=$(date +%s)
echo "   Actual time: $((END - START))s"
echo ""

# Test 5: CPU Workload - Medium
echo "5. Testing CPU Workload (Medium - 10M iterations)"
echo "   GET /workload?iterations=10000000"
START=$(date +%s)
curl -s "http://localhost:8080/workload?iterations=10000000" | jq .
END=$(date +%s)
echo "   Actual time: $((END - START))s"
echo ""

# Test 6: CPU Workload - Heavy
echo "6. Testing CPU Workload (Heavy - 50M iterations)"
echo "   GET /workload?iterations=50000000"
START=$(date +%s)
curl -s "http://localhost:8080/workload?iterations=50000000" | jq .
END=$(date +%s)
echo "   Actual time: $((END - START))s"
echo ""

echo "=========================================="
echo "All tests completed!"
echo "=========================================="
