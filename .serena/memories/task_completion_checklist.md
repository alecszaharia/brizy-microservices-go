# Task Completion Checklist

When completing a coding task in this project, follow this checklist to ensure quality and consistency.

## 1. Code Changes Complete

- [ ] All requested functionality implemented
- [ ] Code follows clean architecture layer separation (service → biz → data)
- [ ] Proper error handling in place (return errors, wrap with context)
- [ ] Context propagation implemented (ctx as first parameter)
- [ ] No layer boundaries violated
- [ ] Interfaces used where appropriate

## 2. Wire Dependencies (if applicable)

If you modified dependency injection (added/removed constructors, changed ProviderSets):

```bash
cd services/{service-name}
make generate
```

This regenerates `wire_gen.go` from `wire.go` definitions.

**Verify**: Check that `wire_gen.go` was regenerated without errors.

## 3. Proto Changes (if applicable)

If you modified `.proto` files in `api/{service}/v1/`:

```bash
# From repository root
make contracts-all
```

This:
- Formats proto files (`buf format`)
- Lints proto files (`buf lint`)
- Checks for breaking changes (`buf breaking`)
- Generates Go code (gRPC, HTTP, validation, OpenAPI)

**Verify**: Check that generated code in `contracts/{service}/v1/` was updated.

## 4. Run Tests (MANDATORY)

**Always run tests before considering a task complete:**

```bash
cd services/{service-name}
make test
```

**All tests must pass.** If tests fail:
- Fix the failing tests
- Update tests if behavior was intentionally changed
- Add new tests for new functionality

### Writing Tests
- Add unit tests for new business logic in `biz/`
- Use table-driven tests with `testify/assert` and `testify/mock`
- Test both success and error cases
- Mock repository interfaces, not implementations

## 5. Code Quality Checks

### Linting
```bash
# Go vet (basic static analysis)
cd services/{service-name}
go vet ./...

# Proto linting (if proto files changed)
cd ../../  # back to root
make contracts-lint
```

### Formatting
```bash
# Go files (should be auto-formatted by editor)
# If not, run:
cd services/{service-name}
gofmt -s -w .

# Proto files
cd ../../  # back to root
make contracts-format
```

## 6. Build Verification

Ensure the service builds successfully:

```bash
cd services/{service-name}
make build
```

Binary should be created in `bin/{service-name}` directory without errors.

**Verify**: Run the binary to ensure it starts:
```bash
./bin/{service-name} -conf configs/config.yaml
```
(Press Ctrl+C to stop)

## 7. Documentation (if applicable)

Update documentation only for significant changes:

- [ ] Updated comments on exported functions/types
- [ ] Updated proto file comments if API changed
- [ ] Updated CLAUDE.md if architectural patterns changed
- [ ] Updated service README.md if deployment/setup changed
- [ ] **Do NOT** create unnecessary documentation files

**Note**: As per CLAUDE.md, always update documentation for big architectural changes that are important for developers to know.

## 8. Commit Guidelines (if committing)

Only commit if the user explicitly asks:

```bash
# Review changes
git status
git diff

# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add symbol CRUD operations"

# Follow conventional commits:
# - feat: new feature
# - fix: bug fix
# - refactor: code refactoring
# - test: adding tests
# - docs: documentation changes
# - chore: maintenance tasks
```

## Quick Checklist Summary

For most tasks, follow this quick workflow:

```bash
# 1. Make code changes

# 2. If dependencies changed (wire.go modified):
cd services/{service-name}
make generate

# 3. If proto files changed:
cd ../../  # back to root
make contracts-all
cd services/{service-name}

# 4. Run tests (REQUIRED - must pass)
make test

# 5. Verify build
make build

# 6. (Optional) Run the service locally
./bin/{service-name} -conf configs/config.yaml
```

## What NOT to Do

- ❌ Don't skip running tests (tests are MANDATORY)
- ❌ Don't commit without explicit user request
- ❌ Don't add unnecessary documentation files
- ❌ Don't over-engineer (keep solutions simple)
- ❌ Don't add features not requested
- ❌ Don't modify code outside the requested scope
- ❌ Don't manually edit generated files (`wire_gen.go`, `*.pb.go`, etc.)
- ❌ Don't violate layer boundaries (service → biz → data)
- ❌ Don't ignore errors
- ❌ Don't skip validation (both transport and business validation)

## CI/CD Requirements

The CI pipeline enforces:
- Proto files must be formatted (`buf format`)
- Proto files must pass linting (`buf lint`)
- Changes must not break API contracts (`buf breaking`)
- All tests must pass (`go test`)
- Code must build successfully (`go build`)

**Always run these checks locally before pushing** to avoid CI failures.

## Optional Checks

### Coverage Check
```bash
cd services/{service-name}
make coverage
# Opens coverage report in browser
```

Aim for high coverage on business logic (biz layer).

### Race Detection
Already included in `make test`, but can run separately:
```bash
cd services/{service-name}
go test -race ./...
```

### Full Clean Build
If you want to verify everything from scratch:
```bash
# From root
make contracts-all

cd services/{service-name}
make generate
make test
make build
```

## Troubleshooting

### Wire Generation Fails
- Check `wire.go` for syntax errors
- Ensure all ProviderSets are imported
- Verify constructors return correct types
- Run with verbose: `wire gen ./cmd/{service-name}`

### Proto Generation Fails
- Check proto syntax: `buf lint`
- Verify imports are correct
- Ensure `buf.yaml` and `buf.gen.yaml` are valid
- Check for breaking changes: `buf breaking --against '.git#branch=main'`

### Tests Fail
- Read error messages carefully
- Check if mocks are set up correctly
- Verify test data matches expectations
- Run single test with `-v` flag: `go test -v -run TestName`

### Build Fails
- Run `go mod tidy` to clean dependencies
- Check for import errors
- Verify all packages are available
- Run `go get` for missing dependencies
