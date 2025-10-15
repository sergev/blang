#
# Runtime library
#
LIB     = libb.a
DESTDIR	= $(HOME)/.local
CFLAGS  = -O -Wall -ffreestanding
OBJS    = char.o \
          exit.o \
          flush.o \
          lchar.o \
          nread.o \
          nwrite.o \
          printd.o \
          printf.o \
          printo.o \
          read.o \
          start.o \
          write.o \
          writeb.o

all: $(LIB)

install:
	install -m 444 $(LIB) $(DESTDIR)/lib/$(LIB)

clean:
	rm -rf *.o *.a

$(LIB): $(OBJS)
	@rm -f $@
	ar cr $@ $(OBJS)
