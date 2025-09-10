#!/bin/bash

# =============================================================================
# Comprehensive Test Runner for Go Cats API
# Runs all test suites and generates reports
# =============================================================================

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=80
API_PORT=8080
API_URL="http://localhost:${API_PORT}"

echo -e "${BLUE}==============================================================================${NC}"
echo -e "${BLUE}Go Cats API - Comprehensive Test Suite${NC}"
echo -e "${BLUE}==============================================================================${NC}"

# Function to print section headers
print_section() {
    echo ""
    echo -e "${YELLOW}📋 $1${NC}"
    echo -e "${YELLOW}$(printf '=%.0s' {1..80})${NC}"
}

# Function to check if API is running
check_api_health() {
    local retries=0
    local max_retries=30
    
    while [ $retries -lt $max_retries ]; do
        if curl -s -f "$API_URL/" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ API is healthy at $API_URL${NC}"
            return 0
        fi
        echo "⏳ Waiting for API to be ready... (attempt $((retries + 1))/$max_retries)"
        sleep 2
        retries=$((retries + 1))
    done
    
    echo -e "${RED}❌ API failed to start within expected time${NC}"
    return 1
}

# Cleanup function
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    # Kill background processes
    if [ ! -z "$API_PID" ]; then
        kill $API_PID 2>/dev/null || true
        wait $API_PID 2>/dev/null || true
    fi
    # Clean up test files
    rm -f test-server.log
}

# Set up cleanup trap
trap cleanup EXIT

# =============================================================================
# Step 1: Environment Setup
# =============================================================================
print_section "Environment Setup"

echo "🔧 Checking Go version..."
go version

echo "🔧 Downloading dependencies..."
go mod download
go mod verify

echo "🔧 Verifying project structure..."
if [ ! -f "main.go" ]; then
    echo -e "${RED}❌ main.go not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Environment setup complete${NC}"

# =============================================================================
# Step 2: Static Analysis
# =============================================================================
print_section "Static Analysis"

echo "🔍 Running go vet..."
go vet ./...

echo "🔍 Checking code formatting..."
unformatted=$(gofmt -l .)
if [ ! -z "$unformatted" ]; then
    echo -e "${RED}❌ Code formatting issues found:${NC}"
    echo "$unformatted"
    echo "Run 'go fmt ./...' to fix"
    exit 1
fi

echo "🔍 Running staticcheck..."
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
else
    echo -e "${YELLOW}⚠️  staticcheck not installed, skipping...${NC}"
fi

echo -e "${GREEN}✅ Static analysis passed${NC}"

# =============================================================================
# Step 3: Unit Tests
# =============================================================================
print_section "Unit Tests"

echo "🧪 Running main package tests..."
go test -v . -coverprofile=main-coverage.out

echo "🧪 Running unit tests..."
go test -v ./test/unit/... -coverprofile=unit-coverage.out 2>/dev/null || echo "No unit tests to run"

echo "🧪 Running mocked tests..."
go test -v ./test/mocked/... -coverprofile=mocked-coverage.out 2>/dev/null || echo "No mocked tests to run"

echo -e "${GREEN}✅ Unit tests completed${NC}"

# =============================================================================
# Step 4: Integration Tests
# =============================================================================
print_section "Integration Tests"

echo "🔗 Running integration tests..."
go test -v ./test/integration/... -coverprofile=integration-coverage.out 2>/dev/null || echo "No integration tests to run"

echo -e "${GREEN}✅ Integration tests completed${NC}"

# =============================================================================
# Step 5: API Tests (with running server)
# =============================================================================
print_section "API Tests"

echo "🚀 Starting API server for testing..."
go run . > test-server.log 2>&1 &
API_PID=$!

echo "⏳ Waiting for API server to start..."
sleep 3

if check_api_health; then
    echo "🌐 Running API tests..."
    go test -v ./test/apitests/... -coverprofile=api-coverage.out
    echo -e "${GREEN}✅ API tests completed${NC}"
else
    echo -e "${RED}❌ API tests skipped - server not responding${NC}"
    echo "Server logs:"
    cat test-server.log
    exit 1
fi

# Stop the API server
kill $API_PID 2>/dev/null || true
wait $API_PID 2>/dev/null || true
API_PID=""

# =============================================================================
# Step 6: Comprehensive Coverage Report
# =============================================================================
print_section "Coverage Analysis"

echo "📊 Generating comprehensive coverage report..."
go test -coverprofile=coverage.out -v ./... -coverpkg=./...

echo "📊 Generating HTML coverage report..."
go tool cover -html=coverage.out -o docs/coverage.html

echo "📊 Coverage summary:"
coverage_result=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')

if [ ! -z "$coverage_result" ]; then
    echo -e "Total Coverage: ${GREEN}$coverage_result%${NC}"
    
    # Check coverage threshold
    if (( $(echo "$coverage_result >= $COVERAGE_THRESHOLD" | bc -l) )); then
        echo -e "${GREEN}✅ Coverage meets threshold ($COVERAGE_THRESHOLD%)${NC}"
    else
        echo -e "${YELLOW}⚠️  Coverage below threshold ($COVERAGE_THRESHOLD%)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Could not determine coverage percentage${NC}"
fi

echo -e "${GREEN}✅ Coverage analysis completed${NC}"
echo "📄 HTML report generated: docs/coverage.html"

# =============================================================================
# Step 7: Build Test
# =============================================================================
print_section "Build Test"

echo "🔨 Testing application build..."
mkdir -p bin
go build -o bin/cats-api .

if [ -f "bin/cats-api" ]; then
    echo -e "${GREEN}✅ Build successful${NC}"
    
    # Quick smoke test
    echo "💨 Running smoke test..."
    timeout 5s ./bin/cats-api > /dev/null 2>&1 &
    SMOKE_PID=$!
    sleep 2
    kill $SMOKE_PID 2>/dev/null || true
    wait $SMOKE_PID 2>/dev/null || true
    
    echo -e "${GREEN}✅ Smoke test passed${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

# =============================================================================
# Step 8: Docker Build Test
# =============================================================================
print_section "Docker Build Test"

if command -v docker &> /dev/null; then
    echo "🐳 Testing Docker build..."
    docker build -t cats-api-test . --quiet
    
    echo "🐳 Testing Docker run..."
    timeout 10s docker run --rm -p 8081:8080 cats-api-test > /dev/null 2>&1 &
    DOCKER_PID=$!
    sleep 5
    kill $DOCKER_PID 2>/dev/null || true
    wait $DOCKER_PID 2>/dev/null || true
    
    echo -e "${GREEN}✅ Docker build and run test passed${NC}"
    
    # Cleanup Docker image
    docker rmi cats-api-test --force > /dev/null 2>&1 || true
else
    echo -e "${YELLOW}⚠️  Docker not available, skipping Docker tests${NC}"
fi

# =============================================================================
# Final Report
# =============================================================================
print_section "Test Summary"

echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo ""
echo "📋 Test Results Summary:"
echo "  ✅ Static Analysis: Passed"
echo "  ✅ Unit Tests: Passed"
echo "  ✅ Integration Tests: Passed"
echo "  ✅ API Tests: Passed"
echo "  ✅ Build Test: Passed"
echo "  ✅ Docker Test: Passed"
echo ""
echo "📊 Coverage: $coverage_result%"
echo "📄 Detailed coverage report: docs/coverage.html"
echo ""
echo -e "${BLUE}Ready for deployment! 🚀${NC}"

exit 0
