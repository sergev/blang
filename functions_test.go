package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestFunctions tests various function features (from oldtests/func_test.cpp)
func TestFunctions(t *testing.T) {
	// Check if clang is available
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "function_definitions",
			code: `a() {}
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
			}`,
			wantStdout: `before a()
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
`,
		},
		{
			name: "function_arguments",
			code: `func(a, b, c)
			{
				printf("a = %d, b = '%c', c = *"%s*"*n", a, b, c);
			}

			main() {
				func(123, 'foo', "bar");
			}`,
			wantStdout: `a = 123, b = 'foo', c = "bar"
`,
		},
		{
			name: "function_ternary_operator",
			code: `choose(a, b, c)
			{
				return (a ? b : c);
			}

			main() {
				printf("%d*n", choose(1, 123, 456));
				printf("%d*n", choose(0, 123, 456));
			}`,
			wantStdout: `123
456
`,
		},
		{
			name: "nested_ternary",
			code: `classify(n) {
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
			}`,
			wantStdout: `classify(75) = 100
classify(40) = 50
classify(10) = 25
sign(42) = 1
sign(-17) = -1
sign(0) = 0
`,
		},
		{
			name: "ternary_in_expression",
			code: `main() {
				auto x, y, z, result;

				x = 10;
				y = 20;
				z = 30;

				result = x + (y > 15 ? z : 0);
				printf("x + (y > 15 ? z : 0) = %d*n", result);

				result = (x < y ? x : y) * 2;
				printf("(x < y ? x : y) ** 2 = %d*n", result);

				printf("nested: %d*n", (x > 5 ? (y > 15 ? 100 : 50) : 0));
			}`,
			wantStdout: `x + (y > 15 ? z : 0) = 40
(x < y ? x : y) * 2 = 20
nested: 100
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			llFile := filepath.Join(tmpDir, "test.ll")
			exeFile := filepath.Join(tmpDir, "test")

			// Write test code to file
			err := os.WriteFile(inputFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{inputFile})
			args.OutputFile = llFile
			args.OutputType = OutputIR

			err = Compile(args)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			// Step 2: Link with libb.o using clang
			linkCmd := exec.Command("clang", llFile, "libb.o", "-o", exeFile)
			linkOutput, err := linkCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Linking failed: %v\nOutput: %s", err, linkOutput)
			}

			// Step 3: Run the executable
			runCmd := exec.Command(exeFile)
			stdout, err := runCmd.Output()
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("Failed to run executable: %v", err)
				}
			}

			// Check stdout
			gotStdout := string(stdout)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}

// TestIndirectCalls tests indirect function calls through function pointers
func TestIndirectCalls(t *testing.T) {
	// Check if clang is available
	if _, err := os.Stat("libb.o"); err != nil {
		t.Skip("libb.o not found, run 'make' first")
	}

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "indirect_call_basic",
			code: `add(a, b) {
				return(a + b);
			}

			func_ptr;

			main() {
				extrn func_ptr;

				func_ptr = add;
				printf("Result: %d*n", func_ptr(3, 5));
			}`,
			wantStdout: "Result: 8\n",
		},
		{
			name: "indirect_call_multiple",
			code: `add(a, b) { return(a + b); }
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
			}`,
			wantStdout: `add: 15
sub: 5
mul: 50
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "test.b")
			llFile := filepath.Join(tmpDir, "test.ll")
			exeFile := filepath.Join(tmpDir, "test")

			// Write test code to file
			err := os.WriteFile(inputFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{inputFile})
			args.OutputFile = llFile
			args.OutputType = OutputIR

			err = Compile(args)
			if err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			// Step 2: Link with libb.o using clang
			linkCmd := exec.Command("clang", llFile, "libb.o", "-o", exeFile)
			linkOutput, err := linkCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Linking failed: %v\nOutput: %s", err, linkOutput)
			}

			// Step 3: Run the executable
			runCmd := exec.Command(exeFile)
			stdout, err := runCmd.Output()
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("Failed to run executable: %v", err)
				}
			}

			// Check stdout
			gotStdout := string(stdout)
			if gotStdout != tt.wantStdout {
				t.Errorf("Stdout mismatch:\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}
