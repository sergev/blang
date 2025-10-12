PROG    = blang
SRC     = main.go

.PHONY: all install clean test

all: ${PROG}

test:
	go test -v

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG} *.s

${PROG}: ${SRC}
	go build
