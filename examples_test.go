package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TestCompileAndRun tests the full pipeline: compile, link, and execute
func TestCompileAndRun(t *testing.T) {
	ensureLibbOrSkip(t)

	tests := []struct {
		name       string
		inputFile  string
		wantExit   int
		wantStdout string
	}{
		// Example programs
		{
			name:       "hello_write",
			inputFile:  "examples/hello.b",
			wantExit:   0,
			wantStdout: "Hello, World!",
		},
		{
			name:       "hello_printf",
			inputFile:  "examples/helloworld.b",
			wantExit:   0,
			wantStdout: "Hello, World!",
		},
		{
			name:       "example_fibonacci",
			inputFile:  "examples/fibonacci.b",
			wantExit:   0,
			wantStdout: "55\n",
		},
		{
			name:       "example_fizzbuzz",
			inputFile:  "examples/fizzbuzz.b",
			wantExit:   0,
			wantStdout: "FizzBuzz", // Check that FizzBuzz appears in output
		},
		// Note: example_e2 is tested separately in TestE2Constant due to long runtime
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			llFile := filepath.Join(tmpDir, tt.name+".ll")
			exeFile := filepath.Join(tmpDir, tt.name)

			compileToLL(t, tt.inputFile, llFile)
			linkWithClang(t, llFile, exeFile)
			stdout, exitCode := runExecutable(t, exeFile)

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}

			if tt.wantStdout != "" {
				gotStdout := string(stdout)
				if !hasSubstring(gotStdout, tt.wantStdout) {
					t.Errorf("Stdout = %q, want substring %q", gotStdout, tt.wantStdout)
				}
			}
		})
	}
}

// TestE2Constant tests the e-2 constant calculation
func TestE2Constant(t *testing.T) {
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	inputFile := "examples/e-2.b"
	llFile := filepath.Join(tmpDir, "e-2.ll")
	exeFile := filepath.Join(tmpDir, "e2")

	compileToLL(t, inputFile, llFile)
	linkWithClang(t, llFile, exeFile)

	// Run the executable with a 3s timeout (preserving error text)
	stdout, _ := runWithTimeout(t, exeFile, 3*time.Second)

	// Check that output starts with expected first line
	wantPrefix := "71828 18284 59045 23536 02874"
	gotStdout := string(stdout)
	if !strings.HasPrefix(gotStdout, wantPrefix) {
		t.Errorf("Output does not start with expected prefix.\nWant prefix: %q\nGot: %q", wantPrefix, gotStdout[:min(len(gotStdout), 100)])
	}
}

// The test compiles the historical PDP-7 B compiler (examples/b.b),
// runs it with examples/b.b as input, and verifies the generated
// output matches the expected PDP-7 code in examples/b.pdp7.
func TestPDP7CompilerB(t *testing.T) {
	ensureLibbOrSkip(t)

	tmpDir := t.TempDir()
	llFile := filepath.Join(tmpDir, "b.ll")
	exeFile := filepath.Join(tmpDir, "b")

	// Compile the PDP-7 B compiler source to IR and link into an executable.
	compileToLL(t, "examples/b.b", llFile)
	linkWithClang(t, llFile, exeFile)

	in, err := os.Open("examples/b.b")
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer in.Close()

	cmd := exec.Command(exeFile)
	cmd.Stdin = in
	got, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("program exited with code %d: %s", ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("run: %v", err)
	}

	// Read expected output
	wantPDP7, err := os.ReadFile("examples/b.pdp7")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(got, []byte(wantPDP7)) {
		diffText := buildLineDiff(string(wantPDP7), string(got), 2, 100)
		t.Errorf("output mismatch (-want +got):\n%s", diffText)
	}
}

// buildLineDiff returns a unified-style diff of want vs got with the given
// number of context lines and a maximum number of printed lines.
func buildLineDiff(want, got string, contextLines, maxLines int) string {
	dmp := diffmatchpatch.New()

	// Normalize line endings to avoid spurious diffs across platforms.
	want = strings.ReplaceAll(want, "\r\n", "\n")
	got = strings.ReplaceAll(got, "\r\n", "\n")

	// Produce a line-based diff using diff-match-patch utilities.
	aChars, bChars, lineArray := dmp.DiffLinesToChars(want, got)
	diffs := dmp.DiffMain(aChars, bChars, false)
	dmp.DiffCleanupSemantic(diffs)
	diffs = dmp.DiffCharsToLines(diffs, lineArray)

	// Expand to per-line operations.
	type lineOp struct {
		op   diffmatchpatch.Operation
		text string
	}
	var ops []lineOp
	for _, d := range diffs {
		parts := strings.SplitAfter(d.Text, "\n")
		for _, p := range parts {
			if p == "" {
				continue
			}
			ops = append(ops, lineOp{op: d.Type, text: p})
		}
	}

	var buf strings.Builder
	printed := 0
	appendLine := func(prefix byte, s string) bool {
		if printed >= maxLines {
			return false
		}
		buf.WriteByte(prefix)
		buf.WriteString(s)
		printed++
		return true
	}

	n := len(ops)
	i := 0
	truncated := false
	for i < n {
		for i < n && ops[i].op == diffmatchpatch.DiffEqual {
			i++
		}
		if i >= n {
			break
		}
		// Start of a hunk at i; include up to contextLines equal lines before.
		start := i
		back := 0
		j := i - 1
		for j >= 0 && back < contextLines && ops[j].op == diffmatchpatch.DiffEqual {
			start = j
			back++
			j--
		}

		// Determine end of hunk, merging nearby changes if separated by <= contextLines equals.
		end := i
		eqAfter := 0
		k := i
		for k < n {
			if ops[k].op == diffmatchpatch.DiffEqual {
				eqAfter++
				if eqAfter > contextLines {
					break
				}
			} else {
				end = k
				eqAfter = 0
			}
			k++
		}
		// Include trailing context equals up to contextLines.
		u := end + 1
		ctx := 0
		for u < n && ctx < contextLines && ops[u].op == diffmatchpatch.DiffEqual {
			end = u
			ctx++
			u++
		}

		// Emit lines in [start, end].
		for idx := start; idx <= end; idx++ {
			var prefix byte
			switch ops[idx].op {
			case diffmatchpatch.DiffEqual:
				prefix = ' '
			case diffmatchpatch.DiffDelete:
				prefix = '-'
			case diffmatchpatch.DiffInsert:
				prefix = '+'
			default:
				prefix = ' '
			}
			if !appendLine(prefix, ops[idx].text) {
				truncated = true
				break
			}
		}
		if truncated {
			break
		}
		i = end + 1
	}

	if truncated {
		buf.WriteString("... (diff truncated)\n")
	}
	return buf.String()
}
