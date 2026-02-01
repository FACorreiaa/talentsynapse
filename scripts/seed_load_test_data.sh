#!/bin/bash

# Seed Load Test Data Script
# Applies load test data directly using psql

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DATABASE_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Seeding Load Test Data${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${YELLOW}Applying seed data migrations...${NC}"

# Apply migration 044 directly
echo "Applying 044_load_test_seed.sql..."
PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -f internal/database/migrations/044_load_test_seed.sql 2>&1 | grep -E '(INSERT|ERROR|DO)' | head -20

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Migration 044 applied${NC}"
else
    echo -e "${YELLOW}Note: Migration 044 may have already been applied or encountered conflicts${NC}"
fi

# Apply migration 045 directly
echo "Applying 045_load_test_sessions_reviews.sql..."
PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -f internal/database/migrations/045_load_test_sessions_reviews.sql 2>&1 | grep -E '(INSERT|ERROR|DO)' | head -20

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Migration 045 applied${NC}"
else
    echo -e "${YELLOW}Note: Migration 045 may have already been applied or encountered conflicts${NC}"
fi

# Verify data
echo ""
echo -e "${YELLOW}Verifying seed data...${NC}"
USER_COUNT=$(PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -t -c "SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com'" | xargs)
SESSION_COUNT=$(PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -t -c "SELECT COUNT(*) FROM sessions WHERE initiator_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com')" | xargs)
CONV_COUNT=$(PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -t -c "SELECT COUNT(*) FROM conversations WHERE user_a_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com')" | xargs)

echo "Load test users: $USER_COUNT"
echo "Test sessions: $SESSION_COUNT"
echo "Test conversations: $CONV_COUNT"

if [ "$USER_COUNT" -ge 100 ]; then
    echo -e "${GREEN}✓ Sufficient test data loaded${NC}"
else
    echo -e "${RED}✗ Insufficient data. Expected 100+ users, found $USER_COUNT${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}Seed data loaded successfully!${NC}"