#!/usr/bin/env bash

rm -rf lib/cuckoo/target/*
rm -rf lib/opencl/*
rm -rf libOpenCL.zip libcuckoo.zip

wget -O libcuckoo.zip https://github.com/Qitmeer/cuckoo-lib/releases/download/v0.0.1/libcuckoo.zip
unzip libcuckoo.zip -d lib/cuckoo/target/
wget -O libOpenCL.zip https://github.com/Qitmeer/OpenCL-ICD-Loader/releases/download/v0.0.1/libopencl.zip
unzip libOpenCL.zip -d lib/opencl/