package rlmconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ContextDir string `json:"context_dir"`
}

type Resolved struct {
	WorkspaceRoot       string `json:"workspace_root"`
	GlobalConfigPath    string `json:"global_config_path"`
	WorkspaceConfigPath string `json:"workspace_config_path"`
	ContextDir          string `json:"context_dir"`
	Source              string `json:"source"`
}

type ResolveOptions struct {
	WorkspaceRoot string
	DirFlag       string
}

func GlobalConfigPath() string {
	d, err := os.UserConfigDir()
	if err != nil {
		h, _ := os.UserHomeDir()
		if h == "" {
			return ""
		}
		d = filepath.Join(h, ".config")
	}
	return filepath.Join(d, "rlm", "config.json")
}

func WorkspaceConfigPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".rlm", "config.json")
}

func DefaultContextDir(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, "large context files")
}

func Resolve(opts ResolveOptions) (Resolved, error) {
	if opts.WorkspaceRoot == "" {
		return Resolved{}, fmt.Errorf("workspace root is required")
	}

	globalPath := GlobalConfigPath()
	workspacePath := WorkspaceConfigPath(opts.WorkspaceRoot)

	envDir := os.Getenv("RLM_CONTEXT_DIR")
	if opts.DirFlag != "" {
		return Resolved{
			WorkspaceRoot:       opts.WorkspaceRoot,
			GlobalConfigPath:    globalPath,
			WorkspaceConfigPath: workspacePath,
			ContextDir:          opts.DirFlag,
			Source:              "flag",
		}, nil
	}
	if envDir != "" {
		return Resolved{
			WorkspaceRoot:       opts.WorkspaceRoot,
			GlobalConfigPath:    globalPath,
			WorkspaceConfigPath: workspacePath,
			ContextDir:          envDir,
			Source:              "env",
		}, nil
	}

	if cfg, ok := ReadConfig(workspacePath); ok {
		if cfg.ContextDir != "" {
			return Resolved{
				WorkspaceRoot:       opts.WorkspaceRoot,
				GlobalConfigPath:    globalPath,
				WorkspaceConfigPath: workspacePath,
				ContextDir:          cfg.ContextDir,
				Source:              "workspace",
			}, nil
		}
	}

	if globalPath != "" {
		if cfg, ok := ReadConfig(globalPath); ok {
			if cfg.ContextDir != "" {
				return Resolved{
					WorkspaceRoot:       opts.WorkspaceRoot,
					GlobalConfigPath:    globalPath,
					WorkspaceConfigPath: workspacePath,
					ContextDir:          cfg.ContextDir,
					Source:              "global",
				}, nil
			}
		}
	}

	return Resolved{
		WorkspaceRoot:       opts.WorkspaceRoot,
		GlobalConfigPath:    globalPath,
		WorkspaceConfigPath: workspacePath,
		ContextDir:          DefaultContextDir(opts.WorkspaceRoot),
		Source:              "default",
	}, nil
}

func ReadConfig(path string) (Config, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, false
	}
	return c, true
}

func WriteConfig(path string, cfg Config) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func DetectWorkspaceRoot(workspacePath string) (string, error) {
	start := workspacePath
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	start, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	cur := start
	for {
		gitPath := filepath.Join(cur, ".git")
		if st, err := os.Stat(gitPath); err == nil && st.IsDir() {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return start, nil
		}
		cur = parent
	}
}

func IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
