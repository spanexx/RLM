package rlmconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrecedence_WorkspaceOverGlobal(t *testing.T) {
	ws := t.TempDir()
	globalDir := t.TempDir()
	workspaceDir := t.TempDir()

	// Force global config path via config dir env.
	cfgRoot := t.TempDir()
	old := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", cfgRoot)
	t.Cleanup(func() {
		_ = os.Setenv("XDG_CONFIG_HOME", old)
	})

	if err := WriteConfig(GlobalConfigPath(), Config{ContextDir: globalDir}); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfig(WorkspaceConfigPath(ws), Config{ContextDir: workspaceDir}); err != nil {
		t.Fatal(err)
	}

	resolved, err := Resolve(ResolveOptions{WorkspaceRoot: ws})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ContextDir != workspaceDir {
		t.Fatalf("expected workspace context_dir %q, got %q", workspaceDir, resolved.ContextDir)
	}
	if resolved.Source != "workspace" {
		t.Fatalf("expected source workspace, got %q", resolved.Source)
	}
}

func TestWorkspaceConfigPath(t *testing.T) {
	ws := "/tmp/example"
	p := WorkspaceConfigPath(ws)
	if filepath.Base(p) != "config.json" {
		t.Fatalf("unexpected path: %s", p)
	}
}
