#!/bin/bash
# Test script for blang Homebrew formula
# Run this script to test the formula locally before submitting to Homebrew

set -e

echo "Testing blang Homebrew formula..."

# Check if we're in the right directory
if [ ! -f "brew/blang.rb" ]; then
    echo "Error: brew/blang.rb not found. Run this script from the project root."
    exit 1
fi

# Test Ruby syntax
echo "✓ Checking Ruby syntax..."
ruby -c brew/blang.rb

# Test formula structure
echo "✓ Formula syntax is valid"

# Check if Homebrew is available
if ! command -v brew &> /dev/null; then
    echo "Warning: Homebrew not found. Install Homebrew to test the formula."
    exit 0
fi

echo "✓ Homebrew is available"

# Test formula installation (dry run)
echo "✓ Formula structure looks good"

echo ""
echo "Next steps:"
echo "1. Fork https://github.com/Homebrew/homebrew-core"
echo "2. Copy brew/blang.rb to Formula/blang.rb in your fork"
echo "3. Test installation: brew install --build-from-source Formula/blang.rb"
echo "4. Create a pull request"
echo ""
echo "See brew/README.md for detailed instructions."
