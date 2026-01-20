# RLM

`rlm` is a fast Go CLI for working with *large context files*: listing files, searching text, peeking byte ranges, and chunking files for LLM/agent workflows.

## Install

### npm (recommended)

Requires **Go** to be installed (the package builds the Go binary during install).

```bash
npm install -g spanexx-rlm@latest
rlm --version
```

### From source

```bash
go build -o rlm ./cmd/rlm
./rlm --help
```

## Quick start

### 1) Set your context directory

The context directory is where you keep your large files.

Precedence:

`--dir` > `RLM_CONTEXT_DIR` > workspace config > global config > default

Default:

`<workspace>/large context files`

Examples:

```bash
# Per-workspace
rlm config set --scope workspace --context-dir "/absolute/path/to/large context files"

# Global
rlm config set --scope global --context-dir "/absolute/path/to/large context files"

# Verify
rlm config show --json
```

### 2) List files

```bash
rlm files
```

### 3) Search

JSON output is default (agent-friendly).

```bash
rlm search --query "SOFTBANK"
rlm search --query "term" --ignore-case
rlm search --query "EX-2\\.1" --regex
rlm search --query "term" --max-matches 20 --max-per-file 5
```

Exit codes:

- `0` matches found
- `1` no matches
- `2` error

### 4) Peek (byte ranges)

```bash
rlm peek "somefile.txt" --start 0
rlm peek "somefile.txt" --start 0 --end 2000
rlm peek "somefile.txt" --start 0 --end -1   # to EOF
```

### 5) Chunk

```bash
rlm chunk "somefile.txt" --size 200000 --overlap 2000
rlm chunk "somefile.txt" --size 200000 --overlap 2000 --out "/tmp/rlm-chunks"
```

## Docs

```bash
rlm readme
```

## Repository

- **Repo URL**: `https://github.com/spanexx/RLM.git`
- **npm package**: `spanexx-rlm`

## Repository structure

```
.
├── .agent/
│   └── skills/
│       └── rlm/
│           └── SKILL.md
├── cmd/
│   └── rlm/
│       └── main.go
├── internal/
│   ├── rlmchunk/
│   ├── rlmconfig/
│   ├── rlmfiles/
│   └── rlmsearch/
├── scripts/
│   └── postinstall.js
├── large context files/
├── LICENSE
├── package.json
└── deployment.md
```

## License

MIT — see [LICENSE](LICENSE).
