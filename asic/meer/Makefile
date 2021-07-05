SRCS := $(wildcard *.c)
CC := arm-linux-gnueabihf-gcc
OUTPUT_FILE := test_meer
all:
	$(CC) -static -O2 $(SRCS) -o $(OUTPUT_FILE)

clean:
	rm $(OUTPUT_FILE)
