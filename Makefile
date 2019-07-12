ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	cd lib/cuckoo && cargo build --release
	cp lib/cuckoo/target/release/libcuckoo.dylib lib/
	go build -ldflags="-r $(ROOT_DIR)lib" main.go

run: build
	./main
