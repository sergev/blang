PROG    = blang
SRC     = main.go

.PHONY: all install clean

all: ${PROG}

run:
	go run .

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG}

${PROG}: ${SRC}
	go build

#TODO: unit tests
#test:
