package main

import "testing"

// Transient test, which can be used for debugging any B program.
// Feel free to replace, suit yourself.
func TestTransient(t *testing.T) {
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

// TestTransient_LookupThenName embeds lookup(), name(), and storage from examples/b.b
// and verifies that lookup(); name(&csym[2]) prints the identifier.
func TestTransient_LookupThenName(t *testing.T) {
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
