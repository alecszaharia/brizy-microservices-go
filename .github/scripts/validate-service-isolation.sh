#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Validating service isolation..."
echo "========================================="
echo ""

# Track violations
VIOLATIONS_FOUND=0
VIOLATION_DETAILS=""

# Get workspace root (script should run from repo root)
REPO_ROOT="$(pwd)"

# Parse go.work to get all modules
if [ ! -f "go.work" ]; then
  echo -e "${RED}ERROR: go.work file not found${NC}"
  exit 1
fi

# Extract service directories and module names
echo "Discovering services from go.work..."

# Parse go.work to find service paths
SERVICE_PATHS=$(grep -A 100 "^use (" go.work | grep -E "^\s*\./services/" | sed 's/^[[:space:]]*//' | sed 's#^./##' || true)

if [ -z "$SERVICE_PATHS" ]; then
  echo -e "${YELLOW}No services found in go.work - skipping validation${NC}"
  exit 0
fi

# Build list of service module names and paths (bash 3.2 compatible)
# Format: "module_name|service_path"
SERVICE_LIST=""

for service_path in $SERVICE_PATHS; do
  if [ -f "$service_path/go.mod" ]; then
    module_name=$(grep "^module " "$service_path/go.mod" | awk '{print $2}')
    SERVICE_LIST="${SERVICE_LIST}${module_name}|${service_path}"$'\n'
    echo "  Found service: $module_name ($service_path)"
  fi
done

# Count services
SERVICE_COUNT=$(echo "$SERVICE_LIST" | grep -c '|' || echo "0")

echo ""
echo "Validating imports for $SERVICE_COUNT services..."
echo ""

# Validate each service
echo "$SERVICE_LIST" | while IFS='|' read -r current_module current_path; do
  if [ -z "$current_module" ]; then
    continue
  fi

  echo "Checking service: $current_module"

  # Change to service directory
  cd "$REPO_ROOT/$current_path"

  # Get all packages and their imports using go list
  # Format: package_path import_path
  IMPORTS=$(go list -f '{{$pkg := .ImportPath}}{{range .Imports}}{{$pkg}} {{.}}{{"\n"}}{{end}}' ./... 2>/dev/null || true)

  # Check each import
  echo "$IMPORTS" | while IFS= read -r line; do
    if [ -z "$line" ]; then
      continue
    fi

    # Parse package and import
    pkg=$(echo "$line" | awk '{print $1}')
    import=$(echo "$line" | awk '{print $2}')

    # Check if import is from another service
    echo "$SERVICE_LIST" | while IFS='|' read -r other_module other_path; do
      if [ -z "$other_module" ]; then
        continue
      fi

      # Skip checking against self
      if [ "$other_module" = "$current_module" ]; then
        continue
      fi

      # Check if import starts with other service module name
      if [[ "$import" == "$other_module" ]] || [[ "$import" == "$other_module/"* ]]; then
        # Get relative package path
        pkg_rel="${pkg#$current_module/}"
        if [ "$pkg_rel" = "$pkg" ]; then
          pkg_rel="."
        fi

        # Find Go files in this package
        go_files=$(find "$pkg_rel" -maxdepth 1 -name "*.go" ! -name "*_test.go" 2>/dev/null || true)

        # Search for the import in these files
        violating_files=""
        for go_file in $go_files; do
          if grep -q "\"$import\"" "$go_file" 2>/dev/null; then
            violating_files="$violating_files\n    $current_path/$go_file"
          fi
        done

        # Report violation immediately
        echo ""
        echo -e "${RED}ERROR: Cross-service import violation detected!${NC}"
        echo -e "  Service: ${YELLOW}$current_module${NC}"
        echo -e "  Package: ${YELLOW}$pkg${NC}"
        echo -e "  Illegal import: ${RED}$import${NC}"
        echo -e "  Target service: ${YELLOW}$other_module${NC} ($other_path)"
        if [ -n "$violating_files" ]; then
          echo -e "  Files with this import:${violating_files}"
        fi

        # Exit immediately on first violation
        cd "$REPO_ROOT"
        echo ""
        echo "========================================="
        echo -e "${YELLOW}Why this matters:${NC}"
        echo "  Services must be isolated to maintain clean microservice architecture."
        echo "  Cross-service imports create tight coupling."
        echo ""
        echo -e "${YELLOW}How to fix:${NC}"
        echo "  1. Remove imports from other services' internal packages"
        echo "  2. Use shared 'contracts' module for API definitions (protobuf)"
        echo "  3. Use shared 'platform' module for common utilities"
        echo "  4. Communicate between services via gRPC/HTTP APIs only"
        echo ""
        echo -e "${YELLOW}Valid imports:${NC}"
        echo "  ✓ contracts/...      - Shared API definitions"
        echo "  ✓ platform/...       - Shared utilities"
        echo "  ✓ {same-service}/... - Internal packages"
        echo "  ✓ github.com/...     - External dependencies"
        echo ""
        echo -e "${RED}Invalid imports:${NC}"
        echo "  ✗ {other-service}/... - Any other service"
        echo "========================================="
        exit 1
      fi
    done
  done

  cd "$REPO_ROOT"
done

echo -e "${GREEN}✓ All services maintain proper isolation${NC}"
echo "  No cross-service imports detected"
echo "========================================="
