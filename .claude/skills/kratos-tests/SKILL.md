---
name: kratos-tests
description: Write comprehensive tests for go-kratos microservices following Clean Architecture patterns. Analyzes existing test patterns, generates table-driven tests with testify/mock, and validates test execution. ALWAYS use this skill whenever you need to write or add tests in this project.
---

<objective>
Generates comprehensive table-driven tests for go-kratos microservices following Clean Architecture patterns. Analyzes existing test patterns in the codebase, creates layer-specific tests (data/biz/service) with testify/mock, and validates test execution.
</objective>

<quick_start>
Choose a workflow: 1) Write tests for existing code - analyzes implementation and generates comprehensive test suite, 2) Add test cases to existing tests - extends test files with additional scenarios, 3) Generate mocks for interfaces - creates testify/mock implementations, 4) Get testing guidance - explains Kratos testing patterns. Skill analyzes existing test patterns in your service, matches the style, and generates table-driven tests with helpers and mocks.
</quick_start>

<success_criteria>
Testing is complete when:
- Test files generated with proper structure
- All test helpers created (setup, cleanup, valid objects, mocks)
- Table-driven tests cover all scenarios (happy path, validation, edge cases, errors)
- Mock expectations properly configured
- `make test` runs successfully
- Test coverage is comprehensive
</success_criteria>

<essential_principles>
<overview>
This skill helps write tests for go-kratos microservices following Clean Architecture patterns observed in symbol-service and other workspace services.
</overview>

<layer_specific_testing>
<data_layer location="internal/data/repo/">
Uses in-memory SQLite for fast, isolated tests. Tests GORM entity transformations (toEntity/toDomain), database operations (Create, Update, FindByID, List, Delete), and error mapping (GORM errors to business errors).
</data_layer>

<biz_layer location="internal/biz/">
Uses testify/mock for repository mocks. Tests business logic and validation, error handling and edge cases. Never touches database directly.
</biz_layer>

<service_layer location="internal/service/">
Uses testify/mock for use case mocks. Tests proto message mapping, request/response transformations, and gRPC/HTTP service handlers.
</service_layer>
</layer_specific_testing>

<table_driven_tests>
All tests use table-driven patterns with:
- Descriptive test names
- Setup functions for test data preparation
- Mock setup functions for dependency configuration
- Check functions for result validation
- Comprehensive scenarios (happy path, validation errors, edge cases, database errors)
</table_driven_tests>

<test_helpers>
Each test file includes helpers:
- `setupTestDB(t)` - In-memory SQLite for data layer
- `validDomainX()` - Valid domain objects for testing
- `validEntityX()` - Valid GORM entities for testing
- `setupXUseCase(mockRepo)` - Use case initialization with mocks
- `cleanupDB(db)` - Test data cleanup
</test_helpers>

<mock_interfaces>
Uses testify/mock for all interfaces:
- Repository interfaces mocked in biz layer tests
- UseCase interfaces mocked in service layer tests
- Compile-time interface checks: `var _ Interface = (*Mock)(nil)`
- Mock expectations verified: `mock.AssertExpectations(t)`
</mock_interfaces>

<validation>
After writing tests:
- Run `make test` from service directory
- Verify all tests pass
- Check test coverage with `make coverage`
- Fix any failures before proceeding
</validation>
</essential_principles>

<intake>
What would you like to do?

1. **Write tests for existing code** - When you have implementation without tests, analyzes code and generates comprehensive test suite
2. **Add test cases** - When you have existing tests but need additional scenarios or edge cases covered
3. **Generate mocks** - When you need testify/mock implementations for repository or use case interfaces
4. **Get testing guidance** - When you need to understand Kratos testing patterns and best practices

**Choose an option above.**
</intake>

<routing>
| Response | Workflow |
|----------|----------|
| 1, "write", "new", "create tests", "test for" | `workflows/write-tests.md` |
| 2, "add", "extend", "more cases", "additional" | `workflows/add-test-cases.md` |
| 3, "mock", "generate mock", "create mock" | `workflows/generate-mocks.md` |
| 4, "guidance", "guide", "help", "explain", "how" | `workflows/testing-guidance.md` |

**After reading the workflow, follow it exactly.**
</routing>

<reference_index>
<testing_pattern_references location="references/">
<layer_patterns>
data-layer-tests.md, biz-layer-tests.md, service-layer-tests.md
</layer_patterns>

<components>
test-helpers.md, mock-patterns.md, table-driven-tests.md
</components>

<validation>
error-testing.md, edge-cases.md
</validation>
</testing_pattern_references>
</reference_index>

<workflows_index>
| Workflow | Purpose |
|----------|---------|
| write-tests.md | Generate comprehensive test suite for existing code |
| add-test-cases.md | Extend existing test file with new scenarios |
| generate-mocks.md | Create testify/mock implementations |
| testing-guidance.md | Explain testing patterns and best practices |
</workflows_index>