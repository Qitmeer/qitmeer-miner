ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
UNAME_S := $(shell uname -s)

build:
	cd lib/cuckoo && cargo build --release
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-r $(ROOT_DIR)lib/cuckoo/target/release" -o linux-qitmeer-miner main.go
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-r $(ROOT_DIR)lib/cuckoo/target/release" -o windows-qitmeer-miner.exe main.go

run: build
	./qitmeer-miner -h
