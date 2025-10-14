PROG = blang

.PHONY: all install clean test cover bench

all: libb.o
	go build

install: all
	install -m 555 ${PROG} /usr/local/bin/${PROG}

clean:
	rm -f ${PROG} *.o *.ll

libb.o: runtime/libb.c
	$(CC) -c -ffreestanding $< -o $@

#
# For testing, please install gotestsum:
#	go install gotest.tools/gotestsum@latest
#
test:
	gotestsum --format dots

cover:
	gotestsum -- -cover .

#
# Run benchmark
#
bench:
	go test -bench=BenchmarkCompile -benchmem
