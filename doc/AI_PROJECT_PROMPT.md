Project AI Prompt — blang

Purpose
- This file serves as the per-project "prompt" to load before assisting on this repo.

Always Load First
- Read `doc/AI_SUMMARY.md` at session start. Use it as primary context instead of rescanning the repo unless files have changed.

Editing Rules
- Match existing formatting and indentation.
- Keep code readable; avoid deep nesting; add comments only for non-obvious rationale.
- Do not introduce linter/type errors.

Build & Test Quickstart
- Build: `make`
- Test: `make test` (uses gotestsum) or `go test ./...`
- Coverage: `make cover`

Compiler Entrypoints
- CLI: `main.go` → constructs `CompileOptions` → calls `Compile`.
- Driver: `driver.go` → IR/Asm/Object/Executable via clang.
- Frontend: `lexer.go`, `parser_decls.go`, `parser_stmt.go`, `expressions.go`.
- IR helpers: `irbuilder.go`.
- Runtime: `runtime/` linked as `-lb` (add `-L runtime_dir`).

When Unsure
- Prefer `doc/AI_SUMMARY.md` over repo-wide scans.
- If new files or structure changes are detected, update `doc/AI_SUMMARY.md` accordingly.


