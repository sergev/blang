package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestPrecedence tests operator precedence (from oldtests/precedence_test.cpp)
func TestPrecedence(t *testing.T) {
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
			name: "add_mul",
			code: `main() {
				printf("3 + 4 ** 2 -> %d*n", 3 + 4 * 2);
			}`,
			wantStdout: "3 + 4 * 2 -> 11\n",
		},
		{
			name: "mul_add_mul",
			code: `main() {
				printf("5 ** 2 + 3 ** 4 -> %d*n", 5 * 2 + 3 * 4);
			}`,
			wantStdout: "5 * 2 + 3 * 4 -> 22\n",
		},
		{
			name: "sub_div",
			code: `main() {
				printf("10 - 6 / 2 -> %d*n", 10 - 6 / 2);
			}`,
			wantStdout: "10 - 6 / 2 -> 7\n",
		},
		{
			name: "mod_add",
			code: `main() {
				printf("7 %% 3 + 2 -> %d*n", 7 % 3 + 2);
			}`,
			wantStdout: "7 % 3 + 2 -> 3\n",
		},
		{
			name: "add_lt",
			code: `main() {
				printf("5 + 3 < 9 -> %d*n", 5 + 3 < 9);
			}`,
			wantStdout: "5 + 3 < 9 -> 1\n",
		},
		{
			name: "lt_eq",
			code: `main() {
				printf("4 < 6 == 1 -> %d*n", 4 < 6 == 1);
			}`,
			wantStdout: "4 < 6 == 1 -> 1\n",
		},
		{
			name: "eq_and",
			code: `main() {
				printf("3 == 3 & 1 -> %d*n", 3 == 3 & 1);
			}`,
			wantStdout: "3 == 3 & 1 -> 1\n",
		},
		{
			name: "and_or",
			code: `main() {
				printf("2 & 3 | 4 -> %d*n", 2 & 3 | 4);
			}`,
			wantStdout: "2 & 3 | 4 -> 6\n",
		},
		{
			name: "mul_add_lt",
			code: `main() {
				printf("2 ** 3 + 4 < 11 -> %d*n", 2 * 3 + 4 < 11);
			}`,
			wantStdout: "2 * 3 + 4 < 11 -> 1\n",
		},
		{
			name: "mul_ge_eq",
			code: `main() {
				printf("5 ** 2 >= 10 == 1 -> %d*n", 5 * 2 >= 10 == 1);
			}`,
			wantStdout: "5 * 2 >= 10 == 1 -> 1\n",
		},
		{
			name: "mul_and_add",
			code: `main() {
				printf("4 ** 2 & 3 + 1 -> %d*n", 4 * 2 & 3 + 1);
			}`,
			wantStdout: "4 * 2 & 3 + 1 -> 0\n",
		},
		{
			name: "div_add_gt_or",
			code: `main() {
				printf("6 / 2 + 1 > 3 | 2 -> %d*n", 6 / 2 + 1 > 3 | 2);
			}`,
			wantStdout: "6 / 2 + 1 > 3 | 2 -> 3\n",
		},
		{
			name: "div_mod",
			code: `main() {
				printf("10 / 2 %% 3 -> %d*n", 10 / 2 % 3);
			}`,
			wantStdout: "10 / 2 % 3 -> 2\n",
		},
		{
			name: "mul_or",
			code: `main() {
				printf("0 ** 5 | 3 -> %d*n", 0 * 5 | 3);
			}`,
			wantStdout: "0 * 5 | 3 -> 3\n",
		},
		{
			name: "mul_lshift",
			code: `main() {
				printf("4 ** 3 << 2 -> %d*n", 4 * 3 << 2);
			}`,
			wantStdout: "4 * 3 << 2 -> 48\n",
		},
		{
			name: "lshift_lt",
			code: `main() {
				printf("1 << 2 < 5 -> %d*n", 1 << 2 < 5);
			}`,
			wantStdout: "1 << 2 < 5 -> 1\n",
		},
		{
			name: "sub_rshift",
			code: `main() {
				printf("16 - 8 >> 1 -> %d*n", 16 - 8 >> 1);
			}`,
			wantStdout: "16 - 8 >> 1 -> 4\n",
		},
		{
			name: "lshift_and",
			code: `main() {
				printf("3 << 2 & 7 -> %d*n", 3 << 2 & 7);
			}`,
			wantStdout: "3 << 2 & 7 -> 4\n",
		},
		{
			name: "or_rshift",
			code: `main() {
				printf("2 | 4 >> 1 -> %d*n", 2 | 4 >> 1);
			}`,
			wantStdout: "2 | 4 >> 1 -> 2\n",
		},
		{
			name: "rshift_eq",
			code: `main() {
				printf("8 >> 2 == 2 -> %d*n", 8 >> 2 == 2);
			}`,
			wantStdout: "8 >> 2 == 2 -> 1\n",
		},
		{
			name: "mul_lshift_add",
			code: `main() {
				printf("5 ** 2 << 1 + 3 -> %d*n", 5 * 2 << 1 + 3);
			}`,
			wantStdout: "5 * 2 << 1 + 3 -> 160\n",
		},
		{
			name: "mod_lshift",
			code: `main() {
				printf("15 %% 4 << 2 -> %d*n", 15 % 4 << 2);
			}`,
			wantStdout: "15 % 4 << 2 -> 12\n",
		},
		{
			name: "lshift_gt_and",
			code: `main() {
				printf("1 << 3 > 5 & 2 -> %d*n", 1 << 3 > 5 & 2);
			}`,
			wantStdout: "1 << 3 > 5 & 2 -> 0\n",
		},
		{
			name: "add_lshift",
			code: `main() {
				printf("12345 + 10 << 4 -> %d*n", 12345 + 10 << 4);
			}`,
			wantStdout: "12345 + 10 << 4 -> 197680\n",
		},
		{
			name: "div_rshift",
			code: `main() {
				printf("16 / 2 >> 1 -> %d*n", 16 / 2 >> 1);
			}`,
			wantStdout: "16 / 2 >> 1 -> 4\n",
		},
		{
			name: "and_lshift_or",
			code: `main() {
				printf("7 & 3 << 2 | 8 -> %d*n", 7 & 3 << 2 | 8);
			}`,
			wantStdout: "7 & 3 << 2 | 8 -> 12\n",
		},
		{
			name: "lshift_ne",
			code: `main() {
				printf("1 << 4 != 15 -> %d*n", 1 << 4 != 15);
			}`,
			wantStdout: "1 << 4 != 15 -> 1\n",
		},
		{
			name: "rshift_ge",
			code: `main() {
				printf("98765 >> 3 >= 12345 -> %d*n", 98765 >> 3 >= 12345);
			}`,
			wantStdout: "98765 >> 3 >= 12345 -> 1\n",
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

// TestExpressions tests various expression features (from oldtests/expr_test.cpp)
func TestExpressions(t *testing.T) {
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
			name: "unary_operators",
			code: `main() {
				extrn x;
				auto y;

				printf("global -x = %d, expect %d*n", -x, -42);
				x = 0;
				printf("global !x = %d, expect %d*n", !x, 1);
				y = 987;
				x = &y;
				printf("global **x = %d, expect %d*n", *x, y);

				printf("local -y = %d, expect %d*n", -y, -987);
				y = 0;
				printf("local !y = %d, expect %d*n", !y, 1);
				x = 42;
				y = &x;
				printf("local **y = %d, expect %d*n", *y, x);
			}

			x 42;
			`,
			wantStdout: `global -x = -42, expect -42
global !x = 1, expect 1
global *x = 987, expect 987
local -y = -987, expect -987
local !y = 1, expect 1
local *y = 42, expect 42
`,
		},
		{
			name: "negation_in_conditional_context",
			code: `main() {
				extrn x;
				auto y;

				y = x + 100;
				printf("x = %d, y = %d*n", x, y);
				if (x)
					printf("if (x) WRONG*n");
				else
					printf("if (x) Correct*n");

				if (y)
					printf("if (y) Correct*n");
				else
					printf("if (y) WRONG*n");

				if (!x)
					printf("if (!x) Correct*n");
				else
					printf("if (!x) WRONG*n");

				if (!y)
					printf("if (!y) WRONG*n");
				else
					printf("if (!y) Correct*n");

				while (!x) {
					printf("while (!x) x = %d*n", x);
					x = 42;
				}
			}

			x;
			`,
			wantStdout: `x = 0, y = 100
if (x) Correct
if (y) Correct
if (!x) Correct
if (!y) Correct
while (!x) x = 0
`,
		},
		{
			name: "postfix_operators",
			code: `incr(x) {
				printf("increment %d*n", x++);
				return (x);
			}

			add(a, b) {
				printf("add %d + %d*n", a, b);
				return (a + b);
			}

			decr(x) {
				printf("decrement %d*n", x--);
				return (x);
			}

			sub(a, b) {
				printf("subtract %d - %d*n", a, b);
				return (a - b);
			}

			assign_local(x) {
				auto result;
				printf("assign local %d*n", x);
				result = x;
				return (result);
			}

			assign_global(x) {
				extrn g;

				printf("assign global %d*n", x);
				g = x;
			}

			main() {
				extrn g;

				printf("%d*n", incr(42));
				printf("%d*n", add(42, 123));
				printf("%d*n", decr(42));
				printf("%d*n", sub(42, 123));
				printf("%d*n", assign_local(42));
				assign_global(42);
				printf("%d*n", g);
			}

			g;
			`,
			wantStdout: `increment 42
43
add 42 + 123
165
decrement 42
41
subtract 42 - 123
-81
assign local 42
42
assign global 42
42
`,
		},
		{
			name: "local_array",
			code: `main() {
				auto l[3];

				l[0] = 123;
				l[1] = 'local';
				l[2] = "string";
				printf("local = %d, '%c', *"%s*"*n", l[0], l[1], l[2]);
			}`,
			wantStdout: `local = 123, 'local', "string"
`,
		},
		{
			name: "global_array",
			code: `g[3] -345, 'foo', "bar";

			main() {
				extrn g;

				printf("global = %d, '%c', *"%s*"*n", g[0], g[1], g[2]);
				printf("address = %d, %d, %d*n", (&g[0]) - g, (&g[1]) - g, (&g[2]) - g);
			}`,
			wantStdout: `global = -345, 'foo', "bar"
address = 0, 8, 16
`,
		},
		{
			name: "local_mix",
			code: `main() {
				auto e, d;
				auto c[1];
				auto b, a;
				auto p;

				a = 11;
				b = 22;
				c[0] = 33;
				d = 44;
				e = 55;

				printf("%d %d %d %d", a, b, c - &c, c[0]);
				printf(" %d %d*n", d, e);
				p = &a;
				printf("%d %d %d %d", p[0], p[1], p[2] - &c, p[3]);
				printf(" %d %d*n", p[4], p[5]);
			}`,
			wantStdout: `11 22 8 33 44 55
11 22 8 33 44 55
`,
		},
		{
			name: "binary_operators",
			code: `x 42;

			main() {
				extrn x;
				auto y;

				y = 345;
				printf("%d + %d -> %d*n", x, y, x + y);
				printf("%d + %d -> %d*n", y, x, y + x);

				printf("%d - %d -> %d*n", x, y, x - y);
				printf("%d - %d -> %d*n", y, x, y - x);

				printf("%d ** %d -> %d*n", x, y, x * y);
				printf("%d ** %d -> %d*n", y, x, y * x);

				printf("%d / %d -> %d*n", x, y, x / y);
				printf("%d / %d -> %d*n", y, x, y / x);

				printf("%d %% %d -> %d*n", x, y, x % y);
				printf("%d %% %d -> %d*n", y, x, y % x);

				printf("%d < %d -> %d*n", x, y, x < y);
				printf("%d < %d -> %d*n", y, x, y < x);

				printf("%d <= %d -> %d*n", x, y, x <= y);
				printf("%d <= %d -> %d*n", y, x, y <= x);

				printf("%d > %d -> %d*n", x, y, x > y);
				printf("%d > %d -> %d*n", y, x, y > x);

				printf("%d >= %d -> %d*n", x, y, x >= y);
				printf("%d >= %d -> %d*n", y, x, y >= x);

				printf("%d == %d -> %d*n", x, y, x == y);
				printf("%d == %d -> %d*n", y, x, y == x);

				printf("%d != %d -> %d*n", x, y, x != y);
				printf("%d != %d -> %d*n", y, x, y != x);

				printf("%d & %d -> %d*n", x, y, x & y);
				printf("%d & %d -> %d*n", y, x, y & x);

				printf("%d | %d -> %d*n", x, y, x | y);
				printf("%d | %d -> %d*n", y, x, y | x);
			}`,
			wantStdout: `42 + 345 -> 387
345 + 42 -> 387
42 - 345 -> -303
345 - 42 -> 303
42 * 345 -> 14490
345 * 42 -> 14490
42 / 345 -> 0
345 / 42 -> 8
42 % 345 -> 42
345 % 42 -> 9
42 < 345 -> 1
345 < 42 -> 0
42 <= 345 -> 1
345 <= 42 -> 0
42 > 345 -> 0
345 > 42 -> 1
42 >= 345 -> 0
345 >= 42 -> 1
42 == 345 -> 0
345 == 42 -> 0
42 != 345 -> 1
345 != 42 -> 1
42 & 345 -> 8
345 & 42 -> 8
42 | 345 -> 379
345 | 42 -> 379
`,
		},
		{
			name: "eq_by_bitmask",
			code: `main() {
				auto cval;

				cval = 51;
				if ((cval & 017777) == cval) {
					printf("Small positive: %d*n", cval);
				} else {
					printf("Wrong: %d*n", cval);
				}
			}`,
			wantStdout: "Small positive: 51\n",
		},
		{
			name: "octal_literals",
			code: `main() {
				auto v;
				v = 012345;
				printf("%d*n", v);
				v = -04567;
				printf("%d*n", v);
			}`,
			wantStdout: `5349
-2423
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

// TestUnaryOperators tests comprehensive unary operator functionality
func TestUnaryOperators(t *testing.T) {
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
			name: "prefix_increment_global",
			code: `main() {
				extrn x;
				x = 15;
				printf("%d*n", ++x);  /* Should print 16 */
				printf("%d*n", x);    /* Should print 16 */
			}
			x 0;`,
			wantStdout: "16\n16\n",
		},
		{
			name: "prefix_decrement_global",
			code: `main() {
				extrn x;
				x = 20;
				printf("%d*n", --x);  /* Should print 19 */
				printf("%d*n", x);    /* Should print 19 */
			}
			x 0;`,
			wantStdout: "19\n19\n",
		},
		{
			name: "nested_unary_operators",
			code: `main() {
				auto x;
				x = 5;
				printf("%d*n", !!x);     /* Double negation: !(!5) = !0 = 1 */
				printf("%d*n", -(-x));   /* Double negation: -(-5) = 5 */
				++x;                     /* Increment x: 5 -> 6 */
				++x;                     /* Increment x: 6 -> 7 */
				printf("%d*n", x);       /* Final value: 7 */
			}`,
			wantStdout: "1\n5\n7\n",
		},
		{
			name: "unary_on_expressions",
			code: `main() {
				auto x, y, sum;
				x = 3; y = 4;
				printf("%d*n", -(x + y));  /* Negate sum: -(3+4) = -7 */
				printf("%d*n", !(x * y));  /* Logical not: !(3*4) = !12 = 0 */
				sum = x + y;               /* Store sum: 3+4 = 7 */
				++sum;                     /* Increment sum: 7 -> 8 */
				printf("%d*n", sum);       /* Final sum: 8 */
			}`,
			wantStdout: "-7\n0\n8\n",
		},
		{
			name: "unary_with_pointers",
			code: `main() {
				auto x;
				auto p;
				x = 100;
				p = &x;
				printf("%d*n", *p);       /* Dereference: *(&x) = 100 */
				++x;                      /* Increment original variable: ++100 = 101 */
				printf("%d*n", *p);       /* Dereferenced value: 101 */
				printf("%d*n", x);        /* Original variable: 101 */
			}`,
			wantStdout: "100\n101\n101\n",
		},
		{
			name: "unary_with_arrays",
			code: `main() {
				auto arr[3];
				auto i;
				arr[0] = 10;
				arr[1] = 20;
				i = 0;
				printf("%d*n", ++arr[i]);     /* Increment arr[0]: ++10 = 11 */
				printf("%d*n", arr[0]);       /* Check arr[0]: 11 */
				printf("%d*n", --arr[++i]);   /* Decrement arr[1]: --20 = 19 */
				printf("%d*n", arr[1]);       /* Check arr[1]: 19 */
			}`,
			wantStdout: "11\n11\n19\n19\n",
		},
		{
			name: "unary_edge_cases",
			code: `main() {
				auto x, y;
				x = 0;
				y = 1;
				printf("%d*n", ++x);  /* Increment 0: ++0 = 1 */
				printf("%d*n", --x);  /* Decrement 1: --1 = 0 */
				printf("%d*n", !x);   /* Logical not 0: !0 = 1 */
				printf("%d*n", !y);   /* Logical not 1: !1 = 0 */
				printf("%d*n", -x);   /* Negate 0: -0 = 0 */
				printf("%d*n", -y);   /* Negate 1: -1 = -1 */
			}`,
			wantStdout: "1\n0\n1\n0\n0\n-1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, tt.name+".ll")
			exeFile := filepath.Join(tmpDir, tt.name)

			// Create temporary B file
			bFile := filepath.Join(tmpDir, tt.name+".b")
			err := os.WriteFile(bFile, []byte(tt.code), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Step 1: Compile B program to LLVM IR
			args := NewCompileOptions("blang", []string{bFile})
			args.OutputFile = llFile

			err = Compile(args)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
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
				t.Errorf("Stdout mismatch.\nGot:\n%s\nWant:\n%s", gotStdout, tt.wantStdout)
			}
		})
	}
}
