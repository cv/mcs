# Git Hooks

Shareable git hooks for enforcing code quality.

## Installation

```bash
./.githooks/install.sh
```

This configures git to use the `.githooks/` directory for hooks.

## Hooks

### pre-commit
Runs on every commit:
- **gofmt** - Check code formatting
- **go vet** - Static analysis
- **golangci-lint** - Comprehensive linting (if installed)
- **go test -short** - Quick tests
- **bd sync** - Flush beads changes (if applicable)

### pre-push
Runs before pushing:
- **go test -race** - Full test suite with race detection
- **go build** - Ensure everything compiles
- **bd sync** - Sync beads to remote (if sync-branch configured)

### commit-msg
Validates commit messages:
- Minimum 10 characters
- Warns if first line exceeds 72 characters

## Bypass

To skip hooks temporarily:
```bash
git commit --no-verify
git push --no-verify
```

## Requirements

- Go 1.21+
- golangci-lint (optional but recommended): `brew install golangci-lint`
