#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directory - use temp dir to avoid polluting the repo
TEST_DIR=$(mktemp -d)
REPO_DIR=$(pwd)

echo -e "${YELLOW}Running workshop tests in: ${TEST_DIR}${NC}"

cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    cd "$REPO_DIR"
    # Stop any running flox services
    if [[ -d "$TEST_DIR/.flox" ]]; then
        cd "$TEST_DIR"
        flox delete --force 2>/dev/null || true
    fi
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

pass() {
    echo -e "${GREEN}✓ $1${NC}"
}

fail() {
    echo -e "${RED}✗ $1${NC}"
    exit 1
}

kill_port_3000() {
    lsof -ti:3000 | xargs kill -9 2>/dev/null || true
    sleep 1
}

# Kill any existing process on port 3000
kill_port_3000

# Copy project files to test directory
cp main.go main_test.go go.mod go.sum quotes.json "$TEST_DIR/"
cd "$TEST_DIR"

echo -e "\n${YELLOW}=== Lab 0: Explore our example app ===${NC}"

# Test that main.go exists
[[ -f main.go ]] && pass "main.go exists" || fail "main.go not found"

# Test that go is NOT available (outside flox)
if command -v go &>/dev/null && [[ -z "${FLOX_ENV:-}" ]]; then
    echo -e "${YELLOW}  (go is available system-wide, skipping 'not found' test)${NC}"
else
    pass "go command not found (as expected)"
fi

echo -e "\n${YELLOW}=== Lab 1: Your first Flox environment ===${NC}"

# Initialize flox
flox init --bare
pass "flox init --bare"

# Install go
flox install go
pass "flox install go"

# Run app with flox and test endpoints
flox activate -- go run main.go quotes.json &
APP_PID=$!
sleep 3

# Test endpoints
curl -s http://localhost:3000/ | grep -q "GET /quotes" && pass "GET / returns endpoints" || fail "GET / failed"
curl -s http://localhost:3000/quotes | grep -q "Steve Jobs" && pass "GET /quotes returns quotes" || fail "GET /quotes failed"
curl -s http://localhost:3000/quotes/0 | grep -q "great work" && pass "GET /quotes/0 returns first quote" || fail "GET /quotes/0 failed"

kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true
pass "App stopped"

echo -e "\n${YELLOW}=== Lab 2: Running a database ===${NC}"

kill_port_3000

# Install redis
flox install redis
pass "flox install redis"

# Update manifest with redis service
cat >> .flox/env/manifest.toml << 'EOF'

[services.redis]
command = "redis-server --port $REDISPORT"

[vars]
REDISPORT = "6379"

[profile]
common = """
  alias load_quotes='redis-cli -p $REDISPORT SET quotesjson "$(cat quotes.json)"'
"""
EOF
pass "Added redis service config"

# Create a helper script to load quotes and test redis
cat > test_redis.sh << 'TESTSCRIPT'
#!/usr/bin/env bash
set -e
sleep 2
redis-cli -p $REDISPORT SET quotesjson "$(cat quotes.json)"
echo "Quotes loaded into Redis"

go run main.go redis &
APP_PID=$!
sleep 3
RESULT=$(curl -s http://localhost:3000/quotes)
kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true
echo "$RESULT" | grep -q "Steve Jobs"
TESTSCRIPT
chmod +x test_redis.sh

# Activate with services, load quotes and test
flox activate --start-services -- ./test_redis.sh && pass "Redis service works with app" || fail "Redis test failed"

echo -e "\n${YELLOW}=== Lab 3: Reusing environments (composition) ===${NC}"

kill_port_3000

# Replace manifest with includes
cat > .flox/env/manifest.toml << 'EOF'
version = 1

[install]

[vars]

[include]
environments = [
    { remote = "flox/go" },
    { remote = "flox/redis" }
]
EOF
pass "Replaced manifest with includes"

# Create helper script to test composed environment
cat > test_composed.sh << 'TESTSCRIPT'
#!/usr/bin/env bash
sleep 2
redis-cli -p $REDISPORT SET quotesjson "$(cat quotes.json)"
go run main.go redis &
APP_PID=$!
sleep 3
RESULT=$(curl -s http://localhost:3000/quotes)
kill $APP_PID 2>/dev/null || true
echo "$RESULT" | grep -q "Steve Jobs"
TESTSCRIPT
chmod +x test_composed.sh

# Test with composed environment
flox activate --start-services -- ./test_composed.sh && pass "Composed environment works" || fail "Composed environment failed"

echo -e "\n${YELLOW}=== Lab 4: Prepare for production (build) ===${NC}"

kill_port_3000

# Add build section
cat >> .flox/env/manifest.toml << 'EOF'

[build.quotes-app]
command = """
  mkdir -p $out/bin $out/share
  cp quotes.json $out/share/
  go build -trimpath -o $out/bin/quotes-app main.go
"""
EOF
pass "Added build config"

# Build
flox build
pass "flox build succeeded"

# Test built binary
[[ -x ./result-quotes-app/bin/quotes-app ]] && pass "Binary exists and is executable" || fail "Binary not found"

./result-quotes-app/bin/quotes-app ./result-quotes-app/share/quotes.json &
APP_PID=$!
sleep 2

curl -s http://localhost:3000/quotes | grep -q "Steve Jobs" && pass "Built binary serves quotes" || fail "Built binary failed"

kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true

# Skip publish (requires authentication)
echo -e "${YELLOW}  (skipping flox publish - requires authentication)${NC}"

echo -e "\n${GREEN}=== All workshop tests passed! ===${NC}"
