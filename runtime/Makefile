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

install: all
	@install -d $(DESTDIR)/lib
	install -m 444 $(LIB) $(DESTDIR)/lib/$(LIB)

uninstall:
	rm -f $(DESTDIR)/lib/$(LIB)

clean:
	rm -rf *.o *.a

$(LIB): $(OBJS)
	@rm -f $@
	ar cr $@ $(OBJS)
###
char.o: char.c *.h
exit.o: exit.c *.h
flush.o: flush.c *.h
lchar.o: lchar.c *.h
nread.o: nread.c *.h
nwrite.o: nwrite.c *.h
printd.o: printd.c *.h
printf.o: printf.c *.h
printo.o: printo.c *.h
read.o: read.c *.h
start.o: start.c *.h
write.o: write.c *.h
writeb.o: writeb.c *.h
