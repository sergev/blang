# Enhanced Makefile demonstrating new blang CLI features

# Basic compilation with automatic linking
hello: hello.b ../blang ../libb.o
	../blang -o hello hello.b

# Optimized release build
hello-opt: hello.b ../blang ../libb.o
	../blang -O2 -o hello-opt hello.b

# Debug build with verbose output
hello-debug: hello.b ../blang ../libb.o
	../blang -g -v -Wall -o hello-debug hello.b

# Generate LLVM IR
hello.ll: hello.b ../blang
	../blang -emit-llvm -o hello.ll hello.b

# Generate object file
hello.o: hello.b ../blang ../libb.o
	../blang -c -o hello.o hello.b

# Generate assembly
hello.s: hello.b ../blang ../libb.o
	../blang -S -o hello.s hello.b

# Clean up generated files
clean:
	rm -f hello hello-opt hello-debug hello.ll hello.o hello.s *.tmp.ll

# Show all available targets
help:
	@echo "Available targets:"
	@echo "  hello      - Basic executable (automatic linking)"
	@echo "  hello-opt  - Optimized executable"
	@echo "  hello-debug- Debug executable with verbose output"
	@echo "  hello.ll   - LLVM IR output"
	@echo "  hello.o    - Object file"
	@echo "  hello.s    - Assembly output"
	@echo "  clean      - Remove generated files"
	@echo "  help       - Show this help"

.PHONY: clean help
