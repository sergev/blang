package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		diffText := buildLineDiff(string(wantPDP7), string(got))
		t.Errorf("output mismatch (-want +got):\n%s", diffText)
	}
}

// Test function name() from examples/b.b
func TestPDP7CompilerName(t *testing.T) {
	ensureLibbOrSkip(t)

	code := `
name(s) {
	while (*s) {
		write(*s);
		s = &s[1];
	}
}

main() {
	auto csym, sym[8];
	csym = &sym[0];
	sym[2] = 'm';
	sym[3] = 'a';
	sym[4] = 'i';
	sym[5] = 'n';
	sym[6] = 0;
	name(&csym[2]);
	write(':');
	write('*n');
}
`

	got := compileLinkRunFromCode(t, "name_copy", code)
	want := "main:\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}

// Embed lookup(), name(), and storage from examples/b.b
// and verifies that lookup(); name(&csym[2]) prints the identifier.
func TestPDP7CompilerLookup(t *testing.T) {
	ensureLibbOrSkip(t)

	code := `
lookup() {
  extrn symtab, symbuf, eof, ns;
  auto np, sp, rp;

  rp = symtab;
  while (rp < ns) {
    np = &rp[2];
    sp = symbuf;
    while (*np==*sp) {
      if (!*np)
        return(rp);
      np = &np[1];
      sp = &sp[1];
    }
    while (*np)
      np = &np[1];
    rp = &np[1];
  }
  sp = symbuf;
  if (ns >= &symtab[290]) {
    error('sf');
    eof = 1;
    return(rp);
  }
  *ns = 0;
  ns[1] = 0;
  ns = &ns[2];
  while (*ns = *sp) {
    ns = &ns[1];
    sp = &sp[1];
  }
  ns = &ns[1];
  return(rp);
}

name(s) {
  while (*s) {
    write(*s);
    s = &s[1];
  }
}

/* minimal stub to satisfy reference from lookup() */
error(code) { return; }

/* storage copied from examples/b.b */
symtab[300] /* class value name */
  1, 5,'a','u','t','o', 0 ,
  1, 6,'e','x','t','r','n', 0 ,
  1,10,'g','o','t','o', 0 ,
  1,11,'r','e','t','u','r','n', 0 ,
  1,12,'i','f', 0 ,
  1,13,'w','h','i','l','e', 0 ,
  1,14,'e','l','s','e', 0 ;

ctab[]
    0,127,127,127,  0,127,127,127,
  127,126,126,127,127,127,127,127,
  127,127,127,127,127,127,127,127,
  127,127,127,127,127,127,127,127,
  126, 34,122,127,127, 44, 47,121,
    6,  7, 42, 40,  9, 41,127, 43,
  124,124,124,124,124,124,124,124,
  124,124,  8,  1, 63, 80, 65, 90,
  127,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,  4,127,  5, 48,127,
  127,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,  2, 48,  3,127,127;

symbuf[10];
peeksym -1;
peekc;
eof;
line 1;
csym;
ns;
cval;
isn;
nerror;
nauto;

main() {
  ns = &symtab[51];
  symbuf[0] = 'm'; symbuf[1] = 'a'; symbuf[2] = 'i'; symbuf[3] = 'n'; symbuf[4] = 0;
  csym = lookup();
  name(&csym[2]);
  write(':');
  write('*n');
}
`

	got := compileLinkRunFromCode(t, "transient_lookup_name", code)
	want := "main:\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}

// Embed symbol(), name(), and required helpers and storage copied from examples/b.b.
// Feed an identifier via stdin and verify that symbol(); name(&csym[2]) prints that identifier.
func TestPDP7CompilerSymbol(t *testing.T) {
	ensureLibbOrSkip(t)

	code := `
lookup() {
  extrn symtab, symbuf, eof, ns;
  auto np, sp, rp;

  rp = symtab;
  while (rp < ns) {
    np = &rp[2];
    sp = symbuf;
    while (*np==*sp) {
      if (!*np)
        return(rp);
      np = &np[1];
      sp = &sp[1];
    }
    while (*np)
      np = &np[1];
    rp = &np[1];
  }
  sp = symbuf;
  if (ns >= &symtab[290]) {
    error('sf');
    eof = 1;
    return(rp);
  }
  *ns = 0;
  ns[1] = 0;
  ns = &ns[2];
  while (*ns = *sp) {
    ns = &ns[1];
    sp = &sp[1];
  }
  ns = &ns[1];
  return(rp);
}

symbol() {
  extrn symbuf, ctab, peeksym, peekc, eof, line, csym, cval;
  auto b, c, ct, sp;

  if (peeksym>=0) {
    c = peeksym;
    peeksym = -1;
    return(c);
  }
  if (peekc) {
    c = peekc;
    peekc = 0;
  } else {
    if (eof)
      return(0);
    c = read();
  }
loop:
  ct = ctab[c];

  if (ct==0) { /* eof */
    eof = 1;
    return(0);
  }

  if (ct==126) { /* white space */
    if (c=='*n')
      line = line+1;
    c = read();
    goto loop;
  }

  if (c=='=')
    return(subseq('=',80,60));

  if (c=='<')
    return(subseq('=',63,62));

  if (c=='>')
    return(subseq('=',65,64));

  if (c=='!')
    return(subseq('=',34,61));

  if (c=='$') {
    if (subseq('(',0,1))
      return(2);
    if (subseq(')',0,1))
      return(3);
  }
  if (c=='/') {
    if (subseq('**',1,0))
      return(43);
com:
    c = read();
com1:
    if (c==4) {
      eof = 1;
      error('**/'); /* eof */
      return(0);
    }
    if (c=='*n')
      line = line+1;
    if (c!='**')
      goto com;
    c = read();
    if (c!='/')
      goto com1;
    c = read();
    goto loop;
  }
  if (ct==124) { /* number */
    cval = 0;
    if (c=='0')
      b = 8;
    else
      b = 10;
    while(c >= '0' & c <= '9') {
      cval = cval*b + c -'0';
      c = read();
    }
    peekc = c;
    return(21);
  }
  if (c=='*'') { /* ' */
    getcc();
    return(21);
  }
  if (ct==123) { /* letter */
    sp = symbuf;
    while(ct==123 | ct==124) {
      if (sp < &symbuf[9]) {
        *sp = c;
        sp = &sp[1];
      }
      ct = ctab[c = read()];
    }
    *sp = 0;
    peekc = c;
    csym = lookup();
    if (csym[0]==1) {
      cval = csym[1];
      return(19); /* keyword */
    }
    return(20); /* name */
  }
  if (ct==127) { /* unknown */
    error('sy');
    c = read();
    goto loop;
  }
  return(ctab[c]);
}

subseq(c,a,b) {
  extrn peekc;

  if (!peekc)
    peekc = read();
  if (peekc != c)
    return(a);
  peekc = 0;
  return(b);
}

getcc() {
  extrn cval;
  auto c;

  cval = 0;
  if ((c = mapch('*'')) < 0)
    return;
  cval = c;
  if ((c = mapch('*'')) < 0)
    return;
  cval = cval * 256 + c;
  if (mapch('*'') >= 0)
    error('cc');
}

mapch(c) {
  extrn peekc;
  auto a;

  if ((a=read())==c)
    return(-1);

  if (a=='*n' | a==0 | a==4) {
    error('cc');
    peekc = a;
    return(-1);
  }

  if (a=='**') {
    a=read();

    if (a=='0')
      return(0);

    if (a=='e')
      return(4);

    if (a=='(')
      return('{');

    if (a==')')
      return('}');

    if (a=='t')
      return('*t');

    if (a=='r')
      return('*r');

    if (a=='n')
      return('*n');
  }
  return(a);
}

name(s) {
  while (*s) {
    write(*s);
    s = &s[1];
  }
}

/* minimal stub to satisfy references */
error(code) { return; }

/* storage copied from examples/b.b */
symtab[300] /* class value name */
  1, 5,'a','u','t','o', 0 ,
  1, 6,'e','x','t','r','n', 0 ,
  1,10,'g','o','t','o', 0 ,
  1,11,'r','e','t','u','r','n', 0 ,
  1,12,'i','f', 0 ,
  1,13,'w','h','i','l','e', 0 ,
  1,14,'e','l','s','e', 0 ;

ctab[]
    0,127,127,127,  0,127,127,127,
  127,126,126,127,127,127,127,127,
  127,127,127,127,127,127,127,127,
  127,127,127,127,127,127,127,127,
  126, 34,122,127,127, 44, 47,121,
    6,  7, 42, 40,  9, 41,127, 43,
  124,124,124,124,124,124,124,124,
  124,124,  8,  1, 63, 80, 65, 90,
  127,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,  4,127,  5, 48,127,
  127,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,123,123,123,123,123,
  123,123,123,  2, 48,  3,127,127;

symbuf[10];
peeksym -1;
peekc;
eof;
line 1;
csym;
ns;
cval;
isn;
nerror;
nauto;

main() {
  ns = &symtab[51];
  symbol();
  name(&csym[2]);
  write(':');
  write('*n');
}
`

	// Compile and link
	dir, bFile, llFile, exeFile := createTempBFile(t, "sym_then_name", code)
	_ = dir
	compileToLL(t, bFile, llFile)
	linkWithClang(t, llFile, exeFile)

	// Run with stdin providing an identifier
	cmd := exec.Command(exeFile)
	cmd.Stdin = strings.NewReader("main\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run program: %v", err)
	}
	got := string(out)
	want := "main:\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}
