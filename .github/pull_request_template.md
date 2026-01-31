## Description

Brief description of changes.

## Type of Change

- [ ] Bug fix (patch bump)
- [ ] New feature (minor bump)
- [ ] Breaking change (major bump)
- [ ] Documentation update
- [ ] CI/tooling change

## Consignment

**Required for code changes:**

- [ ] Consignment created via `shipyard add --summary "..." --bump [major|minor|patch]`
- [ ] Or `skip-consignment` label applied (if no versioning needed)

## Testing

How was this tested?

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist

- [ ] Consignment added (if code changes)
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests pass locally (`just ci`)
- [ ] No new warnings

---

**Note**: The `require-consignment` workflow will automatically check for consignments on PRs that modify code. Documentation-only changes are exempt.
