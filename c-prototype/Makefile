PROG    = blang
SRC     = compiler.c list.c main.c
CFLAGS  = -O -g -Wall -Wextra -Wshadow

.PHONY: all install clean

all: ${PROG}

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f *.o *.a *.out ${PROG}
	rm -rf ${PROG}.dSYM

${PROG}: ${SRC}
	${CC} ${CFLAGS} ${SRC} -o $@
