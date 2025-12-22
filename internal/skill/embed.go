// Package skill provides embedded Claude Code skill files for installation.
package skill

import "embed"

// SkillFiles contains the embedded skill files (SKILL.md, COMMAND_REFERENCE.md).
// These are installed to ~/.claude/skills/mcs-control/ via `mcs skill install`.
//
//go:embed files/*
var SkillFiles embed.FS

// SkillName is the directory name used when installing the skill.
const SkillName = "mcs-control"
