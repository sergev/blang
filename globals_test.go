package main

import (
	"testing"
)

// TestGlobals tests global and local variable features (from oldtests/globals_test.cpp)
func TestGlobals(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "global_scalars",
			code: `a;
			b 123;
			c -345, 'foo', "bar";

			main() {
				extrn a, b, c;

				printf("a = %d*n", a);
				printf("b = %d*n", b);
				a = &c;
				printf("c = %d, '%c', *"%s*"*n", c, a[1], a[2]);
			}`,
			wantStdout: `a = 0
b = 123
c = -345, 'foo', "bar"
`,
		},
		{
			name: "global_vectors",
			code: `a[];
			b[] 123;
			c[4] -345, 'foo', "bar";

			main() {
				extrn a, b, c;

				printf("a = %d*n", a[1]);
				printf("b = %d*n", b[0]);
				printf("c = %d, '%c', *"%s*", %d*n", c[0], c[1], c[2], c[3]);
			}`,
			wantStdout: `a = 123
b = 123
c = -345, 'foo', "bar", 0
`,
		},
		{
			name: "local_scalars",
			code: `main() {
				auto a;
				auto b;
				auto c;

				printf("offset a = %d*n", (&a) - &c);
				printf("offset b = %d*n", (&b) - &c);
				printf("offset c = %d*n", (&c) - &c);
			}`,
			wantStdout: `offset a = 16
offset b = 8
offset c = 0
`,
		},
		{
			name: "local_vectors",
			code: `main() {
				auto a[124];
				auto b['x'];
				auto c[1];

				printf("offset a = %d*n", (&a) - &c);
				printf("offset b = %d*n", (&b) - &c);
				printf("offset c = %d*n", (&c) - &c);
			}`,
			wantStdout: `offset a = 984
offset b = 16
offset c = 0
`,
		},
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
