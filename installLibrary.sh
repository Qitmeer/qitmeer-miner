#!/usr/bin/env bash

rm -rf lib/cuckoo/target/*
rm -rf libcuckoo.zip

wget -O libcuckoo.zip https://github.com/Qitmeer/cuckoo-lib/releases/download/v0.0.1/libcuckoo.zip
unzip libcuckoo.zip -d lib/cuckoo/target/