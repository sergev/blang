#
# B compiler
#
PROG    = blang
DESTDIR	= $(HOME)/.local

.PHONY: all install clean test cover bench

all:
	go build
	$(MAKE) -C runtime $@

install: all
	-mkdir -p $(DESTDIR)/bin $(DESTDIR)/lib
	install -m 555 ${PROG} $(DESTDIR)/bin/${PROG}
	$(MAKE) -C runtime $@

clean:
	rm -f ${PROG} *.o *.ll
	$(MAKE) -C runtime $@

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
