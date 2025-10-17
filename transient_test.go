package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestTransient_SymbolThenName embeds symbol(), name(), and required helpers
// and storage copied from examples/b.b. It feeds an identifier via stdin and
// verifies that symbol(); name(&csym[2]) prints that identifier.
func TestTransient_SymbolThenName(t *testing.T) {
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
