---
name: rlm
description: Use the rlm CLI to list, search, peek, and chunk large context files that exceed typical context window limits. Ideal for analyzing multi-megabyte text files, codebases, or documents when you need targeted access to specific sections without loading the entire file.
metadata:
  version: "0.1.4"
  author: spanexx
---

# RLM CLI Skill

The `rlm` (Recursive Language Model) CLI is a Go-based tool for efficiently working with large context files. It helps you explore, search, and decompose files that are too large to fit into a single LLM context window.

## When to Use This Skill

Use `rlm` when:
- You need to analyze a file larger than ~100KB that won't fit in context
- You want to search across multiple large files for specific terms or patterns
- You need to inspect specific byte ranges of a file (e.g., to find line numbers)
- You want to chunk a large file into manageable pieces for sequential processing
- The user mentions "large context files", "chunking", or "searching across documents"

## Prerequisites

The `rlm` CLI must be installed globally or locally:

```bash
# Install via npm (requires Go)
npm install -g spanexx-rlm

# Or build from source
go build -o rlm ./cmd/rlm
```

Verify installation:
```bash
rlm --version
```

## Configuration

`rlm` uses a context directory to locate large files. Configuration precedence:

1. `--dir` flag (highest priority)
2. `RLM_CONTEXT_DIR` environment variable
3. Workspace config: `<workspace>/.rlm/config.json`
4. Global config: `~/.config/rlm/config.json`
5. Default: `<workspace>/large context files`

### Set context directory

```bash
# Workspace-specific
rlm config set --scope workspace --context-dir "/path/to/large-files"

# Global
rlm config set --scope global --context-dir "/path/to/large-files"

# Check current configuration
rlm config show
rlm config show --json
```

## Core Commands

### 1. List Files (`rlm files`)

List all files in the configured context directory. Useful for discovering what's available.

```bash
rlm files
```

Output (newline-separated paths):
```
/home/user/project/large context files/document.txt
/home/user/project/large context files/codebase.py
```

JSON output for agent parsing:
```bash
rlm files --json
```

### 2. Search (`rlm search`)

Search for text or regex patterns across all files in the context directory. Returns JSON by default.

**Basic search:**
```bash
rlm search --query "search term"
```

**Case-insensitive:**
```bash
rlm search --query "SEARCH TERM" --ignore-case
```

**Regex search:**
```bash
rlm search --query "EX-2\\.1" --regex
```

**Limit results:**
```bash
rlm search --query "term" --max-matches 20 --max-per-file 5
```

**Exit codes:**
- `0` - matches found
- `1` - no matches
- `2` - error

**JSON output format:**
```json
{
  "query": "search term",
  "context_dir": "/path/to/context",
  "matches": [
    {
      "path": "/path/to/file.txt",
      "line": 42,
      "column": 5,
      "snippet": "context around the match..."
    }
  ],
  "scanned_files": 5,
  "duration_ms": 1
}
```

### 3. Peek (`rlm peek`)

Extract a specific byte range from a file. Useful for inspecting sections around search results.

**Peek first ~8KB:**
```bash
rlm peek "file.txt" --start 0
```

**Specific range:**
```bash
rlm peek "file.txt" --start 1000 --end 2000
```

**Peek to EOF:**
```bash
rlm peek "file.txt" --start 0 --end -1
```

**Use absolute paths:**
```bash
rlm peek "/absolute/path/to/file.txt" --start 0 --end 1000
```

### 4. Chunk (`rlm chunk`)

Split a file into fixed-size chunks with optional overlap. Useful for processing large files sequentially.

**Basic chunking:**
```bash
rlm chunk "file.txt" --size 200000 --overlap 0
```

**With overlap:**
```bash
rlm chunk "file.txt" --size 200000 --overlap 2000
```

**Custom output directory:**
```bash
rlm chunk "file.txt" --size 200000 --overlap 2000 --out "/tmp/rlm-chunks"
```

**Output:** Creates files named `file.txt.000`, `file.txt.001`, etc.

### 5. Display Documentation (`rlm readme`)

Show the project README for reference.

```bash
rlm readme
```

### 6. Version Check (`rlm --version`)

Display the installed version.

```bash
rlm --version
```

## Agent Workflow Pattern

When working with large files, follow this pattern:

1. **Discover files:**
   ```bash
   rlm files
   ```

2. **Search for relevant content:**
   ```bash
   rlm search --query "keyword"
   ```

3. **Inspect matches:**
   ```bash
   rlm peek "file.txt" --start <byte_offset> --end <byte_offset + 2000>
   ```

4. **If file is very large, chunk it:**
   ```bash
   rlm chunk "file.txt" --size 200000 --overlap 2000
   ```

5. **Process chunks sequentially:**
   Read each chunk file (`file.txt.000`, `file.txt.001`, ...) and analyze.

## Tips for Agents

- **Always check context directory first** with `rlm config show` to ensure you're looking in the right place
- **Use JSON output** (`--json` flag) when parsing results programmatically
- **Prefer `rlm search` over reading entire files** - it's faster and uses less context
- **Use `rlm peek` to get context around search results** before deciding to chunk
- **Chunk size of 200KB-500KB** is usually optimal for most LLMs
- **Add overlap (1-2KB) when chunking** to avoid cutting off mid-sentence or mid-context
- **Binary files are automatically skipped** during search operations
- **All text-based formats are supported** (`.txt`, `.md`, `.json`, `.html`, `.csv`, etc.)

## Common Errors

**"no such file or directory"**: The context directory is misconfigured. Run `rlm config show` and fix with `rlm config set`.

**"Unknown command"**: You're using an old version of `rlm`. Update with `npm install -g spanexx-rlm@latest`.

**Search returns no matches but you know the term exists**: Try case-insensitive search (`--ignore-case`) or regex mode (`--regex`).

## Example Agent Session

```bash
# User: "Find all mentions of 'SOFTBANK' in the legal documents"

# Agent:
rlm config show  # Check context directory
rlm search --query "SOFTBANK"  # Search across all files

# Output shows matches in EX-2.1.txt at lines 30, 2190, 2191
rlm peek "EX-2.1.txt" --start 0 --end 2000  # Inspect first match
rlm peek "EX-2.1.txt" --start 100000 --end 102000  # Inspect later matches
```

## References

For more details on RLM architecture and theory, see the project README:
```bash
rlm readme
```
