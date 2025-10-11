# blang - B compiler for modern systems

The B programming language was developed by Ken Thompson and Dennis Ritchie at Bell Labs in 1969 and was later replaced by C.

**blang** is a compiler for the B language, written in C.
It emits assembly code for X86_64 architecture.
Due to blang's simplicity, only **`gnu-linux-x86_64`**-systems are supported.

### Installation

To install blang, first build the project:
```
$ make
```
To install blang on your computer globally, use:
```
# make install
```
> **Warning**
> this requires root privileges and modifies system files

### Usage

To compile a B source file (`.b`), use:
```
$ blang <your file>
```

To get help, type:
```
$ blang --help
```

### Licensing
blang is licensed under the MIT License. See `LICENSE` in this repository for further information.

### References
- [Bell Labs User's Reference to B](https://www.bell-labs.com/usr/dmr/www/kbman.pdf) by Ken Thompson (Jan. 7, 1972)
- Wikipedia entry: [B (programming language)](https://en.wikipedia.org/wiki/B_(programming_language))
