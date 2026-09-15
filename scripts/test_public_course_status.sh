#!/usr/bin/env bash
# test_public_course_status.sh
# Smoke tests for GET /api/public/courses/:courseId status filtering.
#
# Requires:
#   - internal/db/seeds/seed_public_course_status_test.sql loaded
#   - server running (just dev or just run)
#
# Usage:
#   BASE_URL=http://localhost:8080 bash scripts/test_public_course_status.sh
#
# To load the seed (docker compose running):
#   docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" \
#     -f /dev/stdin < internal/db/seeds/seed_public_course_status_test.sql

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

check() {
    local desc="$1" id="$2" expected="$3"
    local got
    got=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/public/courses/$id")
    if [ "$got" = "$expected" ]; then
        echo -e "  ${GREEN}PASS${NC}: $desc (got $got)"
        PASS=$((PASS + 1))
    else
        echo -e "  ${RED}FAIL${NC}: $desc (expected $expected, got $got)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Public course detail — status visibility ==="
echo "    BASE_URL: $BASE_URL"

echo ""
echo "--- Should return 200 (publicly visible) ---"
check "published      (id=910)" 910 200
check "closed         (id=916)" 916 200

echo ""
echo "--- Should return 404 (pipeline / non-public statuses) ---"
check "draft           (id=911)" 911 404
check "in_review       (id=912)" 912 404
check "needs_changes   (id=913)" 913 404
check "approved        (id=914)" 914 404
check "pending_deletion(id=915)" 915 404

echo ""
echo "--- Non-existent course ---"
check "not found (id=99999)" 99999 404

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
