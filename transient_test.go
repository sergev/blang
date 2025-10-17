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
