# Test Coverage Improvement Plan

**Current Coverage:** 76.6%
**Target Coverage:** 85%+

## Executive Summary

This document outlines a comprehensive plan to increase test coverage by targeting untested and under-tested code paths. The analysis identified 8 functions with 0% coverage and 23 functions with less than 70% coverage.

---

## 📊 Coverage Analysis by File

### High Impact (0% Coverage - Critical)

These functions are **completely untested** and should be prioritized:

| Function | File | Coverage | Impact | Priority |
|----------|------|----------|--------|----------|
| `LoadValue` | codegen.go | 0% | High | **P0** |
| `StoreValue` | codegen.go | 0% | High | **P0** |
| `PeekChar` | lexer.go | 0% | Medium | **P1** |
| `IsEOF` | lexer.go | 0% | Low | P2 |
| `parseKeywordOrExpression` | parser.go | 0% | Low | P2 |
| `Eprintf` | compiler.go | 0% | Low | P3 |
| `main` | main.go | 0% | N/A | Skip |
| `usage` | main.go | 0% | N/A | Skip |

### Medium Impact (50-70% Coverage)

These functions are **partially tested** and have significant gaps:

| Function | File | Coverage | Missing Paths | Priority |
|----------|------|----------|---------------|----------|
| `ExpectChar` | lexer.go | 50% | Error cases | **P1** |
| `parseUnary` | expr.go | 52.9% | Prefix ++, edge cases | **P0** |
| `NewBlock` | codegen.go | 60% | Edge cases | P2 |
| `parseVector` | parser.go | 65.3% | Error paths | **P1** |
| `parseReturn` | parser.go | 65% | Error paths | **P1** |
| `parseCase` | control.go | 67.9% | Complex case chains | **P1** |
| `parseIvalConst` | parser.go | 67.7% | Identifier refs | **P1** |
| `ParseDeclarations` | parser.go | 68% | Error handling | P2 |

---

## 🎯 Proposed Test Cases

### P0: Critical Coverage Gaps (Target: +5-7%)

#### 1. **Unary Operator Tests** (expr.go:parseUnary 52.9% → 90%)

```go
func TestUnaryOperators(t *testing.T) {
    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "prefix_increment",
            code: `main() {
                auto x;
                x = 5;
                printf("%d*n", ++x);  /* Should print 6 */
                printf("%d*n", x);    /* Should print 6 */
            }`,
            wantStdout: "6\n6\n",
        },
        {
            name: "prefix_decrement",
            code: `main() {
                auto x;
                x = 10;
                printf("%d*n", --x);  /* Should print 9 */
                printf("%d*n", x);    /* Should print 9 */
            }`,
            wantStdout: "9\n9\n",
        },
        {
            name: "nested_unary",
            code: `main() {
                auto x;
                x = 5;
                printf("%d*n", !!x);   /* Double negation */
                printf("%d*n", -(-x)); /* Double negation */
            }`,
            wantStdout: "1\n5\n",
        },
        {
            name: "unary_on_expressions",
            code: `main() {
                auto x, y;
                x = 3; y = 4;
                printf("%d*n", -(x + y));  /* Negate sum */
            }`,
            wantStdout: "-7\n",
        },
    }
    // Run tests...
}
```

**Coverage Impact:** +2-3%
**Untested Paths:**
- Prefix increment (`++x`)
- Edge cases in prefix decrement
- Nested unary operators

---

#### 2. **LoadValue/StoreValue Tests** (codegen.go 0% → 100%)

These functions are currently untested but might be **dead code**. Let me verify:

```go
func TestDirectLoadStore(t *testing.T) {
    // Test if LoadValue/StoreValue are used anywhere
    // If they're not used in current implementation, they should be removed
    // or documented as part of the public API

    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "simple_load_store",
            code: `main() {
                auto x;
                x = 42;
                printf("%d*n", x);
            }`,
            wantStdout: "42\n",
        },
    }
}
```

**Note:** These functions appear to be **alternative APIs** that are not used by the current parser. Consider:
1. **Remove them** if they're truly unused (dead code elimination)
2. **Document them** as part of the public API if they're intended for future use
3. **Test them directly** via unit tests (not integration tests)

**Coverage Impact:** +1% (if tested) or better code quality (if removed)

---

#### 3. **Error Handling in parseReturn** (parser.go 65% → 90%)

```go
func TestReturnErrors(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr string
    }{
        {
            name: "return_missing_semicolon",
            code: `main() { return }`,
            wantErr: "expect ';' after 'return' statement",
        },
        {
            name: "return_missing_close_paren",
            code: `main() { return(42 }`,
            wantErr: "expect ')' after 'return' statement",
        },
        {
            name: "return_invalid_expression",
            code: `main() { return(++); }`,
            wantErr: "expect expression",
        },
    }
    // Test error cases...
}

func TestReturnSyntax(t *testing.T) {
    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "return_without_parens",
            code: `main() { return; }`,
            wantStdout: "",  // Returns 0
        },
        {
            name: "return_with_parens",
            code: `main() { return(42); }`,
            wantStdout: "",  // Exit code 42
        },
    }
}
```

**Coverage Impact:** +1-2%

---

### P1: High Value Coverage Improvements (Target: +3-5%)

#### 4. **Array Initialization Error Cases** (parser.go:parseVector 65.3% → 85%)

```go
func TestVectorErrors(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr string
    }{
        {
            name: "vector_missing_size",
            code: `vec[];`,
            wantErr: "expect ']' or number",
        },
        {
            name: "vector_invalid_initializer",
            code: `vec[3] 1, 2;`,  // Missing third value
            wantErr: "expect initializer or ','",
        },
        {
            name: "vector_too_many_initializers",
            code: `vec[2] 1, 2, 3, 4;`,
            wantErr: "too many initializers",
        },
        {
            name: "vector_missing_semicolon",
            code: `vec[3] 1, 2, 3`,
            wantErr: "expect ';'",
        },
    }
}
```

**Coverage Impact:** +1-2%

---

#### 5. **Switch/Case Edge Cases** (control.go:parseCase 67.9% → 90%)

```go
func TestSwitchCaseEdgeCases(t *testing.T) {
    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "empty_case",
            code: `main() {
                auto x;
                x = 1;
                switch x {
                    case 1:
                    case 2:
                        printf("1 or 2*n");
                }
            }`,
            wantStdout: "1 or 2\n",
        },
        {
            name: "case_with_no_default",
            code: `main() {
                auto x;
                x = 99;
                switch x {
                    case 1: printf("one*n");
                    case 2: printf("two*n");
                }
                printf("done*n");
            }`,
            wantStdout: "done\n",  // Falls through to end
        },
        {
            name: "case_expression",
            code: `main() {
                auto x;
                x = 2;
                switch x + 1 {
                    case 3: printf("three*n");
                }
            }`,
            wantStdout: "three\n",
        },
        {
            name: "multiple_cases_same_value",
            code: `main() {
                auto x;
                x = 5;
                switch x {
                    case 5: printf("five*n");
                    case 5: printf("also five*n");
                }
            }`,
            wantStdout: "five\nalso five\n",
        },
    }
}
```

**Coverage Impact:** +1%

---

#### 6. **Initializer Constant Edge Cases** (parser.go:parseIvalConst 67.7% → 90%)

```go
func TestIvalConstEdgeCases(t *testing.T) {
    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "ival_identifier_reference",
            code: `
                x 100;
                y x;  /* Reference to x - currently returns 0 */
                main() {
                    printf("%d*n", y);
                }
            `,
            wantStdout: "0\n",  // TODO: Should be 100 when references implemented
        },
        {
            name: "ival_negative_char",
            code: `arr[3] -'A', 'B', -1;
                main() {
                    printf("%d*n", arr[0]);
                    printf("%d*n", arr[1]);
                    printf("%d*n", arr[2]);
                }
            `,
            wantStdout: "-65\n66\n-1\n",
        },
        {
            name: "ival_string_in_array",
            code: `arr[2] "hello", "world";
                main() {
                    printf("%s*n", arr[0]);
                }
            `,
            wantStdout: "hello\n",
        },
    }
}
```

**Coverage Impact:** +1%

---

#### 7. **Lexer Helper Functions** (lexer.go:ExpectChar 50% → 100%)

```go
func TestLexerHelpers(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr string
    }{
        {
            name: "expect_char_eof",
            code: `main() { auto x`,  // Missing semicolon and closing brace
            wantErr: "expect ';'",
        },
        {
            name: "expect_char_wrong",
            code: `main() { auto x, }`,  // Comma instead of semicolon
            wantErr: "expect ';' or identifier",
        },
    }
}

func TestLexerPeekAndEOF(t *testing.T) {
    // These might be dead code - verify usage
    // If unused, consider removing them
}
```

**Coverage Impact:** +0.5%

---

### P2: Lower Priority (Target: +1-2%)

#### 8. **Block Creation Edge Cases** (codegen.go:NewBlock 60% → 90%)

```go
func TestNestedBlocks(t *testing.T) {
    tests := []struct {
        name       string
        code       string
        wantStdout string
    }{
        {
            name: "deeply_nested_if",
            code: `main() {
                auto x;
                x = 5;
                if (x > 0)
                    if (x > 3)
                        if (x > 4)
                            if (x == 5)
                                printf("deeply nested*n");
            }`,
            wantStdout: "deeply nested\n",
        },
        {
            name: "many_sequential_blocks",
            code: `main() {
                auto i;
                i = 0;
                while (i < 100) {
                    i++;
                }
                printf("%d*n", i);
            }`,
            wantStdout: "100\n",
        },
    }
}
```

**Coverage Impact:** +0.5%

---

#### 9. **Top-Level Declaration Error Handling** (parser.go:ParseDeclarations 68% → 85%)

```go
func TestDeclarationErrors(t *testing.T) {
    tests := []struct {
        name    string
        code    string
        wantErr string
    }{
        {
            name: "unexpected_token_at_top_level",
            code: `123`,
            wantErr: "unexpected character",
        },
        {
            name: "unclosed_function",
            code: `main() { auto x;`,
            wantErr: "expect '}'",
        },
        {
            name: "invalid_identifier",
            code: `123abc;`,
            wantErr: "expect identifier",
        },
    }
}
```

**Coverage Impact:** +0.5%

---

## 📈 Expected Coverage Improvement

| Priority | Tests | Expected Coverage Gain | Cumulative |
|----------|-------|----------------------|------------|
| P0 | 4 test suites | +5-7% | 81.6-83.6% |
| P1 | 5 test suites | +3-5% | 84.6-88.6% |
| P2 | 2 test suites | +1-2% | 85.6-90.6% |

**Total new tests:** ~35-45 individual test cases
**Estimated implementation time:** 4-6 hours
**Target final coverage:** 85-90%

---

## 🚨 Dead Code Candidates

These functions may be **unused** and could be removed for better code quality:

1. **`LoadValue` / `StoreValue`** (codegen.go) - Not called by parser
2. **`PeekChar`** (lexer.go) - May be unused
3. **`IsEOF`** (lexer.go) - May be unused
4. **`parseKeywordOrExpression`** (parser.go) - Wrapper for `parseKeywordOrExpressionWithSwitch`

**Recommendation:**
- Run dead code analysis: `go run honnef.co/go/tools/cmd/staticcheck@latest ./...`
- Remove unused functions or add documentation explaining their purpose
- If they're part of a public API, add unit tests

---

## 📝 Implementation Strategy

### Phase 1: Quick Wins (Week 1)
1. Add unary operator tests (prefix ++/--)
2. Add return statement error tests
3. Add switch/case edge cases
4. **Expected:** +3-4% coverage

### Phase 2: Error Paths (Week 2)
1. Add vector/array initialization errors
2. Add ival constant edge cases
3. Add lexer helper tests
4. **Expected:** +2-3% coverage

### Phase 3: Cleanup (Week 3)
1. Identify and remove dead code
2. Add block nesting tests
3. Add declaration error tests
4. **Expected:** +1-2% coverage

### Phase 4: Final Push (Optional)
1. Add fuzzing tests for parser
2. Add property-based tests
3. Add performance benchmarks
4. **Expected:** +2-3% coverage (85-90% total)

---

## 🎯 Success Metrics

- **Coverage:** 76.6% → 85%+ (minimum target)
- **Test count:** 144 → 180+ tests
- **Code quality:** Remove 50+ lines of dead code
- **Robustness:** Cover all error paths in parser

---

**Last Updated:** October 13, 2025
**Author:** Coverage Analysis Tool
**Status:** Ready for implementation
