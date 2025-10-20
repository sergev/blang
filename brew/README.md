# Homebrew Formula for Blang

This directory contains the Homebrew formula for the Blang B programming language compiler.

## Files

- `blang.rb` - The Homebrew formula definition

## Submitting to Homebrew

To submit this formula to the official Homebrew repository, follow these steps:

### 1. Fork the Homebrew Core Repository

```bash
# Fork https://github.com/Homebrew/homebrew-core on GitHub
# Then clone your fork locally
git clone https://github.com/YOUR_USERNAME/homebrew-core.git
cd homebrew-core
```

### 2. Create a New Branch

```bash
git checkout -b blang
```

### 3. Add the Formula

```bash
# Copy the formula to the correct location
cp /path/to/blang/brew/blang.rb Formula/blang.rb
```

### 4. Update the Formula

Before submitting, you need to:

1. **Get the correct SHA256 hash**:
   ```bash
   # Download the source tarball
   curl -L https://github.com/sergev/blang/archive/refs/heads/main.tar.gz -o blang-main.tar.gz
   
   # Calculate SHA256
   shasum -a 256 blang-main.tar.gz
   ```

2. **Update the formula** with the correct SHA256 hash:
   ```ruby
   sha256 "ACTUAL_SHA256_HASH_HERE"
   ```

3. **Test the formula locally**:
   ```bash
   brew install --build-from-source Formula/blang.rb
   ```

### 5. Test Installation

```bash
# Test that blang works
blang --version

# Test compilation
echo 'main() { extrn putchar; putchar(72); putchar(10); }' > test.b
blang test.b -o test
./test
```

### 6. Commit and Push

```bash
git add Formula/blang.rb
git commit -m "blang: add formula

A modern B programming language compiler written in Go with LLVM IR backend."
git push origin blang
```

### 7. Create Pull Request

1. Go to https://github.com/Homebrew/homebrew-core
2. Click "New pull request"
3. Select your fork and the `blang` branch
4. Fill out the PR template with:
   - Description of the software
   - Why it's useful
   - Any special considerations

### 8. Homebrew Review Process

The Homebrew maintainers will:
- Review the formula for correctness
- Test installation on multiple macOS versions
- Check for any issues with dependencies
- Provide feedback if changes are needed

## Formula Details

The formula installs:
- `blang` executable in `/usr/local/bin/` (or `/opt/homebrew/bin/` on Apple Silicon)
- `libb.a` runtime library in `/usr/local/lib/` (or `/opt/homebrew/lib/`)
- `blang.1` man page in `/usr/local/share/man/man1/`
- Example `.b` files in `/usr/local/share/doc/blang/`

## Dependencies

- **Go**: Required to build the compiler
- **LLVM**: Required for the LLVM IR backend and linking

## Testing

The formula includes a test that:
1. Creates a simple B program
2. Compiles it with blang
3. Runs the executable and verifies output

## Troubleshooting

### Common Issues

1. **SHA256 mismatch**: Make sure you're using the correct hash for the main branch tarball
2. **Build failures**: Ensure all dependencies are properly specified
3. **Test failures**: Verify the test program works manually before submitting

### Getting Help

- Homebrew documentation: https://docs.brew.sh/
- Formula cookbook: https://docs.brew.sh/Formula-Cookbook
- Homebrew maintainers: https://github.com/Homebrew/homebrew-core/issues

## Alternative: Homebrew Tap

If you prefer not to submit to the main Homebrew repository, you can create your own tap:

```bash
# Create a tap repository on GitHub named homebrew-blang
# Then users can install with:
brew install YOUR_USERNAME/blang/blang
```

This is easier but requires users to know about your tap.
