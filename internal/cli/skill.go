package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cv/mcs/internal/skill"
	"github.com/spf13/cobra"
)

// skillVersionFile is the name of the file that stores the mcs version used to install the skill.
const skillVersionFile = ".mcs-version"

// skillsDirEnv names the environment variable that overrides the skills directory.
const skillsDirEnv = "MCS_SKILLS_DIR"

// getSkillPath returns the path where the skill should be installed:
// $MCS_SKILLS_DIR/mcs-control if set, otherwise ~/.agents/skills/mcs-control.
func getSkillPath() (string, error) {
	if dir := os.Getenv(skillsDirEnv); dir != "" {
		return filepath.Join(dir, skill.SkillName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".agents", "skills", skill.SkillName), nil
}

// uninstallSkill removes the skill directory if it exists.
func uninstallSkill() error {
	skillPath, err := getSkillPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		// Already doesn't exist, nothing to do
		return nil
	}

	if err := os.RemoveAll(skillPath); err != nil {
		return fmt.Errorf("failed to remove skill directory: %w", err)
	}

	return nil
}

// NewSkillCmd creates the skill command.
func NewSkillCmd(cfg *CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   skillCommandName,
		Short: "Manage the mcs agent skill",
		Long: `Install, uninstall, or show the path of the mcs agent skill for natural language vehicle control.

The skill uses the Agent Skills format (SKILL.md) and is installed to
~/.agents/skills/mcs-control/ by default. Set MCS_SKILLS_DIR to install it
into another agent's skills directory instead.`,
		Example: `  # Install the skill to ~/.agents/skills/mcs-control/
  mcs skill install

  # Install into a different skills directory
  MCS_SKILLS_DIR=/path/to/skills mcs skill install

  # Uninstall the skill
  mcs skill uninstall

  # Show where the skill would be installed
  mcs skill path`,
	}

	cmd.AddCommand(NewSkillInstallCmd(cfg))
	cmd.AddCommand(NewSkillUninstallCmd())
	cmd.AddCommand(NewSkillPathCmd())

	return cmd
}

// copySkillFile handles copying a single embedded skill file to the destination.
func copySkillFile(path string, d fs.DirEntry, skillPath string) error {
	// Skip the root "files" directory
	if path == "files" {
		return nil
	}

	// Get the relative path (strip "files/" prefix)
	relPath := path[len("files/"):]
	destPath := filepath.Join(skillPath, relPath)

	if d.IsDir() {
		return os.MkdirAll(destPath, 0755)
	}

	// Read embedded file
	content, err := skill.SkillFiles.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read embedded file %s: %w", path, err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// installSkillFiles copies all embedded skill files to the skill directory.
func installSkillFiles(skillPath string) error {
	return fs.WalkDir(skill.SkillFiles, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		return copySkillFile(path, d, skillPath)
	})
}

// NewSkillInstallCmd creates the skill install subcommand.
func NewSkillInstallCmd(cfg *CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the mcs agent skill",
		Long: `Install the mcs agent skill to ~/.agents/skills/mcs-control/
(or $MCS_SKILLS_DIR/mcs-control/ when MCS_SKILLS_DIR is set).

This enables natural language control of your vehicle through any coding agent
that supports Agent Skills. For example, you can say "warm up the car" and the
agent will run "mcs climate on".`,
		Example: `  # Install the skill
  mcs skill install

  # After installation, you can ask your agent:
  # "warm up the car" -> mcs climate on
  # "lock the doors" -> mcs lock
  # "what's my battery level?" -> mcs status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillPath, err := getSkillPath()
			if err != nil {
				return err
			}

			// Remove old version first.
			if err := uninstallSkill(); err != nil {
				return fmt.Errorf("failed to remove old skill: %w", err)
			}

			// Create skill directory.
			if err := os.MkdirAll(skillPath, 0755); err != nil {
				return fmt.Errorf("failed to create skill directory: %w", err)
			}

			// Copy embedded files to skill directory.
			if err := installSkillFiles(skillPath); err != nil {
				return fmt.Errorf("failed to install skill files: %w", err)
			}

			// Write version file to track which mcs version installed the skill.
			versionPath := filepath.Join(skillPath, skillVersionFile)
			if err := os.WriteFile(versionPath, []byte(cfg.Version), 0644); err != nil {
				return fmt.Errorf("failed to write version file: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Skill installed to %s\n", skillPath)

			return nil
		},
		SilenceUsage: true,
	}

	return cmd
}

// SkillVersionStatus represents the result of checking skill version compatibility.
type SkillVersionStatus int

const (
	// SkillNotInstalled means the skill directory does not exist.
	SkillNotInstalled SkillVersionStatus = iota
	// SkillVersionMatch means the installed skill version matches the current mcs version.
	SkillVersionMatch
	// SkillVersionMismatch means the installed skill version differs from the current mcs version.
	SkillVersionMismatch
	// SkillVersionUnknown means the skill is installed but has no version file (legacy install).
	SkillVersionUnknown
)

// CheckSkillVersion checks if the installed skill version matches the given version.
// Returns the status and the installed version (empty string if not available).
func CheckSkillVersion(currentVersion string) (SkillVersionStatus, string) {
	skillPath, err := getSkillPath()
	if err != nil {
		return SkillNotInstalled, ""
	}

	// Check if skill directory exists.
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return SkillNotInstalled, ""
	}

	// Read version file.
	versionPath := filepath.Join(skillPath, skillVersionFile)
	content, err := os.ReadFile(versionPath)
	if err != nil {
		// Skill exists but no version file - legacy install.
		return SkillVersionUnknown, ""
	}

	installedVersion := strings.TrimSpace(string(content))
	if installedVersion == currentVersion {
		return SkillVersionMatch, installedVersion
	}

	return SkillVersionMismatch, installedVersion
}

// NewSkillUninstallCmd creates the skill uninstall subcommand.
func NewSkillUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall the mcs agent skill",
		Long:    `Remove the mcs agent skill from ~/.agents/skills/mcs-control/ (or $MCS_SKILLS_DIR/mcs-control/).`,
		Example: `  mcs skill uninstall`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillPath, err := getSkillPath()
			if err != nil {
				return err
			}

			// Check if skill exists
			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Skill not installed at %s\n", skillPath)

				return nil
			}

			if err := uninstallSkill(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Skill uninstalled from %s\n", skillPath)

			return nil
		},
		SilenceUsage: true,
	}

	return cmd
}

// NewSkillPathCmd creates the skill path subcommand.
func NewSkillPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path",
		Short:   "Show skill installation path",
		Long:    `Print the path where the mcs agent skill is installed.`,
		Example: `  mcs skill path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillPath, err := getSkillPath()
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), skillPath)

			return nil
		},
		SilenceUsage: true,
	}

	return cmd
}
