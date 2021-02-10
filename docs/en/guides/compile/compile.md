# Compiling

## Platform
Linux and MacOS are supported in SkyWalking Satellite. Windows is not due to following components. 

1. mmap-queue

## Command
```shell script
git clone https://github.com/apache/skywalking-satellite
cd skywalking-satellite
git submodule init
git submodule update
make build
```