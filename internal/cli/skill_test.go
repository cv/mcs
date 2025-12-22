package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/cv/mcs/internal/skill"
)

// TestGetSkillPath tests that getSkillPath returns the correct path
func TestGetSkillPath(t *testing.T) {
	path, err := getSkillPath()
	if err != nil {
		t.Fatalf("Expected getSkillPath to succeed, got error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Expected UserHomeDir to succeed, got error: %v", err)
	}

	expected := filepath.Join(home, ".claude", "skills", skill.SkillName)
	if path != expected {
		t.Errorf("Expected path to be '%s', got '%s'", expected, path)
	}

	// Verify it ends with the skill name
	if filepath.Base(path) != skill.SkillName {
		t.Errorf("Expected path to end with '%s', got '%s'", skill.SkillName, filepath.Base(path))
	}
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
				// Create the directory structure
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				// Add a file to verify recursive removal
				testFile := filepath.Join(skillPath, "test.txt")
				if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "handles non-existent directory gracefully",
			setupFunc: func(t *testing.T, tempDir string) {
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

			// Temporarily override the home directory
			originalHome := os.Getenv("HOME")
			t.Cleanup(func() {
				if originalHome != "" {
					_ = os.Setenv("HOME", originalHome)
				} else {
					_ = os.Unsetenv("HOME")
				}
			})
			_ = os.Setenv("HOME", tempDir)

			// Run uninstall
			err := uninstallSkill()

			if (err != nil) != tt.wantErr {
				t.Errorf("uninstallSkill() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify directory is removed
			skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
			if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
				t.Errorf("Expected skill directory to be removed, but it still exists")
			}
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
	assertSubcommandsExist(t, cmd, []string{"install", "uninstall", "path"}, true)
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

	// Temporarily override the home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})
	_ = os.Setenv("HOME", tempDir)

	// Execute install command
	cmd := NewSkillInstallCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected install to succeed, got error: %v", err)
	}

	// Verify output message
	output := outBuf.String()
	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
	expectedOutput := "Skill installed to " + skillPath + "\n"
	if output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}

	// Verify directory was created
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Error("Expected skill directory to be created")
	}

	// Verify files were copied
	expectedFiles := []string{
		"SKILL.md",
		"COMMAND_REFERENCE.md",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(skillPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}

		// Verify file has content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Expected to read file %s, got error: %v", file, err)
		}
		if len(content) == 0 {
			t.Errorf("Expected file %s to have content", file)
		}
	}
}

// TestSkillInstallCommand_ReinstallRemovesOld tests that reinstalling removes old files
func TestSkillInstallCommand_ReinstallRemovesOld(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Temporarily override the home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})
	_ = os.Setenv("HOME", tempDir)

	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)

	// Create old installation with extra file
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}
	oldFilePath := filepath.Join(skillPath, "old_file.txt")
	if err := os.WriteFile(oldFilePath, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	// Execute install command
	cmd := NewSkillInstallCmd()
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected install to succeed, got error: %v", err)
	}

	// Verify old file is gone
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Error("Expected old file to be removed during reinstall")
	}

	// Verify new files exist
	newFilePath := filepath.Join(skillPath, "SKILL.md")
	if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
		t.Error("Expected new files to be installed")
	}
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
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}
				testFile := filepath.Join(skillPath, "test.txt")
				if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
			expectedOutput: "Skill uninstalled from ",
		},
		{
			name: "handles non-existent installation",
			setupFunc: func(t *testing.T, tempDir string) {
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

			// Temporarily override the home directory
			originalHome := os.Getenv("HOME")
			t.Cleanup(func() {
				if originalHome != "" {
					_ = os.Setenv("HOME", originalHome)
				} else {
					_ = os.Unsetenv("HOME")
				}
			})
			_ = os.Setenv("HOME", tempDir)

			// Execute uninstall command
			cmd := NewSkillUninstallCmd()
			var outBuf bytes.Buffer
			cmd.SetOut(&outBuf)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Expected uninstall to succeed, got error: %v", err)
			}

			// Verify output message
			output := outBuf.String()
			if len(output) == 0 {
				t.Error("Expected output message")
			}
			if len(tt.expectedOutput) > 0 && !bytes.Contains([]byte(output), []byte(tt.expectedOutput)) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, output)
			}

			// Verify directory is removed
			skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
			if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
				t.Error("Expected skill directory to be removed")
			}
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

	// Temporarily override the home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})
	_ = os.Setenv("HOME", tempDir)

	// Execute path command
	cmd := NewSkillPathCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected path command to succeed, got error: %v", err)
	}

	// Verify output is the correct path
	output := outBuf.String()
	expectedPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName) + "\n"
	if output != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, output)
	}
}

// TestSkillPathCommand_OutputFormat tests that path output is just the path
func TestSkillPathCommand_OutputFormat(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Temporarily override the home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})
	_ = os.Setenv("HOME", tempDir)

	// Execute path command
	cmd := NewSkillPathCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected path command to succeed, got error: %v", err)
	}

	// Verify output is a single line
	output := outBuf.String()
	lines := bytes.Split([]byte(output), []byte("\n"))
	// Should have exactly 2 elements: the path and an empty string after final newline
	if len(lines) != 2 || len(lines[1]) != 0 {
		t.Errorf("Expected single line output, got %d lines", len(lines)-1)
	}

	// Verify it's a valid path
	pathStr := string(lines[0])
	if !filepath.IsAbs(pathStr) {
		t.Errorf("Expected absolute path, got '%s'", pathStr)
	}
}

// TestSkillInstallCommand_WritesVersionFile tests that install creates a version file
func TestSkillInstallCommand_WritesVersionFile(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Temporarily override the home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
	})
	_ = os.Setenv("HOME", tempDir)

	// Execute install command
	cmd := NewSkillInstallCmd()
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected install to succeed, got error: %v", err)
	}

	// Verify version file was created
	skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
	versionPath := filepath.Join(skillPath, ".mcs-version")

	content, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("Expected version file to exist, got error: %v", err)
	}

	// Verify content matches current version
	if string(content) != Version {
		t.Errorf("Expected version file to contain '%s', got '%s'", Version, string(content))
	}
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
				// Don't create anything
			},
			expectedStatus:  SkillNotInstalled,
			expectedVersion: "",
		},
		{
			name: "returns SkillVersionUnknown when skill exists without version file",
			setupFunc: func(t *testing.T, tempDir string) {
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}
				// Create SKILL.md but no version file (legacy install)
				if err := os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create SKILL.md: %v", err)
				}
			},
			expectedStatus:  SkillVersionUnknown,
			expectedVersion: "",
		},
		{
			name: "returns SkillVersionMatch when versions match",
			setupFunc: func(t *testing.T, tempDir string) {
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}
				versionPath := filepath.Join(skillPath, ".mcs-version")
				if err := os.WriteFile(versionPath, []byte(Version), 0644); err != nil {
					t.Fatalf("Failed to create version file: %v", err)
				}
			},
			expectedStatus:  SkillVersionMatch,
			expectedVersion: Version,
		},
		{
			name: "returns SkillVersionMismatch when versions differ",
			setupFunc: func(t *testing.T, tempDir string) {
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}
				versionPath := filepath.Join(skillPath, ".mcs-version")
				if err := os.WriteFile(versionPath, []byte("1.0.0"), 0644); err != nil {
					t.Fatalf("Failed to create version file: %v", err)
				}
			},
			expectedStatus:  SkillVersionMismatch,
			expectedVersion: "1.0.0",
		},
		{
			name: "handles version file with whitespace",
			setupFunc: func(t *testing.T, tempDir string) {
				skillPath := filepath.Join(tempDir, ".claude", "skills", skill.SkillName)
				if err := os.MkdirAll(skillPath, 0755); err != nil {
					t.Fatalf("Failed to create skill directory: %v", err)
				}
				versionPath := filepath.Join(skillPath, ".mcs-version")
				// Write version with trailing newline
				if err := os.WriteFile(versionPath, []byte(Version+"\n"), 0644); err != nil {
					t.Fatalf("Failed to create version file: %v", err)
				}
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

			// Temporarily override the home directory
			originalHome := os.Getenv("HOME")
			t.Cleanup(func() {
				if originalHome != "" {
					_ = os.Setenv("HOME", originalHome)
				} else {
					_ = os.Unsetenv("HOME")
				}
			})
			_ = os.Setenv("HOME", tempDir)

			status, version := CheckSkillVersion()

			if status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, status)
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version '%s', got '%s'", tt.expectedVersion, version)
			}
		})
	}
}
