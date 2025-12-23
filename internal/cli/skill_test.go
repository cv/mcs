package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cv/mcs/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSkillPath tests that getSkillPath returns the correct path
func TestGetSkillPath(t *testing.T) {
	path, err := getSkillPath()
	require.NoError(t, err, "Expected getSkillPath to succeed, got error: %v")

	home, err := os.UserHomeDir()
	require.NoError(t, err, "Expected UserHomeDir to succeed, got error: %v")

	expected := filepath.Join(home, ".claude", "skills", skill.SkillName)
	assert.Equal(t, expected, path)

	// Verify it ends with the skill name
	assert.Equal(t, skill.SkillName, filepath.Base(path))
}

// TestUninstallSkill tests the uninstallSkill function
func TestUninstallSkill(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string)
		wantErr   bool
	}{
		{
			name: "removes existing directory",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				// Create the directory structure
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create test directory: %v")

				// Add a file to verify recursive removal
				testFile := filepath.Join(skillPath, "test.txt")
				err = os.WriteFile(testFile, []byte("test"), 0644)
				require.NoError(t, err, "Failed to create test file: %v")
			},
			wantErr: false,
		},
		{
			name: "handles non-existent directory gracefully",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				// Don't create anything - directory doesn't exist
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir := t.TempDir()
			tt.setupFunc(t, tempDir)

			// Override the home directory (t.Setenv auto-restores)
			t.Setenv("HOME", tempDir)

			// Run uninstall
			err := uninstallSkill()

			if tt.wantErr {
				require.Error(t, err, "uninstallSkill() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "uninstallSkill() error = %v, wantErr %v")
			}

			// Verify directory is removed
			skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
			assert.NoDirExists(t, skillPath)
		})
	}
}

// TestSkillCommand tests the skill parent command
func TestSkillCommand(t *testing.T) {
	cmd := NewSkillCmd()
	assertCommandBasics(t, cmd, "skill")
}

// TestSkillCommand_Subcommands tests that skill subcommands exist
func TestSkillCommand_Subcommands(t *testing.T) {
	cmd := NewSkillCmd()
	assertSubcommandsExist(t, cmd, []string{"install", "uninstall", "path"})
}

// TestSkillInstallCommand tests the skill install subcommand
func TestSkillInstallCommand(t *testing.T) {
	cmd := NewSkillInstallCmd()
	assertCommandBasics(t, cmd, "install")
	assertNoArgsCommand(t, cmd)
}

// TestSkillInstallCommand_Execute tests executing the install command
func TestSkillInstallCommand_Execute(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Execute install command
	cmd := NewSkillInstallCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err := cmd.Execute()
	require.NoError(t, err, "Expected install to succeed, got error: %v")

	// Verify output message
	output := outBuf.String()
	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
	expectedOutput := "Skill installed to " + skillPath + "\n"
	assert.Equal(t, expectedOutput, output)

	// Verify directory was created
	assert.DirExists(t, skillPath)

	// Verify files were copied
	expectedFiles := []string{
		"SKILL.md",
		"COMMAND_REFERENCE.md",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(skillPath, file)
		assert.FileExists(t, filePath, "Expected file %s to exist", file)

		// Verify file has content
		content, err := os.ReadFile(filePath)
		require.NoErrorf(t, err, "Expected to read file %s, got error: %v", file, err)
		assert.NotEmptyf(t, content, "Expected file %s to have content", file)
	}
}

// TestSkillInstallCommand_ReinstallRemovesOld tests that reinstalling removes old files
func TestSkillInstallCommand_ReinstallRemovesOld(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)

	// Create old installation with extra file
	err := os.MkdirAll(skillPath, 0755)
	require.NoError(t, err, "Failed to create skill directory: %v")

	oldFilePath := filepath.Join(skillPath, "old_file.txt")
	err = os.WriteFile(oldFilePath, []byte("old content"), 0644)
	require.NoError(t, err, "Failed to create old file: %v")

	// Execute install command
	cmd := NewSkillInstallCmd()
	cmd.SetOut(&bytes.Buffer{})

	err = cmd.Execute()
	require.NoError(t, err, "Expected install to succeed, got error: %v")

	// Verify old file is gone
	assert.NoFileExists(t, oldFilePath)

	// Verify new files exist
	newFilePath := filepath.Join(skillPath, "SKILL.md")
	assert.FileExists(t, newFilePath)
}

// TestSkillUninstallCommand tests the skill uninstall subcommand
func TestSkillUninstallCommand(t *testing.T) {
	cmd := NewSkillUninstallCmd()
	assertCommandBasics(t, cmd, "uninstall")
	assertNoArgsCommand(t, cmd)
}

// TestSkillUninstallCommand_Execute tests executing the uninstall command
func TestSkillUninstallCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tempDir string)
		expectedOutput string
	}{
		{
			name: "removes existing installation",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create skill directory: %v")

				testFile := filepath.Join(skillPath, "test.txt")
				err = os.WriteFile(testFile, []byte("test"), 0644)
				require.NoError(t, err, "Failed to create test file: %v")
			},
			expectedOutput: "Skill uninstalled from ",
		},
		{
			name: "handles non-existent installation",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				// Don't create anything
			},
			expectedOutput: "Skill not installed at ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir := t.TempDir()
			tt.setupFunc(t, tempDir)
			t.Setenv("HOME", tempDir)

			// Execute uninstall command
			cmd := NewSkillUninstallCmd()
			var outBuf bytes.Buffer
			cmd.SetOut(&outBuf)

			err := cmd.Execute()
			require.NoError(t, err, "Expected uninstall to succeed, got error: %v")

			// Verify output message
			output := outBuf.String()
			assert.NotEmpty(t, output, "Expected output message")
			if len(tt.expectedOutput) > 0 {
				assert.Contains(t, output, tt.expectedOutput)
			}

			// Verify directory is removed
			skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
			assert.NoDirExists(t, skillPath)
		})
	}
}

// TestSkillPathCommand tests the skill path subcommand
func TestSkillPathCommand(t *testing.T) {
	cmd := NewSkillPathCmd()
	assertCommandBasics(t, cmd, "path")
	assertNoArgsCommand(t, cmd)
}

// TestSkillPathCommand_Execute tests executing the path command
func TestSkillPathCommand_Execute(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Execute path command
	cmd := NewSkillPathCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err := cmd.Execute()
	require.NoError(t, err, "Expected path command to succeed, got error: %v")

	// Verify output is the correct path
	output := outBuf.String()
	expectedPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName) + "\n"
	assert.Equal(t, expectedPath, output)
}

// TestSkillPathCommand_OutputFormat tests that path output is just the path
func TestSkillPathCommand_OutputFormat(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Execute path command
	cmd := NewSkillPathCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err := cmd.Execute()
	require.NoError(t, err, "Expected path command to succeed, got error: %v")

	// Verify output is a single line
	output := outBuf.String()
	lines := bytes.Split([]byte(output), []byte("\n"))
	// Should have exactly 2 elements: the path and an empty string after final newline
	assert.Len(t, lines, 2)
	assert.Empty(t, lines[1])

	// Verify it's a valid path
	pathStr := string(lines[0])
	assert.Truef(t, filepath.IsAbs(pathStr), "Expected absolute path, got '%s'", pathStr)
}

// TestSkillInstallCommand_WritesVersionFile tests that install creates a version file
func TestSkillInstallCommand_WritesVersionFile(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Execute install command
	cmd := NewSkillInstallCmd()
	cmd.SetOut(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err, "Expected install to succeed, got error: %v")

	// Verify version file was created
	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
	versionPath := filepath.Join(skillPath, ".mcs-version")

	content, err := os.ReadFile(versionPath)
	require.NoError(t, err, "Expected version file to exist, got error: %v")

	// Verify content matches current version
	assert.Equal(t, Version, string(content))
}

// TestCheckSkillVersion tests the CheckSkillVersion function
func TestCheckSkillVersion(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(t *testing.T, tempDir string)
		expectedStatus  SkillVersionStatus
		expectedVersion string
	}{
		{
			name: "returns SkillNotInstalled when skill directory does not exist",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				// Don't create anything
			},
			expectedStatus:  SkillNotInstalled,
			expectedVersion: "",
		},
		{
			name: "returns SkillVersionUnknown when skill exists without version file",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create skill directory: %v")

				// Create SKILL.md but no version file (legacy install)
				err = os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("test"), 0644)
				require.NoError(t, err, "Failed to create SKILL.md: %v")

			},
			expectedStatus:  SkillVersionUnknown,
			expectedVersion: "",
		},
		{
			name: "returns SkillVersionMatch when versions match",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create skill directory: %v")

				versionPath := filepath.Join(skillPath, ".mcs-version")
				err = os.WriteFile(versionPath, []byte(Version), 0644)
				require.NoError(t, err, "Failed to create version file: %v")
			},
			expectedStatus:  SkillVersionMatch,
			expectedVersion: Version,
		},
		{
			name: "returns SkillVersionMismatch when versions differ",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create skill directory: %v")

				versionPath := filepath.Join(skillPath, ".mcs-version")
				err = os.WriteFile(versionPath, []byte("1.0.0"), 0644)
				require.NoError(t, err, "Failed to create version file: %v")
			},
			expectedStatus:  SkillVersionMismatch,
			expectedVersion: "1.0.0",
		},
		{
			name: "handles version file with whitespace",
			setupFunc: func(t *testing.T, tempDir string) {
				t.Helper()
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				err := os.MkdirAll(skillPath, 0755)
				require.NoError(t, err, "Failed to create skill directory: %v")

				versionPath := filepath.Join(skillPath, ".mcs-version")
				// Write version with trailing newline
				err = os.WriteFile(versionPath, []byte(Version+"\n"), 0644)
				require.NoError(t, err, "Failed to create version file: %v")
			},
			expectedStatus:  SkillVersionMatch,
			expectedVersion: Version,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir := t.TempDir()
			tt.setupFunc(t, tempDir)
			t.Setenv("HOME", tempDir)

			status, version := CheckSkillVersion()

			assert.Equal(t, tt.expectedStatus, status)

			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}
