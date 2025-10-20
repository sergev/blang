class Blang < Formula
  desc "Modern B programming language compiler with LLVM IR backend"
  homepage "https://github.com/sergev/blang"
  url "https://github.com/sergev/blang/archive/refs/heads/main.tar.gz"
  version "0.1"
  sha256 "ac7f0e1f3b52cf38606404e54f961775f7fca22c230c57b395c77770638bd28e"
  license "MIT"

  depends_on "go" => :build
  depends_on "llvm" => :build

  def install
    # Set up Go environment
    ENV["GOPATH"] = buildpath
    ENV["GOOS"] = "darwin"
    ENV["GOARCH"] = "arm64" if Hardware::CPU.arm?
    ENV["GOARCH"] = "amd64" if Hardware::CPU.intel?

    # Build the compiler
    system "go", "build", "-o", "blang"

    # Build the runtime library
    cd "runtime" do
      system "make", "CFLAGS=-O -Wall -ffreestanding"
    end

    # Install binary
    bin.install "blang"

    # Install runtime library
    lib.install "runtime/libb.a"

    # Install man page
    man1.install "doc/blang.1"

    # Install examples
    (share/"doc/blang").install Dir["examples/*.b"]
  end

  test do
    # Test version command
    assert_match "blang version", shell_output("#{bin}/blang --version")
    
    # Test basic compilation
    (testpath/"hello.b").write <<~EOS
      main() {
          write('Hello,');
          write(' World');
          write('!*n');
      }
    EOS

    system "#{bin}/blang", "hello.b", "-o", "hello"
    assert_equal "Hello, World!", shell_output("./hello").strip
  end
end
