# Compile

## Platform
Linux and MacOs is supported in SkyWalking Satellite. Windows is not good supported beacuse some features is not adaptive on the Windows, such as the mmap feature. If you want to run it on the windows platform, please read [the doc](../../FAQ/running_on_windows.md).

The Windows platform does not support plugins list:
1. mmap-queue

## Command
```
make build
```