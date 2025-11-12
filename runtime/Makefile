#
# Runtime library
#
CC          ?= gcc
AR          ?= ar
INSTALL     ?= install
LIB         = libb.a
DESTDIR     = $(HOME)/.local
CFLAGS      = -O2 -Wall -Wextra -ffreestanding
BUILD_DIR   = build
OBJ_DIR     = $(BUILD_DIR)/objs
LIB_DIR     = $(BUILD_DIR)/lib
TARGET      = $(LIB_DIR)/$(LIB)

SRCS        := $(wildcard *.c)
OBJS        := $(patsubst %.c,$(OBJ_DIR)/%.o,$(SRCS))
DEPS        := $(wildcard *.h)

.PHONY: all install uninstall clean

all: $(TARGET)

$(TARGET): $(OBJS) | $(LIB_DIR)
	$(AR) rcs $@ $(OBJS)

$(OBJ_DIR)/%.o: %.c $(DEPS) | $(OBJ_DIR)
	$(CC) $(CFLAGS) -c $< -o $@

$(OBJ_DIR) $(LIB_DIR):
	@mkdir -p $@

install: all
	@$(INSTALL) -d $(DESTDIR)/lib
	@$(INSTALL) -m 444 $(TARGET) $(DESTDIR)/lib/$(LIB)

uninstall:
	@rm -f $(DESTDIR)/lib/$(LIB)

clean:
	@rm -rf $(BUILD_DIR)
