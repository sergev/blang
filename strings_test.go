package main

import (
	"testing"
)

// TestStrings tests string and character literal features (from oldtests/string_test.cpp)
func TestStrings(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "string_literals",
			code: `sa "*t*0x";
			sb "foo*ebar";

			main() {
				extrn sa, sb;

				printf("*(*)***"*n");
				printf("%d %d %d*n", char(sa, 0), char(sa, 1), char(sa, 2));
				printf("%d %d %d %d %d*n", char(sb, 0), char(sb, 1), char(sb, 2), char(sb, 3), char(sb, 4));
			}`,
			wantStdout: `{}*"
9 0 120
102 111 111 0 98
`,
		},
		{
			name: "char_literals",
			code: `main() {
				printf("%d*n", '*0');
				printf("%d*n", '*e');
				printf("%d*n", '*t');
				printf("%d*n", '*n');
				printf("%d*n", '*r');
				printf("%c*n", '*(*)***'*"');
			}`,
			wantStdout: `0
0
9
10
13
{}*'"
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
