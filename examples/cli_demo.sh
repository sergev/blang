#!/bin/bash
# blang CLI Features Demonstration Script

echo "=== blang CLI Features Demonstration ==="
echo ""

# Check if we're in the right directory
if [ ! -f "../blang" ]; then
    echo "Error: Run this script from the examples/ directory"
    echo "Usage: ./cli_demo.sh"
    exit 1
fi

# Check if libb.o exists
if [ ! -f "../libb.o" ]; then
    echo "Error: libb.o not found. Run 'make' from the project root first."
    exit 1
fi

echo "1. Basic CLI Options"
echo "==================="
echo "Help output:"
../blang -help | head -5
echo ""
echo "Version output:"
../blang -version
echo ""

echo "2. Output Formats"
echo "================="
echo "a) Default executable (automatic linking):"
../blang -v -o demo_hello hello.b
echo "   Generated: $(ls -la demo_hello 2>/dev/null || echo 'No demo_hello')"
echo "   Running: $(./demo_hello 2>/dev/null || echo 'Failed to run')"
echo ""

echo "b) LLVM IR output:"
../blang -emit-llvm -o demo.ll hello.b
echo "   Generated: $(ls -la demo.ll 2>/dev/null || echo 'No demo.ll')"
echo ""

echo "c) Object file output:"
../blang -c -o demo.o hello.b
echo "   Generated: $(ls -la demo.o 2>/dev/null || echo 'No demo.o')"
echo ""

echo "d) Assembly output:"
../blang -S -o demo.s hello.b
echo "   Generated: $(ls -la demo.s 2>/dev/null || echo 'No demo.s')"
echo ""

echo "e) Preprocessed output:"
../blang -E -o demo.i hello.b
echo "   Generated: $(ls -la demo.i 2>/dev/null || echo 'No demo.i')"
echo ""

echo "3. Optimization Levels"
echo "======================"
for opt in O0 O1 O2 O3; do
    echo "Optimization level $opt:"
    ../blang -$opt -o demo_$opt hello.b
    echo "   Generated: $(ls -la demo_$opt 2>/dev/null || echo 'No demo_$opt')"
done
echo ""

echo "4. Debug and Verbose Options"
echo "============================"
echo "Debug build with verbose output:"
../blang -g -v -O2 -o demo_debug hello.b
echo ""

echo "5. Warning Options"
echo "=================="
echo "Build with all warnings:"
../blang -Wall -v -o demo_warnings hello.b
echo ""

echo "6. Path and Library Options"
echo "==========================="
echo "Build with include/library paths:"
../blang -I /tmp -L /tmp -l c -o demo_paths hello.b
echo ""

echo "7. Save Temporary Files"
echo "======================="
echo "Build with temporary files preserved:"
../blang -save-temps -v -o demo_save_temps hello.b
echo "   Temporary files: $(ls -la demo_save_temps.tmp.* 2>/dev/null || echo 'No temp files')"
echo ""

echo "8. Combined Flags"
echo "================="
echo "All flags combined:"
../blang -v -O3 -g -Wall -save-temps -std b -o demo_all hello.b
echo ""

echo "9. Error Handling"
echo "================="
echo "Testing error conditions:"
echo "a) No input files:"
../blang 2>&1 | head -2
echo ""
echo "b) Invalid file extension:"
../blang test.txt 2>&1 | head -2
echo ""
echo "c) Nonexistent file:"
../blang nonexistent.b 2>&1 | head -2
echo ""

echo "10. Generated Files Summary"
echo "==========================="
echo "All generated files:"
ls -la demo_* demo.* 2>/dev/null || echo "No demo files found"
echo ""

echo "11. Cleanup"
echo "==========="
read -p "Clean up generated files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f demo_* demo.* *.tmp.ll
    echo "Cleaned up generated files."
else
    echo "Generated files preserved."
fi

echo ""
echo "=== Demonstration Complete ==="
echo "For more information, see:"
echo "  - README.md - Quick start guide"
echo "  - doc/CLI.md - Comprehensive CLI documentation"
echo "  - doc/Testing.md - Testing guide including CLI tests"
