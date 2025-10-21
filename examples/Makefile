# Enhanced Makefile demonstrating new blang CLI features
.PHONY: clean help

# Show all available targets
help:
	@echo "Available targets:"
	@echo "  hello       - Basic executable (automatic linking)"
	@echo "  hello-opt   - Optimized executable"
	@echo "  hello-debug - Debug executable with verbose output"
	@echo "  hello.ll    - LLVM IR output"
	@echo "  hello.o     - Object file"
	@echo "  hello.s     - Assembly output"
	@echo "  clean       - Remove generated files"
	@echo "  help        - Show this help"

# Basic compilation with automatic linking
hello: hello.b ../blang
	../blang -L../runtime hello.b

# Optimized release build
hello-opt: hello.b ../blang
	../blang -O2 -L../runtime -o hello-opt hello.b

# Debug build with verbose output
hello-debug: hello.b ../blang
	../blang -g -v -Wall -L../runtime -o hello-debug hello.b

# Generate LLVM IR
hello.ll: hello.b ../blang
	../blang --emit-llvm hello.b

# Generate object file
hello.o: hello.b ../blang
	../blang -c hello.b

# Generate assembly
hello.s: hello.b ../blang
	../blang -S hello.b

# Clean up generated files
clean:
	rm -f hello hello-opt hello-debug hello.ll hello.o hello.s *.tmp.ll b
