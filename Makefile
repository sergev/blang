#
# B compiler
#
PROG    = blang
DESTDIR	= $(HOME)/.local

.PHONY: all install uninstall clean test cover bench gotestsum source

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
	rm -f *.o *.ll ${PROG}
	rm -f ${PROG}_*.deb ${PROG}_*.gz ${PROG}_*.xz ${PROG}_*.dsc ${PROG}_*.build ${PROG}_*.buildinfo ${PROG}_*.changes
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

# Debian-specific variables (only evaluated for deb/source targets)
deb source: VERSION    = $(shell dpkg-parsechangelog -S Version)
deb source: UPSTREAM   = $(shell dpkg-parsechangelog --show-field Version | cut -d- -f1)
deb source: MAINTAINER = $(shell dpkg-parsechangelog --show-field Maintainer)
deb source: ARCH       = $(shell dpkg --print-architecture)

#
# Build Debian package
#
deb: all
	@echo "Building Debian package..."
	@rm -rf debian-pkg
	@mkdir -p debian-pkg/DEBIAN
	@mkdir -p debian-pkg/usr/bin
	@mkdir -p debian-pkg/usr/lib
	@mkdir -p debian-pkg/usr/share/man/man1
	@mkdir -p debian-pkg/usr/share/doc/$(PROG)
	@$(MAKE) install DESTDIR=debian-pkg
	@echo "Package: $(PROG)" > debian-pkg/DEBIAN/control
	@echo "Version: $(VERSION)" >> debian-pkg/DEBIAN/control
	@echo "Architecture: $(ARCH)" >> debian-pkg/DEBIAN/control
	@echo "Maintainer: $(MAINTAINER)" >> debian-pkg/DEBIAN/control
	@echo "Depends: libc6, clang" >> debian-pkg/DEBIAN/control
	@echo "Description: Compiler for the B programming language" >> debian-pkg/DEBIAN/control
	@echo " A modern B programming language compiler written in Go with LLVM IR backend" >> debian-pkg/DEBIAN/control
	@echo " and clang-like command-line interface." >> debian-pkg/DEBIAN/control
	@cp debian/copyright debian-pkg/DEBIAN/
	@if [ -f debian-pkg/usr/share/man/man1/$(PROG).1 ]; then gzip -9 debian-pkg/usr/share/man/man1/$(PROG).1; fi
	@dpkg-deb --build --root-owner-group debian-pkg $(PROG)_$(VERSION)_$(ARCH).deb
	@rm -rf debian-pkg
	@echo "Package created:"
	@dpkg-deb -c $(PROG)_*.deb

#
# Build Debian source package
#
source:
	@echo "Building Debian source package..."
	@rm -rf debian-source
	@mkdir -p debian-source
	@echo "Copying files from manifest..."
	tar cfT - debian/manifest | tar xf - -C debian-source
	@echo "Creating upstream tarball..."
	@cd debian-source && tar --transform "s,^\.,$(PROG)-$(UPSTREAM)," -czf ../$(PROG)_$(UPSTREAM).orig.tar.gz .
	@echo "Building source package..."
	cd debian-source && debuild -S -sa
	@echo "Running lintian checks..."
	lintian $(PROG)_$(VERSION).dsc
	@rm -rf debian-source
	@echo "Source package created:"
	@ls -lh $(PROG)_*-*.dsc $(PROG)_*.orig.tar.gz $(PROG)_*.debian.tar.xz
