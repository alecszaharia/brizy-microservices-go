---
name: sync-skills
description: Updates project skills to reflect current codebase patterns and best practices
args:
  - name: skill_name
    description: Name of the skill to update (optional - if not provided, analyzes all skills)
    required: false
---

You are updating project skills to align with the latest codebase patterns and conventions.

## Context

This project has dedicated skills that encode domain knowledge about how to develop features:
- `creating-gorm-entity` - GORM entity generation patterns
- `kratos-repo` - Repository implementation patterns
- `kratos-tests` - Testing patterns for Kratos microservices
- `path-finder` - File location conventions

As the codebase evolves, these skills may become outdated. This command analyzes current code and updates skill documentation.

## Process

<step name="determine_scope">
If `{{skill_name}}` is provided:
- Focus on updating that specific skill
- Example: `{{skill_name}}` = "creating-gorm-entity"

If no skill name provided:
- Ask user which skill(s) to update
- Show list of available skills
- Allow "all" option to update everything
</step>

<step name="analyze_current_patterns">
For each skill being updated:

1. **Read the current skill documentation**
   - Location: `.claude/skills/{{skill_name}}/SKILL.md`
   - Note the patterns it documents
   - Identify what it claims as "best practices"

2. **Analyze current codebase**
   - Find recent examples of the pattern in action
   - For GORM entities: check `internal/data/model/*.go`
   - For repositories: check `internal/data/repo/*.go`
   - For tests: check `*_test.go` files
   - Look at git history to see recent changes

3. **Identify discrepancies**
   - Patterns in skill that aren't used in code
   - Patterns in code that aren't documented in skill
   - Outdated examples or templates
   - New best practices not yet captured
</step>

<step name="gather_evidence">
For each discrepancy found:

1. **Collect code examples**
   - Find 2-3 recent examples from the codebase
   - Include file paths and line numbers
   - Note when they were last modified (git log)

2. **Document the pattern**
   - What's the current approach?
   - Why is it better than what's documented?
   - Are there any edge cases?
   - Is this consistent across the codebase?
</step>

<step name="update_skill">
Update the skill documentation:

1. **Update essential patterns section**
   - Revise code examples to match current usage
   - Add new patterns discovered in codebase
   - Remove deprecated patterns
   - Update field ordering, naming conventions, etc.

2. **Update templates and references**
   - Check `references/` subdirectory if it exists
   - Update complete templates
   - Revise code snippets

3. **Update validation checklist**
   - Add checks for new patterns
   - Remove checks for deprecated patterns
   - Ensure checklist matches current best practices

4. **Update success criteria**
   - Align with current code quality standards
   - Add new requirements if discovered
</step>

<step name="validate_changes">
After updating the skill:

1. **Test the updated skill**
   - Run the skill in a test scenario
   - Verify it generates code matching current patterns
   - Check if generated code would pass review

2. **Document what changed**
   - Summarize updates made
   - List removed deprecated patterns
   - List new patterns added
   - Note any breaking changes

3. **Show diff to user**
   - Present key changes made
   - Explain rationale for each change
   - Ask for confirmation before saving
</step>

## Workflow

When user runs `/sync-skills` or `/sync-skills creating-gorm-entity`:

1. **Announce intent**
   ```
   Analyzing codebase to update {{skill_name}} skill...
   ```

2. **Show findings**
   ```
   Found discrepancies:
   - Pattern X: skill says Y, but code uses Z (example: path/to/file.go:123)
   - New pattern A: not documented in skill, used in 5 recent files
   - Deprecated pattern B: still in skill docs, removed from codebase 3 commits ago
   ```

3. **Present proposed changes**
   ```
   Proposed updates to {{skill_name}}:

   1. Update <essential_patterns> section:
      - Change field ordering rule from X to Z
      - Add new pattern for A
      - Remove deprecated pattern B

   2. Update validation checklist:
      - Add check for A
      - Remove check for B
   ```

4. **Apply changes with user confirmation**
   - Show before/after for critical sections
   - Update SKILL.md file
   - Update any referenced files in references/ directory

5. **Verify updates**
   - Confirm skill documentation is consistent
   - Run basic validation (if applicable)
   - Report completion

## Output Format

Present findings as:

```
## Skill Update Report: {{skill_name}}

### Current State Analysis
- Analyzed {{N}} code files
- Found {{M}} instances of this pattern
- Last modified: {{date}} ({{commit}})

### Discrepancies Found
1. **Pattern name**: Description
   - Skill says: ...
   - Codebase uses: ...
   - Evidence: file.go:123, file2.go:456
   - Recommendation: Update skill to match codebase

### Proposed Changes
1. **Section: essential_patterns**
   - Update example code for X
   - Add new subsection for Y
   - Remove deprecated pattern Z

2. **Section: validation_checklist**
   - Add: [ ] Check for Y
   - Remove: [ ] Check for Z

### Files to Update
- .claude/skills/{{skill_name}}/SKILL.md
- .claude/skills/{{skill_name}}/references/example.md (if exists)
```

## Important Notes

- **Preserve skill structure**: Don't change the YAML front matter or core XML structure
- **Keep it concise**: Update patterns, don't rewrite entire skill
- **Evidence-based**: Every change should reference actual code in the codebase
- **Backward compatible**: Note if changes break existing usage patterns
- **Test after update**: Verify the skill still works correctly after changes

## Example Usage

```bash
# Update specific skill
/sync-skills creating-gorm-entity

# Update all skills (will prompt for confirmation)
/sync-skills

# After git commits that change patterns
git commit -m "refactor: improve GORM entity structure"
/sync-skills creating-gorm-entity
```

## Common Update Scenarios

1. **New field added to entities**: Update field ordering rules, add to validation checklist
2. **Naming convention changed**: Update naming section, update all examples
3. **New GORM tag pattern**: Add to essential patterns with explanation
4. **Deprecated approach**: Remove from skill, add note about migration if needed
5. **New testing pattern**: Update test helpers, add to table-driven test examples