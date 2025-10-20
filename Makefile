#
# B compiler
#
PROG    = blang
DESTDIR	= $(HOME)/.local

.PHONY: all install uninstall clean test cover bench gotestsum

all:
	go build
	$(MAKE) -C runtime $@

install: all
	@install -d $(DESTDIR)/usr/bin
	install -m 555 ${PROG} $(DESTDIR)/usr/bin/${PROG}
	@install -d $(DESTDIR)/usr/lib
	install -m 644 runtime/libb.a $(DESTDIR)/usr/lib/libb.a
	@install -d $(DESTDIR)/usr/share/man/man1
	install -m 644 doc/blang.1 $(DESTDIR)/usr/share/man/man1/blang.1
	@install -d $(DESTDIR)/usr/share/doc/blang
	install -m 644 examples/*.b $(DESTDIR)/usr/share/doc/blang/

uninstall:
	rm -f $(DESTDIR)/bin/${PROG}
	$(MAKE) -C runtime $@

clean:
	rm -f ${PROG} *.o *.ll
	$(MAKE) -C runtime $@

#
# For testing, please install gotestsum:
#	go install gotest.tools/gotestsum@latest
#
test: gotestsum
	gotestsum --format dots

cover: gotestsum
	gotestsum -- -cover .

gotestsum:
	@command -v gotestsum >/dev/null || go install gotest.tools/gotestsum@latest

#
# Run benchmark
#
bench:
	go test -bench=BenchmarkCompile -benchmem

#
# Build Debian package
#
deb: all
	@echo "Building Debian package..."
	@mkdir -p debian-pkg/DEBIAN
	@mkdir -p debian-pkg/usr/bin
	@mkdir -p debian-pkg/usr/lib
	@mkdir -p debian-pkg/usr/share/man/man1
	@mkdir -p debian-pkg/usr/share/doc/blang
	@$(MAKE) install DESTDIR=debian-pkg
	@echo "Package: blang" > debian-pkg/DEBIAN/control
	@echo "Version: 0.1-1" >> debian-pkg/DEBIAN/control
	@echo "Architecture: amd64" >> debian-pkg/DEBIAN/control
	@echo "Maintainer: Serge Vakulenko <serge@vakulenko.org>" >> debian-pkg/DEBIAN/control
	@echo "Depends: libc6, clang" >> debian-pkg/DEBIAN/control
	@echo "Description: B programming language compiler" >> debian-pkg/DEBIAN/control
	@echo " A modern B programming language compiler written in Go with LLVM IR backend" >> debian-pkg/DEBIAN/control
	@echo " and clang-like command-line interface." >> debian-pkg/DEBIAN/control
	@cp debian/copyright debian-pkg/DEBIAN/
	@if [ -f debian-pkg/usr/share/man/man1/blang.1 ]; then gzip -9 debian-pkg/usr/share/man/man1/blang.1; fi
	@dpkg-deb --build debian-pkg blang_0.1-1_amd64.deb
	@rm -rf debian-pkg
	@echo "Package created: blang_0.1-1_amd64.deb"
