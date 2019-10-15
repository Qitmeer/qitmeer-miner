ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	cd lib/cuckoo && cargo build --release
	go build -ldflags="-r $(ROOT_DIR)lib/cuckoo/target/release" -o qitmeer-miner main.go

run: build
	./qitmeer-miner
