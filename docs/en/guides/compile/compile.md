# Compiling

## Platform
Linux, MacOS and Windows are supported in SkyWalking Satellite. However, some components don't fit the Windows platform, including:
1. mmap-queue

## Command
```shell script
git clone https://github.com/apache/skywalking-satellite
cd skywalking-satellite
git submodule init
git submodule update
make build
```