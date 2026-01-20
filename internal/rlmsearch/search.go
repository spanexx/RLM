package rlmsearch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	ContextDir   string
	Query        string
	Regex        bool
	IgnoreCase   bool
	MaxMatches   int
	MaxPerFile   int
	MaxLineChars int
}

type Match struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Snippet string `json:"snippet"`
}

type Result struct {
	Query      string  `json:"query"`
	ContextDir string  `json:"context_dir"`
	Matches    []Match `json:"matches"`
	Files      int     `json:"scanned_files"`
	DurationMs int64   `json:"duration_ms"`
}

func SearchDir(opts Options) (Result, error) {
	if opts.ContextDir == "" {
		return Result{}, fmt.Errorf("context_dir is required")
	}
	if opts.Query == "" {
		return Result{}, fmt.Errorf("query is required")
	}
	if opts.MaxMatches <= 0 {
		opts.MaxMatches = 50
	}
	if opts.MaxPerFile <= 0 {
		opts.MaxPerFile = 20
	}
	if opts.MaxLineChars <= 0 {
		opts.MaxLineChars = 800
	}

	var re *regexp.Regexp
	var err error
	qFixed := opts.Query
	if opts.Regex {
		pattern := opts.Query
		if opts.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return Result{}, err
		}
	} else {
		if opts.IgnoreCase {
			qFixed = strings.ToLower(qFixed)
		}
	}

	res := Result{Query: opts.Query, ContextDir: opts.ContextDir}

	skipDirs := map[string]bool{
		".git":         true,
		".rlm":         true,
		"node_modules": true,
	}

	err = filepath.WalkDir(opts.ContextDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			if skipDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip hidden files.
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		res.Files++

		matchesInFile := 0
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		// Binary detection: if there are NUL bytes in the first chunk, skip.
		if isLikelyBinary(f) {
			_ = f.Close()
			return nil
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			_ = f.Close()
			return err
		}

		reader := bufio.NewReaderSize(f, 256*1024)
		lineNo := 0
		colBase := 0
		const maxFragmentBytes = 256 * 1024
		for {
			if len(res.Matches) >= opts.MaxMatches {
				_ = f.Close()
				return filepath.SkipAll
			}
			if matchesInFile >= opts.MaxPerFile {
				break
			}

			frag, gotNL, readErr := readLineFragment(reader, maxFragmentBytes)
			if readErr != nil {
				_ = f.Close()
				return readErr
			}
			if len(frag) == 0 && !gotNL {
				break
			}

			// If we're at the start of a new line, allocate a new line number.
			if colBase == 0 {
				lineNo++
			}
			base := colBase

			line := string(frag)
			check := line
			if opts.IgnoreCase && !opts.Regex {
				check = strings.ToLower(check)
			}

			if opts.Regex {
				loc := re.FindStringIndex(line)
				if loc != nil {
					matchesInFile++
					res.Matches = append(res.Matches, Match{
						Path:    path,
						Line:    lineNo,
						Column:  base + loc[0] + 1,
						Snippet: trimLine(line, opts.MaxLineChars),
					})
				}
			} else {
				idx := strings.Index(check, qFixed)
				if idx >= 0 {
					matchesInFile++
					res.Matches = append(res.Matches, Match{
						Path:    path,
						Line:    lineNo,
						Column:  base + idx + 1,
						Snippet: trimLine(line, opts.MaxLineChars),
					})
				}
			}

			if gotNL {
				colBase = 0
			} else {
				// Continuation of a very long line.
				colBase += len(frag)
			}
		}
		_ = f.Close()
		return nil
	})
	if err != nil {
		return Result{}, err
	}

	return res, nil
}

func isLikelyBinary(f *os.File) bool {
	const sampleSize = 4096
	buf := make([]byte, sampleSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]
	for _, b := range buf {
		if b == 0x00 {
			return true
		}
	}
	return false
}

func readLineFragment(r *bufio.Reader, maxBytes int) ([]byte, bool, error) {
	if maxBytes <= 0 {
		maxBytes = 256 * 1024
	}

	frag, err := r.ReadSlice('\n')
	if err == nil {
		return frag, true, nil
	}
	if err == bufio.ErrBufferFull {
		if len(frag) > maxBytes {
			frag = frag[:maxBytes]
		}
		return frag, false, nil
	}
	if err == io.EOF {
		if len(frag) == 0 {
			return nil, false, nil
		}
		return frag, false, nil
	}
	return nil, false, err
}

func trimLine(s string, max int) string {
	s = strings.TrimRight(s, "\r\n")
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}
