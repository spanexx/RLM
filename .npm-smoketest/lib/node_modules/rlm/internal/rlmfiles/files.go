package rlmfiles

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func List(contextDir string) ([]FileInfo, error) {
	var out []FileInfo
	err := filepath.WalkDir(contextDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		out = append(out, FileInfo{Path: path, Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
