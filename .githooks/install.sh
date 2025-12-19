#!/bin/sh
# Install git hooks for mcs development
# Usage: ./.githooks/install.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Installing git hooks..."

# Configure git to use .githooks directory
git config core.hooksPath .githooks

# Make hooks executable
chmod +x "$SCRIPT_DIR/pre-commit"
chmod +x "$SCRIPT_DIR/pre-push"
chmod +x "$SCRIPT_DIR/commit-msg"

echo "Git hooks installed successfully!"
echo ""
echo "Hooks enabled:"
echo "  pre-commit  - gofmt, go vet, golangci-lint, tests"
echo "  pre-push    - full test suite with race detection, build"
echo "  commit-msg  - validates commit message format"
echo ""
echo "To disable hooks temporarily, use: git commit --no-verify"
