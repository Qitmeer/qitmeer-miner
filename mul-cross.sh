#!/bin/sh

# mac cross compile file
# brew install FiloSottile/musl-cross/musl-cross
# brew install mingw-w64
# rustup target add x86_64-pc-windows-gnu
# rustup target add x86_64-unknown-linux-musl
# rustup target add x86_64-apple-darwin
# compile opencl static library
# windows mkdir lib64 && gendef - /c/Windows/system32/OpenCL.dll> lib64/OpenCL.def
# dlltool -l libopencl.a -d OpenCL.def -k -A
# git clone https://github.com/KhronosGroup/OpenCL-ICD-Loader
# git clone https://github.com/KhronosGroup/OpenCL-Headers
# cmake -DOPENCL_ICD_LOADER_HEADERS_DIR=../inc/OpenCL-Headers/ -DBUILD_SHARED_LIBS=OFF ..
#

wget -O libcuckoo.zip https://github.com/Qitmeer/cuckoo-lib/releases/download/v0.0.1/libcuckoo.zip
unzip libcuckoo.zip -d lib/cuckoo/target/
wget -O libOpenCL.zip https://github.com/Qitmeer/OpenCL-ICD-Loader/releases/download/v0.0.1/libopencl.zip
unzip libOpenCL.zip -d lib/opencl/

rm -rf linux-miner mac-miner win-miner.exe libOpenCL.zip libcuckoo.zip

#### mac miner
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o mac-miner main.go

#### linux miner
CGO_ENABLED=1 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o linux-miner main.go

#### win miner
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o win-miner.exe main.go

cp example.pool.conf pool.conf
cp example.solo.conf solo.conf
zip -r win-miner.zip solo.conf pool.conf win-miner.exe
zip -r mac-miner.zip solo.conf pool.conf mac-miner
zip -r linux-miner.zip solo.conf pool.conf linux-miner

rm -rf pool.conf solo.conf linux-miner mac-miner win-miner.exe