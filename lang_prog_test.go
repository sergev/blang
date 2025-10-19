package main

import (
	"bytes"
	"os/exec"
	"testing"
)

// Consolidated program-structure tests: functions, control flow, globals, runtime

// TestFunctions tests various function features (from oldtests/func_test.cpp)
func TestFunctions(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{name: "function_definitions", code: `a() {}
            b();
            c() label:;
            d() label: goto label;
            e() return;
            f(x) return(x);
            g(x) x;
            h(x) if(x) 123;
            i(x) if(x) 123; else 456;
            j(x) while(x);
            k(x) switch(x);
            l(x) switch(x) case 1:;
            m() extrn x;
            n() auto x;

            main() {
                printf("before a()*n");
                a();    printf("after a(), before b()*n");
                b();    printf("after b(), before c()*n");
                c();    printf("after c(), before e()*n");
                e();    printf("after e(), before f()*n");
                f(42);  printf("after f(), before g()*n");
                g(42);  printf("after g(), before h()*n");
                h(42);  printf("after h(), before i()*n");
                i(42);  printf("after i(), before j()*n");
                j(0);   printf("after j(), before k()*n");
                k(42);  printf("after k(), before l()*n");
                l(42);  printf("after l(), before m()*n");
                m();    printf("after m(), before n()*n");
                n();    printf("after n()*n");
            }`, wantStdout: `before a()
after a(), before b()
after b(), before c()
after c(), before e()
after e(), before f()
after f(), before g()
after g(), before h()
after h(), before i()
after i(), before j()
after j(), before k()
after k(), before l()
after l(), before m()
after m(), before n()
after n()
`},
		{name: "function_arguments", code: `func(a, b, c)
            {
                printf("a = %d, b = '%c', c = *"%s*"*n", a, b, c);
            }

            main() {
                func(123, 'foo', "bar");
            }`, wantStdout: `a = 123, b = 'foo', c = "bar"
`},
		{name: "function_ternary_operator", code: `choose(a, b, c)
            {
                return (a ? b : c);
            }

            main() {
                printf("%d*n", choose(1, 123, 456));
                printf("%d*n", choose(0, 123, 456));
            }`, wantStdout: `123
456
`},
		{name: "nested_ternary", code: `classify(n) {
                return (n > 50 ? 100 : (n > 25 ? 50 : 25));
            }

            sign(n) {
                return (n > 0 ? 1 : (n < 0 ? -1 : 0));
            }

            main() {
                printf("classify(75) = %d*n", classify(75));
                printf("classify(40) = %d*n", classify(40));
                printf("classify(10) = %d*n", classify(10));

                printf("sign(42) = %d*n", sign(42));
                printf("sign(-17) = %d*n", sign(-17));
                printf("sign(0) = %d*n", sign(0));
            }`, wantStdout: `classify(75) = 100
classify(40) = 50
classify(10) = 25
sign(42) = 1
sign(-17) = -1
sign(0) = 0
`},
		{name: "ternary_in_expression", code: `main() {
                auto x, y, z, result;

                x = 10;
                y = 20;
                z = 30;

                result = x + (y > 15 ? z : 0);
                printf("x + (y > 15 ? z : 0) = %d*n", result);

                result = (x < y ? x : y) * 2;
                printf("(x < y ? x : y) ** 2 = %d*n", result);

                printf("nested: %d*n", (x > 5 ? (y > 15 ? 100 : 50) : 0));
            }`, wantStdout: `x + (y > 15 ? z : 0) = 40
(x < y ? x : y) * 2 = 20
nested: 100
`},
		{name: "three_args", code: `
main() {
	abc(12, 34, 56);
}
abc(a, b, c) {
	printf("%d %d %d*n", a, b, c);
}
`,				wantStdout: "12 34 56\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout := compileLinkRunFromCode(t, tt.name, tt.code)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestIndirectCalls tests indirect function calls through function pointers
func TestIndirectCalls(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{name: "indirect_call_basic", code: `add(a, b) {
                return(a + b);
            }

            func_ptr;

            main() {
                extrn func_ptr;

                func_ptr = add;
                printf("Result: %d*n", func_ptr(3, 5));
            }`, wantStdout: "Result: 8\n"},
		{name: "indirect_call_multiple", code: `add(a, b) { return(a + b); }
            sub(a, b) { return(a - b); }
            mul(a, b) { return(a * b); }

            op;

            main() {
                extrn op;

                op = add;
                printf("add: %d*n", op(10, 5));

                op = sub;
                printf("sub: %d*n", op(10, 5));

                op = mul;
                printf("mul: %d*n", op(10, 5));
            }`, wantStdout: `add: 15
sub: 5
mul: 50
`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout := compileLinkRunFromCode(t, tt.name, tt.code)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestNestedLoops tests nested while loops with unique labels
func TestNestedLoops(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{name: "nested_while_basic", code: `main() {
                auto i, j, sum;

                sum = 0;
                i = 0;
                while (i < 3) {
                    j = 0;
                    while (j < 3) {
                        sum = sum + 1;
                        j++;
                    }
                    i++;
                }

                printf("sum = %d*n", sum);
            }`, wantStdout: "sum = 9\n"},
		{name: "nested_while_complex", code: `main() {
                auto i, j, k, count;

                count = 0;
                i = 1;
                while (i <= 2) {
                    j = 1;
                    while (j <= 2) {
                        k = 1;
                        while (k <= 2) {
                            count++;
                            k++;
                        }
                        j++;
                    }
                    i++;
                }

                printf("count = %d*n", count);
            }`, wantStdout: "count = 8\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout := compileLinkRunFromCode(t, tt.name, tt.code)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestGlobals tests global and local variable features (from oldtests/globals_test.cpp)
func TestGlobals(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{name: "global_scalars", code: `a;
            b 123;
            c -345, 'foo', "bar";

            main() {
                extrn a, b, c;

                printf("a = %d*n", a);
                printf("b = %d*n", b);
                a = &c;
                printf("c = %d, '%c', *"%s*"*n", c, a[1], a[2]);
            }`, wantStdout: `a = 0
b = 123
c = -345, 'foo', "bar"
`},
		{name: "global_vectors", code: `a[];
            b[] 123;
            c[4] -345, 'foo', "bar";

            main() {
                extrn a, b, c;

                printf("a = %d*n", a[1]);
                printf("b = %d*n", b[0]);
                printf("c = %d, '%c', *"%s*", %d*n", c[0], c[1], c[2], c[3]);
            }`, wantStdout: `a = 123
b = 123
c = -345, 'foo', "bar", 0
`},
		{name: "local_scalars", code: `main() {
                auto a;
                auto b;
                auto c;

                printf("offset a = %d*n", (&a) - &c);
                printf("offset b = %d*n", (&b) - &c);
                printf("offset c = %d*n", (&c) - &c);
            }`, wantStdout: `offset a = 16
offset b = 8
offset c = 0
`},
		{name: "local_vectors", code: `main() {
                auto a[124];
                auto b['x'];
                auto c[1];

                printf("offset a = %d*n", (&a) - &c);
                printf("offset b = %d*n", (&b) - &c);
                printf("offset c = %d*n", (&c) - &c);
            }`, wantStdout: `offset a = 984
offset b = 16
offset c = 0
`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout := compileLinkRunFromCode(t, tt.name, tt.code)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestLibbFunctions tests runtime library functions
func TestLibbFunctions(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "runtime_write",
			code: `main() {
                write('Hello,');
                write(' World');
                write('!*n');
            }`,
			wantStdout: "Hello, World!\n",
		},
		{
			name: "runtime_printf",
			code: `main() {
                printf("Hello, World!*n");
                printf("%% %% %%%%*n");
                printf("format %%d: %d %d*n", 123, -123);
                printf("format %%o: %o %o*n", 234, -234);
                printf("format %%c: %c %c*n", 'foo', 'bar');
                printf("format %%s: *"%s*" *"%s*"*n", "Hello", "World");
                printf("unknown format: %q*n", "foo");
            }`,
			wantStdout: `Hello, World!
% % %%
format %d: 123 -123
format %o: 352 -352
format %c: foo bar
format %s: "Hello" "World"
unknown format: %q
`,
		},
		{
			name: "runtime_exit",
			code: `main() {
                printf("before exit()*n");
                exit();
                printf("after exit()*n");
            }`,
			wantStdout: "before exit()\n",
		},
		{
			name: "runtime_char",
			code: `main() {
                write(char("fubar", 2));
                write(char("fubar", 4));
                write(char("fubar", 1));
                write(char("fubar", 0));
                write(char("fubar", 3));
                write('*n');
            }`,
			wantStdout: "brufa\n",
		},
		{
			name: "runtime_lchar",
			code: `main() {
                auto str;

                lchar(&str, 0, 'f');
                lchar(&str, 1, 'u');
                lchar(&str, 2, 'b');
                lchar(&str, 3, 'a');
                lchar(&str, 4, 'r');
                lchar(&str, 5, 0);
                printf("%s*n", &str);
            }`,
			wantStdout: "fubar\n",
		},
		{
			name: "runtime_nwrite",
			code: `main() {
                nwrite(1, "foobar*n", 7);
            }`,
			wantStdout: "foobar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStdout := compileLinkRunFromCode(t, tt.name, tt.code)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout = %q, want %q", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestRuntimeRead tests the runtime read() routine, including EOF behavior (octal 4).
func TestRuntimeRead(t *testing.T) {
	ensureLibbOrSkip(t)

	const bcode = `
main() {
    auto c;
    while ((c = read()) != 4) {
        printf("%d ", c);
    }
    printf("EOF=%d*n", c);
}
`

	runReadProg := func(in []byte) string {
		dir, bFile, llFile, exeFile := createTempBFile(t, "read_prog", bcode)
		_ = dir
		compileToLL(t, bFile, llFile)
		linkWithClang(t, llFile, exeFile)
		cmd := exec.Command(exeFile)
		cmd.Stdin = bytes.NewReader(in)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		return string(out)
	}

	t.Run("ASCIIAndEOF", func(t *testing.T) {
		got := runReadProg([]byte("ABC"))
		want := "65 66 67 EOF=4\n"
		if got != want {
			t.Fatalf("unexpected output: got %q, want %q", got, want)
		}
	})

	t.Run("ImmediateEOF", func(t *testing.T) {
		got := runReadProg(nil)
		want := "EOF=4\n"
		if got != want {
			t.Fatalf("unexpected output: got %q, want %q", got, want)
		}
	})

	t.Run("EmbeddedEOFChar", func(t *testing.T) {
		got := runReadProg([]byte{'A', 0x04, 'B'})
		want := "65 EOF=4\n"
		if got != want {
			t.Fatalf("unexpected output: got %q, want %q", got, want)
		}
	})
}
