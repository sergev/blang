package main

import (
	"testing"
)

// TestCompoundAssignments tests compound assignment operators (from oldtests/assignment_test.cpp)
func TestCompoundAssignments(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "assign_add",
			code: `main() {
				auto x;
				x = 10;
				x =+ 5;
				printf("%d*n", x);
			}`,
			wantStdout: "15\n",
		},
		{
			name: "assign_subtract",
			code: `main() {
				auto x;
				x = 10;
				x =- 3;
				printf("%d*n", x);
			}`,
			wantStdout: "7\n",
		},
		{
			name: "assign_multiply",
			code: `main() {
				auto x;
				x = 4;
				x =* 3;
				printf("%d*n", x);
			}`,
			wantStdout: "12\n",
		},
		{
			name: "assign_divide",
			code: `main() {
				auto x;
				x = 15;
				x =/ 3;
				printf("%d*n", x);
			}`,
			wantStdout: "5\n",
		},
		{
			name: "assign_modulo",
			code: `main() {
				auto x;
				x = 17;
				x =% 5;
				printf("%d*n", x);
			}`,
			wantStdout: "2\n",
		},
		{
			name: "assign_shift_left",
			code: `main() {
				auto x;
				x = 2;
				x =<< 2;
				printf("%d*n", x);
			}`,
			wantStdout: "8\n",
		},
		{
			name: "assign_shift_right",
			code: `main() {
				auto x;
				x = 16;
				x =>> 2;
				printf("%d*n", x);
			}`,
			wantStdout: "4\n",
		},
		{
			name: "assign_less_or_equal",
			code: `main() {
				auto x;
				x = 5;
				x =<= 3;
				printf("%d*n", x);
				x = 2;
				x =<= 3;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_less_than",
			code: `main() {
				auto x;
				x = 5;
				x =< 5;
				printf("%d*n", x);
				x = 4;
				x =< 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_greater_or_equal",
			code: `main() {
				auto x;
				x = 3;
				x =>= 5;
				printf("%d*n", x);
				x = 5;
				x =>= 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_greater_than",
			code: `main() {
				auto x;
				x = 4;
				x => 5;
				printf("%d*n", x);
				x = 6;
				x => 5;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_not_equal",
			code: `main() {
				auto x;
				x = 5;
				x =!= 5;
				printf("%d*n", x);
				x = 5;
				x =!= 3;
				printf("%d*n", x);
			}`,
			wantStdout: "0\n1\n",
		},
		{
			name: "assign_equal",
			code: `main() {
				auto x;
				x = 5;
				x === 5;
				printf("%d*n", x);
				x = 5;
				x === 6;
				printf("%d*n", x);
			}`,
			wantStdout: "1\n0\n",
		},
		{
			name: "assign_and",
			code: `main() {
				auto x;
				x = 12;
				x =& 10;
				printf("%d*n", x);
			}`,
			wantStdout: "8\n",
		},
		{
			name: "assign_or",
			code: `main() {
				auto x;
				x = 12;
				x =| 10;
				printf("%d*n", x);
			}`,
			wantStdout: "14\n",
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
