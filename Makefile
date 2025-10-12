PROG    = blang
SRC     = codegen.go compiler.go control.go expr.go lexer.go main.go parser.go

.PHONY: all install clean test

all: ${PROG} libb.o

test:
	go test -v

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG} *.o *.ll

${PROG}: ${SRC}
	go build

libb.o: libb/libb.c
	$(CC) -c -ffreestanding $< -o $@
