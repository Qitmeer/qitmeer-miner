#!/usr/bin/env bash


if [ ! -f "libcuckoo.zip" ]; then
    rm -rf lib/cuckoo/target/* libcuckoo.zip
    wget -O libcuckoo.zip https://github.com/Qitmeer/cuckoo-lib/releases/download/v0.0.1/libcuckoo.zip
    unzip libcuckoo.zip -d lib/cuckoo/target/
fi
if [ ! -f "cuckoo.dll" ]; then
    rm -rf libcuckoo.zip libcudacuckoo.so libcuckoo.so cuckoo.dll cudacuckoo.dll libcuckoo.dylib
    wget https://github.com/Qitmeer/cuckoo-lib/releases/download/v1.0.0/qitmeer-miner-lib.zip
    unzip qitmeer-miner-lib.zip
fi