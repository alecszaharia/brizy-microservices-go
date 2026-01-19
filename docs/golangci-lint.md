# golangci-lint Configuration

This document describes the golangci-lint setup for the brizy-go-services monorepo.

## Installation

Install golangci-lint >=2.8.0 using the root Makefile:

```bash
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.8.0

# or install it into ./bin/
curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.8.0
```

## Usage

### Running Lint Checks

**Lint platform module:**
```bash
make lint-platform  # from root
# or
cd platform && make lint
```

**Lint specific service:**
```bash
cd services/symbols && make lint
```

**Lint all modules:**
```bash
make lint-all  # Lints platform + all services
```

**Auto-fix issues:**
```bash
make lint-platform-fix      # Fix platform code
cd services/symbols && make lint-fix  # Fix service code
make lint-all-fix           # Fix all modules
```

### Module-Specific Commands

Each module supports local lint commands:

```bash
# Platform module
cd platform
make lint          # Run linting
make lint-fix      # Run with auto-fix

# Service module
cd services/symbols
make lint          # Run linting
make lint-fix      # Run with auto-fix
```

**Note:** Running `make lint` or `make lint-fix` at the repository root will show an error message with guidance, as these commands are ambiguous in a monorepo context.

## Configuration

### Configuration Architecture

The monorepo uses a **hierarchical configuration structure**:

- **Platform Module**: `.golangci.yml` at repository root (applies to `platform/` only)
- **Services**: Each service has its own `.golangci.yml` in the service directory
  - `services/symbols/.golangci.yml` - Symbols service configuration
- **Contracts**: No configuration (generated code, automatically excluded)

This structure allows:
- Platform code maintains strict, consistent linting as foundation code
- Services can evolve different requirements independently
- Clear separation of concerns across modules
- Scalable pattern for future services

### Enabled Linters

The configuration enables a curated set of linters organized by category:

**Bugs & Correctness:**
- `errcheck` - Unchecked errors
- `gosimple` - Code simplification
- `govet` - Standard Go vet checks
- `ineffassign` - Ineffectual assignments
- `staticcheck` - Advanced static analysis
- `unused` - Unused code detection
- `bodyclose` - HTTP body close checks
- `nilerr` - Nil error detection
- `nilnil` - Nil with nil error
- `noctx` - Missing context in HTTP requests

**Style & Formatting:**
- `gofmt` - Go formatting
- `goimports` - Import formatting
- `misspell` - Spelling checks
- `whitespace` - Whitespace checks
- `nolintlint` - nolint directive validation

**Complexity:**
- `gocyclo` - Cyclomatic complexity (max: 15)
- `gocognit` - Cognitive complexity (max: 20)
- `funlen` - Function length (100 lines, 50 statements)
- `nestif` - Nested if depth (max: 5)

**Performance:**
- `prealloc` - Slice preallocation

**Error Handling:**
- `errname` - Error naming conventions
- `errorlint` - Error wrapping
- `wrapcheck` - Error wrapping validation

**Code Quality:**
- `goconst` - Repeated strings
- `gocritic` - Opinionated checks
- `revive` - Fast linter with many rules
- `unconvert` - Unnecessary conversions
- `unparam` - Unused parameters
- `wastedassign` - Wasted assignments

**Security:**
- `gosec` - Security issues

**SQL:**
- `rowserrcheck` - SQL rows.Err()
- `sqlclosecheck` - SQL Close() calls

### Exclusions

The following are automatically excluded from linting:

**Directories:**
- `vendor/`
- `third_party/`
- `testdata/`
- `.idea/`
- `.github/`
- `.claude/`

**Files:**
- `*.pb.go` - Generated protobuf files
- `*.pb.gw.go` - Generated gRPC gateway files
- `*_gen.go` - All generated files
- `wire_gen.go` - Wire dependency injection
- `mock_*.go` - Mock files for testing

### Test File Exceptions

Test files (`*_test.go`) have relaxed rules:
- No cyclomatic complexity checks
- No error checking requirements
- No security checks (gosec)
- No function length limits
- No constant extraction requirements

### Special Cases

**Wire files (`wire.go`):**
- Unused parameter checks disabled (DI patterns)

**Server setup files (`internal/server/`):**
- Function length limits relaxed
- Complexity checks relaxed

**Data layer (`internal/data/`):**
- Function length limits relaxed (complex queries)

**Main files (`cmd/*/main.go`):**
- Some gosec checks disabled (globals allowed)

### Managing Service Configurations

When creating a new service, follow this pattern:

1. **Copy baseline configuration:**
   ```bash
   cp .golangci.yml services/new-service/.golangci.yml
   ```

2. **Update header comment:**
   ```yaml
   # golangci-lint configuration for new-service
   # Based on platform configuration with service-specific customizations
   # https://golangci-lint.run/usage/configuration/
   ```

3. **Remove platform-specific directives:**
   - Delete the `skip-dirs` section (not needed in service configs)
   - Update `exclude-dirs` to only include standard exclusions

4. **Add service-specific Makefile targets:**
   ```makefile
   lint:
       @echo "Running golangci-lint on new-service..."
       golangci-lint run ./...

   lint-fix:
       @echo "Running golangci-lint with auto-fix on new-service..."
       golangci-lint run --fix ./...
   ```

5. **Customize as needed:**
   - Services can adjust linter settings independently
   - Document any deviations from platform baseline
   - Keep changes minimal to maintain consistency

**Guidelines:**
- Start with platform baseline configuration
- Only customize when service requirements genuinely differ
- Document reasons for service-specific settings
- Periodically sync with platform updates for bug fixes

## IDE Integration

### GoLand / IntelliJ IDEA

1. Go to **Settings** → **Tools** → **File Watchers**
2. Click **+** and select **golangci-lint**
3. Set **Program**: `golangci-lint`
4. Set **Arguments**: `run --config=$ProjectFileDir$/.golangci.yml $FilePath$`
5. Set **Working directory**: `$ProjectFileDir$`

Alternatively, use the built-in golangci-lint integration:
1. Go to **Settings** → **Tools** → **Go** → **golangci-lint**
2. Enable golangci-lint
3. Set configuration file path to `.golangci.yml`

### VS Code

Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go) and add to settings.json:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--config=${workspaceFolder}/.golangci.yml",
    "--fast"
  ],
  "go.lintOnSave": "workspace"
}
```

## CI/CD Integration

The linting should be integrated into CI/CD pipelines. With the hierarchical configuration, lint each module separately.

**Example GitHub Actions workflow:**

```yaml
# Lint platform module
- name: Lint platform
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    working-directory: platform
    args: --config=../.golangci.yml

# Lint services
- name: Lint symbols service
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    working-directory: services/symbols
    args: --config=.golangci.yml
```

**Alternative: Use Makefile targets**

```yaml
- name: Lint all modules
  run: make lint-all
```

This approach:
- Lints platform with root config
- Lints each service with its own config
- Skips contracts (generated code)
- Fails the build if any module has issues

## Ignoring Issues

### Inline Comments

Disable specific linters for a line:
```go
var bad = "example" //nolint:goconst // Justification here
```

Disable all linters for a line:
```go
var bad = example() //nolint:all // Must be ignored because...
```

Disable specific linters for a function:
```go
//nolint:gocyclo,funlen // Complex but necessary
func processComplexData() {
    // ...
}
```

### Configuration Changes

For project-wide exclusions, update `.golangci.yml`:

```yaml
issues:
  exclude-rules:
    - path: my/specific/file.go
      linters:
        - errcheck
```

## Troubleshooting

### "Too many issues" Error

If you see "too many issues", fix critical issues first or increase limits in `.golangci.yml`:

```yaml
issues:
  max-issues-per-linter: 50
  max-same-issues: 10
```

### Slow Performance

For faster runs during development:

```bash
golangci-lint run --fast
```

### Cache Issues

Clear the cache if you encounter strange behavior:

```bash
golangci-lint cache clean
```

## Best Practices

1. **Run before committing:** `make lint-fix` auto-fixes many issues
2. **Fix issues incrementally:** Don't disable linters globally
3. **Justify nolint directives:** Always add a reason
4. **Keep config updated:** Review and update linter settings periodically
5. **Address root causes:** Fix patterns, not just symptoms

## Resources

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Linter Descriptions](https://golangci-lint.run/usage/linters/)
- [Configuration Reference](https://golangci-lint.run/usage/configuration/)