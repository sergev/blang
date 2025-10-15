package main

import (
	"testing"
)

// TestLibbFunctions tests runtime library functions (from oldtests/libb_test.cpp)
func TestLibbFunctions(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "libb_write",
			code: `main() {
				write('Hello,');
				write(' World');
				write('!*n');
			}`,
			wantStdout: "Hello, World!\n",
		},
		{
			name: "libb_printf",
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
			name: "libb_exit",
			code: `main() {
				printf("before exit()*n");
				exit();
				printf("after exit()*n");
			}`,
			wantStdout: "before exit()\n",
		},
		{
			name: "libb_char",
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
			name: "libb_lchar",
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
			name: "libb_nwrite",
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
