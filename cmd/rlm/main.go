package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Brainqub3/claude_code_RLM/internal/rlmchunk"
	"github.com/Brainqub3/claude_code_RLM/internal/rlmconfig"
	"github.com/Brainqub3/claude_code_RLM/internal/rlmfiles"
	"github.com/Brainqub3/claude_code_RLM/internal/rlmsearch"
)

type exitCodeError struct {
	code int
	err  error
}

func (e exitCodeError) Error() string { return e.err.Error() }
func (e exitCodeError) Unwrap() error { return e.err }

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	if len(argv) == 0 {
		printUsage()
		return 2
	}

	cmd := argv[0]
	args := argv[1:]

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "readme":
		return cmdReadme(args)
	case "docs":
		return cmdDocs(args)
	case "config":
		return cmdConfig(args)
	case "files":
		return cmdFiles(args)
	case "search":
		return cmdSearch(args)
	case "peek":
		return cmdPeek(args)
	case "chunk":
		return cmdChunk(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, strings.TrimSpace(`rlm - retrieve context from large files

Usage:
  rlm <command> [args]

Commands:
	 docs     Show built-in project docs (README / SKILL)
  config   Show/set configuration (global or per-workspace)
  files    List files in the configured context directory
  search   Search for a string/regex across context files
  peek     Extract a byte range from a file
  chunk    Write fixed-size chunks of a file to disk

Environment:
  RLM_CONTEXT_DIR  Overrides configured context directory

Precedence for context directory:
  --dir > RLM_CONTEXT_DIR > workspace config > global config > default

`)+"\n")
}

func cmdDocs(argv []string) int {
	if len(argv) == 0 {
		fmt.Fprintln(os.Stderr, "Missing subcommand: readme|skill")
		return 2
	}

	sub := argv[0]

	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	switch sub {
	case "readme":
		return catFile(filepath.Join(wsRoot, "README.md"))
	case "skill":
		return catFile(filepath.Join(wsRoot, ".claude", "skills", "rlm", "SKILL.md"))
	default:
		fmt.Fprintf(os.Stderr, "Unknown docs subcommand: %s\n", sub)
		return 2
	}
}

func catFile(path string) int {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	return 0
}

func cmdReadme(argv []string) int {
	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	return catFile(filepath.Join(wsRoot, "claude_code_RLM", "README.md"))
}

func cmdConfig(argv []string) int {
	if len(argv) == 0 {
		fmt.Fprintln(os.Stderr, "Missing subcommand: show|set")
		return 2
	}

	sub := argv[0]
	args := argv[1:]

	switch sub {
	case "show":
		fs := flag.NewFlagSet("rlm config show", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		workspace := fs.String("workspace", "", "Workspace path (defaults to current working directory)")
		jsonOut := fs.Bool("json", false, "Output JSON")
		if err := fs.Parse(args); err != nil {
			return 2
		}

		wsRoot, err := rlmconfig.DetectWorkspaceRoot(*workspace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 2
		}

		resolved, err := rlmconfig.Resolve(rlmconfig.ResolveOptions{WorkspaceRoot: wsRoot})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 2
		}

		if *jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(resolved)
			return 0
		}

		fmt.Println("rlm config")
		fmt.Printf("  workspace_root: %s\n", resolved.WorkspaceRoot)
		fmt.Printf("  global_config: %s\n", resolved.GlobalConfigPath)
		fmt.Printf("  workspace_config: %s\n", resolved.WorkspaceConfigPath)
		fmt.Printf("  context_dir: %s\n", resolved.ContextDir)
		fmt.Printf("  source: %s\n", resolved.Source)
		return 0

	case "set":
		fs := flag.NewFlagSet("rlm config set", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		scope := fs.String("scope", "workspace", "Config scope: workspace|global")
		workspace := fs.String("workspace", "", "Workspace path (defaults to current working directory)")
		contextDir := fs.String("context-dir", "", "Directory containing large context files")
		if err := fs.Parse(args); err != nil {
			return 2
		}
		if strings.TrimSpace(*contextDir) == "" {
			fmt.Fprintln(os.Stderr, "--context-dir is required")
			return 2
		}

		wsRoot, err := rlmconfig.DetectWorkspaceRoot(*workspace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 2
		}

		path := ""
		switch *scope {
		case "workspace":
			path = rlmconfig.WorkspaceConfigPath(wsRoot)
		case "global":
			path = rlmconfig.GlobalConfigPath()
		default:
			fmt.Fprintln(os.Stderr, "--scope must be workspace or global")
			return 2
		}

		cfg := rlmconfig.Config{ContextDir: *contextDir}
		if err := rlmconfig.WriteConfig(path, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 2
		}
		fmt.Printf("Wrote config: %s\n", path)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Unknown config subcommand: %s\n", sub)
		return 2
	}
}

func cmdFiles(argv []string) int {
	fs := flag.NewFlagSet("rlm files", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dirFlag := fs.String("dir", "", "Override context directory")
	jsonOut := fs.Bool("json", false, "Output JSON")
	if err := fs.Parse(argv); err != nil {
		return 2
	}

	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	resolved, err := rlmconfig.Resolve(rlmconfig.ResolveOptions{WorkspaceRoot: wsRoot, DirFlag: *dirFlag})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	files, err := rlmfiles.List(resolved.ContextDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"context_dir": resolved.ContextDir, "files": files})
		return 0
	}

	for _, f := range files {
		fmt.Println(f.Path)
	}
	return 0
}

func cmdSearch(argv []string) int {
	fs := flag.NewFlagSet("rlm search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dirFlag := fs.String("dir", "", "Override context directory")
	query := fs.String("query", "", "Query string or regex")
	regex := fs.Bool("regex", false, "Treat query as regex")
	fixed := fs.Bool("fixed", false, "Treat query as fixed substring")
	ignoreCase := fs.Bool("ignore-case", false, "Case-insensitive matching")
	maxMatches := fs.Int("max-matches", 50, "Maximum total matches")
	maxPerFile := fs.Int("max-per-file", 20, "Maximum matches per file")
	jsonOut := fs.Bool("json", true, "Output JSON")
	if err := fs.Parse(argv); err != nil {
		return 2
	}

	q := strings.TrimSpace(*query)
	if q == "" {
		fmt.Fprintln(os.Stderr, "--query is required")
		return 2
	}

	if *regex && *fixed {
		fmt.Fprintln(os.Stderr, "--regex and --fixed are mutually exclusive")
		return 2
	}
	if !*regex && !*fixed {
		*fixed = true
	}

	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	resolved, err := rlmconfig.Resolve(rlmconfig.ResolveOptions{WorkspaceRoot: wsRoot, DirFlag: *dirFlag})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	start := time.Now()
	result, err := rlmsearch.SearchDir(rlmsearch.Options{
		ContextDir:   resolved.ContextDir,
		Query:        q,
		Regex:        *regex,
		IgnoreCase:   *ignoreCase,
		MaxMatches:   *maxMatches,
		MaxPerFile:   *maxPerFile,
		MaxLineChars: 800,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	result.DurationMs = time.Since(start).Milliseconds()

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		if len(result.Matches) == 0 {
			return 1
		}
		return 0
	}

	for _, m := range result.Matches {
		fmt.Printf("%s:%d:%s\n", m.Path, m.Line, m.Snippet)
	}
	if len(result.Matches) == 0 {
		return 1
	}
	return 0
}

func cmdPeek(argv []string) int {
	fs := flag.NewFlagSet("rlm peek", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dirFlag := fs.String("dir", "", "Override context directory")
	start := fs.Int64("start", 0, "Start byte offset")
	end := fs.Int64("end", 0, "End byte offset (exclusive). Use -1 for EOF")
	jsonOut := fs.Bool("json", false, "Output JSON")
	argv = normalizeAndReorderArgs(fs, argv)
	if err := fs.Parse(argv); err != nil {
		return 2
	}

	args := fs.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: rlm peek <file> --start N [--end M] (M=-1 for EOF)")
		return 2
	}
	if *end < *start {
		fmt.Fprintln(os.Stderr, "--end must be >= --start")
		return 2
	}

	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	resolved, err := rlmconfig.Resolve(rlmconfig.ResolveOptions{WorkspaceRoot: wsRoot, DirFlag: *dirFlag})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	p := resolveFileArg(resolved.ContextDir, args[0])
	f, err := os.Open(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	s := *start
	e := *end
	if s < 0 {
		s = 0
	}
	if e < 0 {
		e = st.Size()
	} else if e == 0 {
		const defaultPeekBytes = int64(8192)
		e = s + defaultPeekBytes
		if e > st.Size() {
			e = st.Size()
		}
	} else if e > st.Size() {
		e = st.Size()
	}
	if e < s {
		e = s
	}

	if _, err := f.Seek(s, io.SeekStart); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	toRead := e - s
	buf, err := io.ReadAll(io.LimitReader(f, toRead))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	out := string(buf)
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"path": p, "start": s, "end": e, "text": out})
		return 0
	}
	fmt.Print(out)
	return 0
}

func cmdChunk(argv []string) int {
	fs := flag.NewFlagSet("rlm chunk", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dirFlag := fs.String("dir", "", "Override context directory")
	size := fs.Int("size", 200_000, "Chunk size in bytes")
	overlap := fs.Int("overlap", 0, "Overlap in bytes")
	outDir := fs.String("out", "", "Output directory (default: <workspace>/.rlm/chunks)")
	prefix := fs.String("prefix", "chunk", "Chunk filename prefix")
	jsonOut := fs.Bool("json", true, "Output JSON")
	argv = normalizeAndReorderArgs(fs, argv)
	if err := fs.Parse(argv); err != nil {
		return 2
	}

	args := fs.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: rlm chunk <file> [--size N --overlap M --out DIR]")
		return 2
	}
	if *size <= 0 {
		fmt.Fprintln(os.Stderr, "--size must be > 0")
		return 2
	}
	if *overlap < 0 || *overlap >= *size {
		fmt.Fprintln(os.Stderr, "--overlap must be >= 0 and < --size")
		return 2
	}

	wsRoot, err := rlmconfig.DetectWorkspaceRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}
	resolved, err := rlmconfig.Resolve(rlmconfig.ResolveOptions{WorkspaceRoot: wsRoot, DirFlag: *dirFlag})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	p := resolveFileArg(resolved.ContextDir, args[0])
	if *outDir == "" {
		*outDir = filepath.Join(wsRoot, ".rlm", "chunks")
	}

	paths, err := rlmchunk.WriteChunks(rlmchunk.Options{
		InPath:   p,
		OutDir:   *outDir,
		Size:     *size,
		Overlap:  *overlap,
		Prefix:   *prefix,
		Encoding: "utf-8",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 2
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"in": p, "out_dir": *outDir, "chunks": paths})
		return 0
	}

	for _, cp := range paths {
		fmt.Println(cp)
	}
	return 0
}

func resolveFileArg(contextDir, arg string) string {
	if filepath.IsAbs(arg) {
		return arg
	}
	return filepath.Join(contextDir, arg)
}

type boolFlag interface {
	IsBoolFlag() bool
}

func normalizeAndReorderArgs(fs *flag.FlagSet, argv []string) []string {
	// 1) Normalize GNU-style flags: "--foo" -> "-foo".
	norm := make([]string, 0, len(argv))
	for _, a := range argv {
		if strings.HasPrefix(a, "--") {
			norm = append(norm, "-"+strings.TrimPrefix(a, "--"))
		} else {
			norm = append(norm, a)
		}
	}

	// 2) Reorder so flags (and their values) come before positional args.
	// This allows: rlm peek <file> --start 0 --end 10
	flags := make([]string, 0, len(norm))
	pos := make([]string, 0, len(norm))

	for i := 0; i < len(norm); i++ {
		a := norm[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)

			name := strings.TrimLeft(a, "-")
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				name = name[:eq]
			}
			f := fs.Lookup(name)
			isBool := false
			if f != nil {
				if bf, ok := f.Value.(boolFlag); ok && bf.IsBoolFlag() {
					isBool = true
				}
			}

			if !isBool && !strings.Contains(a, "=") {
				if i+1 < len(norm) && !strings.HasPrefix(norm[i+1], "-") {
					flags = append(flags, norm[i+1])
					i++
				}
			}
			continue
		}
		pos = append(pos, a)
	}

	return append(flags, pos...)
}
