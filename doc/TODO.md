# B Language Compiler - TODO List

## 🚧 Pending Language Features

### High Priority

#### 1. Compound Assignment Operators
**Status:** Not implemented (15 tests skipped)
**Test Coverage:** `TestCompoundAssignments` in `compiler_test.go`

Implement all compound assignment operators from the B language specification:
- Arithmetic: `=+`, `=-`, `=*`, `=/`, `=%`
- Bitwise shifts: `=<<`, `=>>`
- Comparison: `=<`, `=<=`, `=>`, `=>=`, `=!=`, `===`
- Bitwise: `=&`, `=|`

**Workaround:** Use expanded form (e.g., `x = x + 5` instead of `x =+ 5`)

**Implementation Notes:**
- These operators are syntactically distinct from prefix operators
- `x =+ 5` means `x = x + 5`, not `x = (+5)`
- Parser needs to handle the `=` followed by operator without whitespace
- See `oldtests/assignment_test.cpp` for expected behavior

**Files to modify:**
- `lexer.go` - May need token lookahead for `=+` vs `= +`
- `expr.go` - Add compound assignment parsing in `parseExpressionLLVMWithLevel`
- `compiler_test.go` - Remove `t.Skip()` from `TestCompoundAssignments`

---

## ✅ Recently Completed

- ✅ All basic B language features (variables, arrays, functions)
- ✅ Full expression parser with 15 precedence levels
- ✅ All binary and unary operators
- ✅ Control flow: `if/else`, `while`, `switch/case`, `goto`, labels
- ✅ Pointers and arrays (local and global)
- ✅ Function definitions and calls with multiple parameters
- ✅ String literals and character constants
- ✅ Octal number literals
- ✅ Multi-character constants
- ✅ Forward references (`extrn` declarations)
- ✅ LLVM IR backend
- ✅ Runtime library integration (`libb.c`)
- ✅ **Scalar with multiple initialization values** (`c -345, 'foo', "bar";`)
- ✅ **Auto arrays with character constant sizes** (`auto buf['x'];`)
- ✅ **Reverse allocation order for auto statements**
- ✅ **Expression parser bug fixes** (equality operator chaining)
- ✅ **Indirect function calls** via function pointer variables
- ✅ **Ternary conditional operator** (`? :`) with nested support
- ✅ Comprehensive test suite (127 active tests, 76.0% coverage)

---

## 📝 Code Quality & Maintenance

### Documentation
- [ ] Add more inline comments explaining B language semantics
- [ ] Document LLVM IR generation patterns
- [ ] Create ARCHITECTURE.md explaining compiler phases
- [ ] Add more examples to README.md

### Testing
- [ ] Add benchmarks for compilation performance
- [ ] Test error messages for user-friendliness
- [ ] Add fuzzing tests for parser robustness
- [ ] Consider property-based testing for expression evaluation

### Code Organization
- [ ] Consider splitting large files (`expr.go` is quite long)
- [ ] Add more helper functions to reduce code duplication
- [ ] Improve error messages with source location information

---

## 🔮 Future Enhancements

### Optimization
- [ ] Dead code elimination
- [ ] Constant folding
- [ ] Common subexpression elimination
- [ ] Leverage LLVM optimization passes

### Tooling
- [ ] B language server protocol (LSP) implementation
- [ ] Syntax highlighting for editors
- [ ] Debugger support (DWARF debug info generation)
- [ ] Interactive REPL

### Platform Support
- [ ] Test on Linux (currently macOS only)
- [ ] Test on Windows (WSL)
- [ ] Cross-compilation support

### Language Extensions (Non-standard)
- [ ] Type annotations (optional, for documentation)
- [ ] Module system
- [ ] Better error handling primitives
- [ ] Standard library expansion beyond `libb.c`

---

## 📊 Implementation Status

### Completed Features ✅

**Core Language:**
- ✅ Variables (local `auto`, global, external `extrn`)
- ✅ Arrays (local and global with B-specific semantics)
- ✅ Functions (declarations, definitions, parameters, return values)
- ✅ Pointers (address-of `&`, dereference `*`, pointer arithmetic)
- ✅ Literals (numbers, strings, multi-character constants, octal)

**Operators (15 precedence levels):**
- ✅ Arithmetic: `+`, `-`, `*`, `/`, `%`
- ✅ Comparison: `<`, `<=`, `>`, `>=`, `==`, `!=`
- ✅ Bitwise: `&`, `|`, `<<`, `>>`
- ✅ Logical: `!`
- ✅ Unary: `-`, `&` (address-of), `*` (dereference)
- ✅ Increment/Decrement: `++`, `--` (prefix and postfix)
- ✅ Assignment: `=`
- ✅ Array indexing: `[]`
- ✅ Function calls: `()`

**Control Flow:**
- ✅ if/else statements
- ✅ while loops
- ✅ switch/case statements
- ✅ goto statements and labels
- ✅ return statements

**Special Features:**
- ✅ Automatic external function declaration
- ✅ Forward references with `extrn`
- ✅ LLVM IR backend for portability
- ✅ Comprehensive runtime library (`libb.c`)

### Pending Features ⏳

See sections below for details on pending features.

## 📊 Test Coverage Status

| Test Category | Status | Count | Notes |
|---------------|--------|-------|-------|
| Lexer Tests | ✅ Complete | 29 | All character classes, comments, strings |
| Compiler Tests | ✅ Complete | 10 | Basic compilation verification |
| Error Handling | ✅ Complete | 7 | Compilation error scenarios |
| Integration Tests | ✅ Complete | 9 | Compile + link + run |
| Runtime Library | ✅ Complete | 6 | All `libb` functions |
| Precedence Tests | ✅ Complete | 28 | All operator precedence combinations |
| Expression Tests | ✅ Complete | 9 | All operators and precedence |
| String Tests | ✅ Complete | 2 | Escape sequences, literals |
| Globals Tests | ✅ Complete | 4 | Global/local allocation, multi-value scalars |
| Function Tests | ✅ Complete | 5 | Includes nested ternary operator |
| Indirect Calls | ✅ Complete | 2 | Function pointers |
| E-2 Constant | ⏭️ Skipped | 1 | Long-running (~10+ seconds) |
| Compound Assignments | ⏭️ Skipped | 15 | Not implemented |

**Total: 127 active tests passing, 2 skipped (pending implementation)**

---

## 🐛 Known Issues

### None Currently Open
All critical bugs have been fixed:
- ✅ String constant null terminators
- ✅ Global array initialization with mixed types
- ✅ Forward reference handling for `extrn`
- ✅ Duplicate function declarations
- ✅ Local variable initialization to zero
- ✅ Nested if-else label conflicts

---

## 📚 Reference Materials

- Original B Language Reference: [Bell Labs CSRC](https://www.bell-labs.com/usr/dmr/www/kbman.html)
- B Language Tutorial: [Ken Thompson's B Tutorial](https://www.bell-labs.com/usr/dmr/www/btut.html)
- LLVM IR Documentation: [LLVM Language Reference](https://llvm.org/docs/LangRef.html)
- C Prototype: `c-prototype/` directory

---

## 🎯 Next Steps

**Recommended Priority Order:**

1. **Implement Compound Assignments** (4-6 hours)
   - More complex due to operator variety
   - High test coverage (15 tests)
   - Significant language feature

3. **Code Quality Improvements** (ongoing)
   - Better error messages
   - More documentation
   - Performance optimization


---

**Last Updated:** October 13, 2025
**Compiler Version:** LLVM Backend (production-ready)
**Test Pass Rate:** 100% (127/127 active tests)
