# Task Completion Checklist

When completing a coding task in this project, follow this checklist:

## 1. Code Changes Complete

- [ ] All requested functionality implemented
- [ ] Code follows clean architecture layer separation
- [ ] Proper error handling in place
- [ ] Context propagation implemented

## 2. Wire Dependencies (if applicable)

If you modified dependency injection:

```bash
cd services/{service-name}
make generate
```

This regenerates `wire_gen.go` from `wire.go` definitions.

## 3. Proto Changes (if applicable)

If you modified `.proto` files:

```bash
# From repository root
make contracts-all
```

This formats, lints, and generates code from proto definitions.

## 4. Run Tests

Always run tests before considering a task complete:

```bash
cd services/{service-name}
make test
```

**All tests must pass.** If tests fail:
- Fix the failing tests
- Update tests if behavior intentionally changed
- Add new tests for new functionality

## 5. Code Quality Checks

### Linting
```bash
# Go vet (basic checks)
cd services/{service-name}
go vet ./...

# Proto linting (if proto files changed)
cd ../../
make contracts-lint
```

### Formatting
```bash
# Go files should already be formatted by editor
# If not, run:
cd services/{service-name}
gofmt -s -w .

# Proto files
cd ../../
make contracts-format
```

## 6. Build Verification

Ensure the service builds successfully:

```bash
cd services/{service-name}
make build
```

Binary should be created in `bin/` directory without errors.

## 7. Documentation (if applicable)

- [ ] Update comments on exported functions/types
- [ ] Update proto file comments if API changed
- [ ] Update CLAUDE.md if architectural changes made
- [ ] No need to create separate README files unless explicitly requested

## 8. Git Operations (if requested)

Only commit if the user explicitly asks:

```bash
git add .
git status  # Review changes
git commit -m "descriptive message"
```

## Quick Checklist Summary

For most tasks:

```bash
# 1. Make code changes
# 2. If dependencies changed:
cd services/{service-name} && make generate

# 3. If proto changed:
cd ../../ && make contracts-all && cd services/{service-name}

# 4. Run tests (REQUIRED)
make test

# 5. Build
make build
```

## What NOT to Do

- ❌ Don't skip running tests
- ❌ Don't commit without user request
- ❌ Don't add unnecessary documentation files
- ❌ Don't over-engineer (keep it simple)
- ❌ Don't add features not requested
- ❌ Don't modify code outside the requested scope
