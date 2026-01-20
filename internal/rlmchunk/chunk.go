package rlmchunk

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Options struct {
	InPath   string
	OutDir   string
	Size     int
	Overlap  int
	Prefix   string
	Encoding string
}

func WriteChunks(opts Options) ([]string, error) {
	if opts.InPath == "" {
		return nil, fmt.Errorf("in_path is required")
	}
	if opts.OutDir == "" {
		return nil, fmt.Errorf("out_dir is required")
	}
	if opts.Size <= 0 {
		return nil, fmt.Errorf("size must be > 0")
	}
	if opts.Overlap < 0 || opts.Overlap >= opts.Size {
		return nil, fmt.Errorf("overlap must be >= 0 and < size")
	}
	if opts.Prefix == "" {
		opts.Prefix = "chunk"
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, err
	}

	f, err := os.Open(opts.InPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	step := opts.Size - opts.Overlap
	buf := make([]byte, opts.Size)

	var out []string
	for i := 0; ; i++ {
		name := fmt.Sprintf("%s_%04d.txt", opts.Prefix, i)
		p := filepath.Join(opts.OutDir, name)

		n, readErr := io.ReadFull(f, buf)
		if readErr != nil {
			// io.ReadFull returns err != nil even if n>0 (e.g. EOF). We still want to write the partial chunk.
			if readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
				return nil, readErr
			}
		}

		if n == 0 {
			break
		}

		if err := os.WriteFile(p, buf[:n], 0o644); err != nil {
			return nil, err
		}
		out = append(out, p)

		// Stop if we reached the end.
		pos, _ := f.Seek(0, io.SeekCurrent)
		if pos >= st.Size() {
			break
		}

		// Apply overlap by seeking backwards.
		if opts.Overlap > 0 {
			if _, err := f.Seek(int64(-opts.Overlap), io.SeekCurrent); err != nil {
				return nil, err
			}
		}

		// Defensive: if step is 0 (shouldn't happen due to validation), avoid infinite loop.
		_ = step
	}

	return out, nil
}
