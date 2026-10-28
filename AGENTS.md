# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Modern Go

Follow [JetBrains Modern Go Guidelines](https://github.com/JetBrains/go-modern-guidelines) via the `use-modern-go` skill (`.agents/skills/use-modern-go`, mirrored in `.claude/skills/`). Before editing Go code, run:

```bash
sh .agents/skills/use-modern-go/scripts/run-tool.sh list --file-path <file.go>
sh .agents/skills/use-modern-go/scripts/run-tool.sh explain <guideline_id>
```

Read the full `list` output; it is scoped to the `go` version in `go.mod`. Update with `npx skills update use-modern-go -p -y` (tracked in `skills-lock.json`).

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

