## Description

### Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes, code improvement)

### What does this PR do?

<!-- Provide a clear description of what this PR does and why. -->

### How was this change tested?

<!-- 
Describe how you tested this change:
- Unit tests
- Acceptance tests (run locally with Podman)
- Manual testing scenarios
-->

### Related Issues

<!-- Link to related issues using: Fixes #123, Related to #456 -->

## Testing Checklist

- [ ] Unit tests added/updated
- [ ] Acceptance tests added/updated (for new resources)
- [ ] `go fmt ./...` passes
- [ ] `go vet ./...` passes
- [ ] Tests run locally with Podman before pushing

## Documentation Checklist

- [ ] Documentation added to `docs/` directory
- [ ] Example added to `examples/` directory
- [ ] Schema attributes have descriptions
- [ ] Import section documented (if applicable)

## Checklist for Specific Changes

### New Resources
- [ ] API client methods in `pkg/rustfs/<name>.go`
- [ ] Resource file in `provider/rustfs_<name>_resource.go`
- [ ] Registered in `provider/provider.go`
- [ ] ImportState implemented
- [ ] Documentation + example

### New Data Sources
- [ ] Data source file in `provider/rustfs_<name>_datasource.go`
- [ ] Registered in `provider/provider.go`
- [ ] Documentation + example

### Bug Fixes
- [ ] Root cause identified
- [ ] Regression test added
- [ ] Edge cases considered
