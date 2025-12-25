---
name: creating-gorm-entity
description: Generates GORM entity structs with proper field tags, relationships, and table naming for go-kratos microservices. Use when adding new database models to symbol-service.
---

<objective>
Generate GORM entity structs that follow the established patterns in symbol-service, including proper field tags, relationship definitions, soft deletes, and table naming conventions.
</objective>

<quick_start>
Generate a basic GORM entity with standard fields:

Example request: "Create a User entity with ID, Email (unique, required), Name (required), and timestamps with soft deletes"

Expected output: Complete .go file in internal/data/model/ with proper package, imports, struct with GORM tags, JSON tags, and TableName() method.
</quick_start>

<essential_patterns>
All GORM entities in symbol-service follow these core patterns:

<package_structure>
```go
package model

import (
	"time"
	"gorm.io/gorm"  // only if using soft deletes
)
```
</package_structure>

<id_field>
**Primary Key**: Always `uint64` type (not `uint` or `uint32`)

```go
ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
```
</id_field>

<required_fields>
**Strings**: Always specify size with not null constraint
**Numbers**: Just not null

```go
Name      string `gorm:"not null;size:255" json:"name"`
ProjectID uint64 `gorm:"not null" json:"project_id"`
```
</required_fields>

<unique_constraints>
**Single**: `gorm:"uniqueIndex"`
**Composite**: Use same index name with different priorities

```go
ProjectID uint64 `gorm:"not null;uniqueIndex:idx_project_uid,priority:1" json:"project_id"`
UID       string `gorm:"not null;size:255;uniqueIndex:idx_project_uid,priority:2" json:"uid"`
```
</unique_constraints>

<timestamps>
Required on all entities:

```go
CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
```
</timestamps>

<soft_deletes>
Optional soft delete support:

```go
DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
```
</soft_deletes>

<table_name_method>
Required on all entities:

```go
func (EntityName) TableName() string {
	return "table_name"  // snake_case plural
}
```
</table_name_method>

<field_ordering>
Follow this order in struct definitions:
1. ID (primary key)
2. Foreign key IDs
3. Business fields (UID, names, descriptive fields)
4. Enum/status fields
5. Numeric fields (version, counts)
6. Relationship fields (pointers to other entities)
7. Timestamps (CreatedAt, UpdatedAt)
8. Soft delete (DeletedAt, if applicable)
</field_ordering>
</essential_patterns>

<references>
For detailed guidance on specific topics, see:

- **references/field-types.md**: Detailed field type selection (numeric, string, binary, relationships, time)
- **references/indexing.md**: Index patterns (composite, single, unique indexes)
- **references/relationships.md**: Relationship examples (one-to-one, cascade delete)
- **references/naming.md**: Naming conventions (structs, fields, tables, JSON tags)
- **references/complete-template.md**: Full entity template with all patterns
</references>

<process>
When user requests a GORM entity:

<gather_requirements>
- Entity name (will be PascalCase)
- Fields needed (name, type, constraints)
- Relationships to other entities
- Whether soft deletes are needed
- Any unique constraints or composite indexes
</gather_requirements>

<generate_structure>
- Start with package and imports
- Create struct with fields in proper order (see <field_ordering>)
- Apply appropriate GORM tags
- Apply JSON tags (snake_case, omitempty where needed)
- Add timestamps (always required)
- Add soft delete if requested
- Add TableName() method
</generate_structure>

<validate_patterns>
- All IDs are `uint64`
- All required strings have `size` specified
- Timestamps use correct types (`time.Time`)
- Soft delete uses `gorm.DeletedAt`
- Foreign keys have proper cascade constraints
- JSON tags are snake_case
- Table name is snake_case plural
</validate_patterns>

<output_location>
**File**: `internal/data/model/{entity_name}.go`
**Package**: `model`
**Note**: Multiple related entities can go in same file
</output_location>
</process>

<validation_checklist>
Before presenting the generated entity, verify:

- [ ] Package is `model`
- [ ] Imports include `time` and `gorm.io/gorm` (if soft delete)
- [ ] ID field is `uint64` with `primaryKey;autoIncrement`
- [ ] All foreign key IDs are `uint64`
- [ ] Required strings have `not null;size:N`
- [ ] Relationships use pointer types and proper foreign key syntax
- [ ] Cascade delete specified for relationships: `constraint:OnDelete:CASCADE`
- [ ] CreatedAt and UpdatedAt are `time.Time` with auto tags
- [ ] DeletedAt (if present) is `gorm.DeletedAt` with `index` tag
- [ ] JSON tags are snake_case
- [ ] Optional/pointer fields have `omitempty` in JSON tags
- [ ] TableName() method uses value receiver
- [ ] Table name is snake_case plural
- [ ] Fields are in logical order (ID, FKs, business, relationships, timestamps)
</validation_checklist>

<success_criteria>
Generated entity:
- Compiles without errors
- Follows all naming conventions from symbol-service
- Uses correct field types (uint64 for IDs, time.Time for timestamps)
- Has proper GORM tags for constraints, indexes, and relationships
- Has correct JSON tags (snake_case with appropriate omitempty)
- Includes required timestamps
- Has TableName() method
- Matches the code style of existing entities in internal/data/model/symbol.go
</success_criteria>
