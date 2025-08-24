#!/bin/bash

# Test runner script for Exim Control Panel
# This script runs all tests including integration, performance, and frontend tests

set -e

echo "🚀 Starting Exim Control Panel Test Suite"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go to run backend tests."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js to run frontend tests."
    exit 1
fi

# Check if Playwright is installed
if ! command -v npx playwright --version &> /dev/null; then
    print_warning "Playwright is not installed. Installing Playwright..."
    cd web && npm install @playwright/test && npx playwright install
    cd ..
fi

# Set test environment variables
export TEST_ENV=true
export DATABASE_PATH=":memory:"
export LOG_LEVEL=error

# Function to run Go tests
run_go_tests() {
    print_status "Running Go integration tests..."
    
    if go test -v ./tests/integration/... -timeout 30s; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        return 1
    fi
    
    print_status "Running Go performance tests..."
    
    if go test -v ./tests/performance/... -timeout 60s -bench=. -benchmem; then
        print_success "Performance tests completed"
    else
        print_error "Performance tests failed"
        return 1
    fi
}

# Function to run frontend tests
run_frontend_tests() {
    print_status "Running frontend tests with Playwright..."
    
    cd web
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        print_status "Installing frontend dependencies..."
        npm install
    fi
    
    # Build frontend for testing
    print_status "Building frontend for testing..."
    npm run build
    
    # Start development server in background
    print_status "Starting development server..."
    npm run dev &
    DEV_SERVER_PID=$!
    
    # Wait for server to start
    sleep 5
    
    # Run Playwright tests
    if npx playwright test ../tests/frontend/ --reporter=html; then
        print_success "Frontend tests passed"
        TEST_RESULT=0
    else
        print_error "Frontend tests failed"
        TEST_RESULT=1
    fi
    
    # Stop development server
    kill $DEV_SERVER_PID 2>/dev/null || true
    
    cd ..
    return $TEST_RESULT
}

# Function to run performance benchmarks
run_performance_benchmarks() {
    print_status "Running performance benchmarks..."
    
    # Create benchmark results directory
    mkdir -p test_results/benchmarks
    
    # Run Go benchmarks with output
    go test -v ./tests/performance/... -bench=. -benchmem -benchtime=10s > test_results/benchmarks/go_benchmarks.txt 2>&1
    
    print_success "Performance benchmarks completed. Results saved to test_results/benchmarks/"
}

# Function to generate test report
generate_test_report() {
    print_status "Generating test report..."
    
    mkdir -p test_results
    
    cat > test_results/test_summary.md << EOF
# Exim Control Panel Test Results

## Test Execution Summary

**Date:** $(date)
**Environment:** Test
**Go Version:** $(go version)
**Node Version:** $(node --version)

## Test Categories

### ✅ Integration Tests
- API endpoint testing
- Authentication flow testing
- Database operations testing
- Error handling testing

### ⚡ Performance Tests
- Database operation benchmarks
- Log processing benchmarks
- Streaming processor benchmarks
- Large dataset performance tests
- Concurrent operation tests

### 🎭 Frontend Tests (Playwright)
- Queue management interface testing
- Log viewer interface testing
- Virtual scrolling performance testing
- Real-time update testing
- Error state handling testing

## Performance Benchmarks

See detailed benchmark results in \`benchmarks/go_benchmarks.txt\`

## Frontend Test Results

Playwright test results are available in the HTML report.

## Recommendations

1. **Database Optimization**: Ensure proper indexing for large datasets
2. **Frontend Performance**: Use virtual scrolling for large lists
3. **Real-time Updates**: Implement efficient WebSocket handling
4. **Error Handling**: Comprehensive error state management
5. **Caching**: Implement appropriate caching strategies

EOF

    print_success "Test report generated at test_results/test_summary.md"
}

# Main execution
main() {
    local exit_code=0
    
    # Create test results directory
    mkdir -p test_results
    
    # Run Go tests
    if ! run_go_tests; then
        exit_code=1
    fi
    
    # Run performance benchmarks
    run_performance_benchmarks
    
    # Run frontend tests
    if ! run_frontend_tests; then
        exit_code=1
    fi
    
    # Generate test report
    generate_test_report
    
    if [ $exit_code -eq 0 ]; then
        print_success "All tests completed successfully! 🎉"
        echo ""
        echo "📊 Test Results:"
        echo "  - Integration Tests: ✅ Passed"
        echo "  - Performance Tests: ✅ Completed"
        echo "  - Frontend Tests: ✅ Passed"
        echo ""
        echo "📁 Results saved to: test_results/"
    else
        print_error "Some tests failed. Please check the output above."
        echo ""
        echo "📊 Test Results:"
        echo "  - Some tests failed ❌"
        echo ""
        echo "📁 Results saved to: test_results/"
    fi
    
    exit $exit_code
}

# Handle script arguments
case "${1:-all}" in
    "integration")
        print_status "Running integration tests only..."
        run_go_tests
        ;;
    "performance")
        print_status "Running performance tests only..."
        run_performance_benchmarks
        ;;
    "frontend")
        print_status "Running frontend tests only..."
        run_frontend_tests
        ;;
    "all"|*)
        print_status "Running all tests..."
        main
        ;;
esac