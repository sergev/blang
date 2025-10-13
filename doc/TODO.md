# B Language Compiler - TODO List

## 🚧 Pending Language Features

### None - All Core Features Complete! 🎉

All B language features have been implemented and tested.

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
- ✅ **Compound assignment operators** (all 15: `=+`, `=-`, `=*`, `=/`, `=%`, `=<<`, `=>>`, `=&`, `=|`, `=<`, `=<=`, `=>`, `=>=`, `=!=`, `===`)
- ✅ **Nested while loops** with unique label generation
- ✅ Comprehensive test suite (144 active tests, 76.6% coverage)

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
| Nested Loops | ✅ Complete | 2 | Nested while loops with unique labels |
| Compound Assignments | ✅ Complete | 15 | All 15 compound operators |
| E-2 Constant | ⏭️ Skipped | 1 | Long-running (~10+ seconds) |

**Total: 144 active tests passing, 1 skipped (long-running computation only)**

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

## 🎯 Next Steps

**All core features complete!** Consider:

1. **Code Quality Improvements** (ongoing)
   - Better error messages with line numbers
   - More inline documentation
   - Performance optimization

2. **Extended Testing**
   - Enable e-2 constant test (currently skipped for runtime)
   - Add more complex integration tests
   - Fuzzing for parser robustness

3. **Advanced Features** (optional)
   - Optimization passes
   - Debug information (DWARF)
   - Additional platforms

---

**Last Updated:** October 13, 2025
**Compiler Version:** LLVM Backend (production-ready, feature-complete)
**Test Pass Rate:** 100% (144/144 active tests)
**Feature Completeness:** 100% - All B language features implemented!
