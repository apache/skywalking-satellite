# Compiling

## Go version

Go version `1.23` is required for compilation.

## Platform
Linux, MacOS and Windows are supported in SkyWalking Satellite. However, some components don't fit the Windows platform, including:
1. mmap-queue

## Command
```shell script
git clone https://github.com/apache/skywalking-satellite
cd skywalking-satellite
make build
```
