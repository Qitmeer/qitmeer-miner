#!/usr/bin/env bash
set -ex
export GO111MODULE=on
./installLibrary.sh
export LD_LIBRARY_PATH=`pwd`/lib/cuckoo/target/x86_64-unknown-linux-musl/release:`pwd`/lib/opencl/linux:$LD_LIBRARY_PATH
echo $LD_LIBRARY_PATH
sudo cp `pwd`/lib/opencl/linux/libOpenCL.a /usr/lib/x86_64-linux-musl/

go mod tidy

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
CGO_ENABLED=1 CGO_ENABLED=1 CC=musl-gcc CXX=g++ GOOS=linux GOARCH=amd64 go build -o linux-miner main.go
echo -e "\n Success!"


