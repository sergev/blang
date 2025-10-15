package main

import (
	"testing"
)

// TestNestedLoops tests nested while loops with unique labels
func TestNestedLoops(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		code       string
		wantStdout string
	}{
		{
			name: "nested_while_basic",
			code: `main() {
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
			}`,
			wantStdout: "sum = 9\n",
		},
		{
			name: "nested_while_complex",
			code: `main() {
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
			}`,
			wantStdout: "count = 8\n",
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
