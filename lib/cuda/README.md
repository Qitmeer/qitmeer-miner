# cuda miner
## Only Nvidia Cards Support

# Env

## Ubuntu
```bash
$ sudo apt-get gcc g++
$ sudo apt-get install beignet-dev nvidia-cuda-dev nvidia-cuda-toolkit
```

## Mac
- Install [**Cuda**](https://developer.nvidia.com/cuda-downloads)
- Install [**Command Line Tools**](https://developer.apple.com/downloads/)

## Windows 

- Install [**Build Tools for Visual Studio**](https://visualstudio.microsoft.com/thank-you-downloading-visual-studio/?sku=BuildTools&rel=16)
- Install [**Cuda**](https://developer.nvidia.com/cuda-downloads)
- Install [**Mingw64**](https://mingw-w64.org/)
- Set Environment variables
- `PATH` 
    - D:\VS\2019\VC\Tools\MSVC\14.23.28105\bin\Hostx64\x64
    - C:\mingw64\bin
    - C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.1\bin
    - D:\VS\2019\Common7\IDE
- `INCLUDE` 
    - C:\Program Files (x86)\Windows Kits\10\Include\10.0.18362.0\ucrt
    - C:\Program Files (x86)\Windows Kits\10\Include\10.0.18362.0\um
    - C:\Program Files (x86)\Windows Kits\10\Include\10.0.18362.0\shared
    - C:\mingw64\include
    - D:\VS\2019\VC\Tools\MSVC\14.23.28105\include
- `LIB`
    - D:\VS\2019\VC\Tools\MSVC\14.23.28105\lib\x64
    - C:\Program Files (x86)\Windows Kits\10\Lib\10.0.18362.0\ucrt\x64
    - C:\Program Files (x86)\Windows Kits\10\Lib\10.0.18362.0\um\x64
    
## Compile

```bash
# mac
$  nvcc -m64 -arch=sm_35 -o libcudacuckoo.dylib --shared -Xcompiler -fPIC -DEDGEBITS=29 -DSIPHASH_COMPAT=1 mean.cu ./crypto/blake2b-ref.c
# ubuntu
$  nvcc -m64 -arch=sm_35 -o libcudacuckoo.so --shared -Xcompiler -fPIC -DEDGEBITS=29 -DSIPHASH_COMPAT=1 mean.cu ./crypto/blake2b-ref.c
# windows 
$  nvcc -m64 -arch=sm_35 -o cudacuckoo.dll --shared -Xcompiler -fPIC -DEDGEBITS=29 -DSIPHASH_COMPAT=1 -DISWINDOWS=1 mean.cu ./crypto/blake2b-ref.c
```
