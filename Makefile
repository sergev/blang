PROG    = blang
SRC     = main.go

.PHONY: all install clean

all: ${PROG}

run:
	go run . test.b
	wc a.s

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG} *.s

${PROG}: ${SRC}
	go build

#TODO: unit tests
#test:
