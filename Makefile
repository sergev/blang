PROG    = blang
SRC     = codegen.go compiler.go lexer.go list.go llvm_codegen.go llvm_expr.go llvm_parser.go main.go parser.go

.PHONY: all install clean test

all: ${PROG}

test:
	go test -v

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG} *.ll

${PROG}: ${SRC}
	go build
