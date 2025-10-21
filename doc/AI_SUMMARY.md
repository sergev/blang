AI Project Summary — blang

Last indexed: 2025-10-17

Overview
- A B language compiler written in Go. Frontend (lexer/parser) builds LLVM IR via llir; backend invokes clang to produce assembly, objects, or executables. A freestanding C runtime (`runtime/`) provides I/O and basic functions, linked as `-lb`.

Build and Test
- Make targets (top-level `Makefile`): `all` (go build + runtime), `install`, `uninstall`, `clean`, `test` (gotestsum), `cover`, `bench`.
- Examples `Makefile`: demonstrates CLI flags and outputs (`--emit-llvm`, `-c`, `-S`, `-O`, `-g`, `-v`, `-L`, `-l`).
- README notes: 221 tests, ~76% coverage.

CLI (main.go)
- Flags: `-o`, `--save-temps`, `--emit-llvm`, `-c`, `-S`, `-O{0..3}`, `-g`, `-v`, `-L <dir>`, `-l <lib>`, `-V`, `-h`.
- Validates inputs (.b, .ll, .s, .o, .a), constructs `CompileOptions`, assembles default library search paths, then calls `Compile`.

Compiler Orchestration (driver.go)
- Output modes: IR, Assembly, Object, Executable.
- `.b` sources are first compiled to temporary `.ll` via the frontend. Then clang is used for `-S`, `-c`, or link; temps are removed unless `--save-temps`.
- Executable: determines default output name, aggregates `.ll/.s/.o/.a` inputs, adds `-L<dirs>` and `-lb` (runtime), plus `-l<user>` libs. On Linux, uses `-static -nostdlib`.

Core Types and Utilities (options.go)
- `CompileOptions` captures inputs, output mode, optimization/debug flags, verbosity, library dirs/libs, and target word size (i64).
- `Eprintf` prints colored errors.

IR Builder (irbuilder.go)
- `Compiler` encapsulates IR state: module, current function/block, symbol tables (locals/globals/functions), string constants, labels, counters.
- Helpers to declare globals (scalars, multi-word scalars, arrays with compact representation for large zero-inited arrays), declare functions, manage blocks/labels, create string constants, and clear top-level context between top-level declarations.

Frontend — Lexing (lexer.go)
- Minimal rune-based reader with pushback; whitespace/comment skipping (`/* ... */`); identifiers; decimal/octal integers; escape sequences (B-style `*` escapes), multi-char character literals packed big-endian into a word; strings with explicit null terminator handling.

Frontend — Parsing Declarations (parser_decls.go)
- Top-level loop recognizes functions `name(...)`, vectors `name[...]`, or scalars `name ... ;`.
- Scalars: may have comma-separated initializers; multiple initializers allocate consecutive words under a single scalar name.
- Vectors: allocate `size+1` words storing a data pointer at index 0; can infer size from initializer count.
- Functions: parse parameter names, start/end function generation, then parse body via the statement parser. Clears declaration context after each top-level entity (no cross-decl leakage).

Frontend — Statements and Control (parser_stmt.go)
- Statements: blocks, null `;`, labels, `return`, `auto`, `extrn`, `if/else`, `while`, `switch/case`, `goto`, and expression statements.
- `auto` allocates locals (scalars and arrays) with B semantics for arrays (pointer in first slot, data after). Allocation order carefully follows B rules.
- `extrn` injects zero-initialized globals for referenced symbols within the current declaration context.
- Structured control flow builds SSA blocks and branches; `switch` gathers case values and constructs an LLVM `switch` in a comparison block.

Frontend — Expressions (expressions.go)
- Full precedence parser with lvalue/rvalue tracking. Returns LLVM values; dereferences lvalues to rvalues unless an lvalue is required.
- Operators: unary `!`, unary `-`, `++/--` (prefix/postfix), `*` deref, `&` address; binary `|`, `&`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `<<`, `>>`, `+`, `-`, `*`, `/`, `%`; ternary `?:`.
- Assignment supports simple `=` and compound forms; also provides a special `===` compound that stores the equality result into the left operand.
- Array indexing scales via GEP over word type; function calls support direct calls to known functions and indirect calls via function pointer variables declared with `extrn`.
- Literals: numbers (octal if leading 0), multi-character `'...'` constants (big-endian pack), strings become global constants with GEP to first element and cast to i64.

Runtime Library (runtime/)
- C sources providing B primitives: `write`, `writeb`, `printf`, `printd`, `printo`, `_char`/`lchar`, `read`, `nread`, `nwrite`, `exit`, `flush`, plus startup glue.
- Built via `runtime/Makefile` into `libb.a`; linked via `-lb` from compiler. Freestanding, syscall-based, macOS/Linux x86_64.

Examples (examples/)
- `hello.b`, `helloworld.b`, `fibonacci.b`, `fizzbuzz.b`, `e-2.b`, `b.b`, `b.pdp7`, `showcase.b` with a `Makefile` demonstrating CLI usage.

Tests
- Test files discovered:
  - `cli_test.go` — CLI behavior and options
  - `lexer_test.go` — lexing/tokenization and literals
  - `parser_test.go` — top-level declarations
  - `lang_expr_test.go` — expression semantics and operators
  - `lang_prog_test.go` — language semantics/program flow
  - `irbuilder_test.go` — IR builder helpers
  - `driver_test.go` — compilation pipeline and outputs
  - `examples_test.go` — running example programs

How to Update This Summary
- When source layout or behavior changes, update the relevant sections above.
- Keep this file short and operational: purpose per file/area, how to build/test, and where tests live.

Quick Pointers
- Entry point: `main.go` → `Compile` (driver) → frontend parse → clang.
- Options/type defs: `options.go`.
- Frontend: `lexer.go`, `parser_decls.go`, `parser_stmt.go`, `expressions.go`.
- IR state/helpers: `irbuilder.go`.
- Runtime: `runtime/` (linked via `-lb`, add `-L` to its folder when invoking `blang`).
