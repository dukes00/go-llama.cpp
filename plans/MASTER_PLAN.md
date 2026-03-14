# go-llama.cpp — Master Implementation Plan

This document is for the HUMAN maintainer. Don't feed this to Cline.

## Phases

| Phase | Name | Depends On | What You Get |
|-------|------|-----------|--------------|
| 1 | Foundation | nothing | Types, home dir init, config CRUD, unit tests |
| 2 | CLI + Serve | Phase 1 | Cobra CLI, `serve` command that launches llama-server |
| 3 | Registry | Phase 2 | `list` and `kill` commands, PID tracking |
| 4 | Wizard | Phase 1+2 | Interactive preset save/load, --override logic |
| 5 | Downloads | Phase 1 | `download` command, HuggingFace GGUF fetching |

## How To Run Each Phase
1. Open Cline
2. Type: `Read plans/PHASE_N.md and implement it, step by step. Run tests after each file.`
3. Review the output. If something is wrong, give specific corrections.
4. After phase is done, sanity check: `go build ./... && go test ./... -v`
5. Commit: `git add -A && git commit -m "phase N: <description>"`
6. Move to next phase.

## What To Do If The Model Gets Confused
- If it tries to use a package that doesn't exist: tell it the correct import path.
- If it creates circular imports: tell it to move the type to the lower-level package.
- If it forgets earlier context: say "Read internal/config/config.go first, then continue."
- If tests fail: paste the error output and say "Fix this test failure."