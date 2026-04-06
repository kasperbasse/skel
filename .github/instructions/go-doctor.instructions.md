---
applyTo: "internal/app/doctor/**"
---

# internal/app/doctor - Review Instructions

## Rules-based design

Tool requirements are data, not code. Adding a new tool means adding a
`sectionToolRule` to the `rules` slice - no other code changes needed.

Each rule has:

- `Section` - the restore section key (matches `restore.Options.Sections` keys)
- `Tool` - the binary name checked via `exec.LookPath`
- `Needed` - predicate: is this tool needed given this profile's data?

## Intentional omissions (do not add without justification)

These tools are NOT in the rules and must stay out:

- `nvim` - restore only writes config files, never invokes nvim
- `git` - restore only writes `.gitconfig`, never invokes git
- `node` - npm existence check via the npm rule already covers this
- `pip3` - pip packages are not restored by the current implementation

Adding them would block sections unnecessarily for users who have the data
but not the tool.

## BlockedSectionTools vs RequiredToolsForSections

- `BlockedSectionTools` - returns missing tools grouped by section; used by the
  TUI to pre-deselect sections the user can't run yet.
- `RequiredToolsForSections` - flat list of required tools for a section filter;
  used by `checkToolRequirements` to run the doctor check display.

Both operate on the same `rules` slice. Do not duplicate the logic.

## Testing rules

Test rule logic by calling `requiredToolsFor` / `BlockedSectionTools` with crafted
profiles. Do not mock `CommandExists` in rule tests - test the rule predicates directly.