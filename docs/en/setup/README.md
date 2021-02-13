# Setup
First and most important thing is, SkyWalking Satellite startup behaviours are driven by configs/satellite_config.yaml. Understanding the setting file will help you to read this document.

## Startup script
Startup Script
```shell script
/bin/startup.sh or /bin/startup.bat
```
## satellite_config.yaml
The core concept behind this setting file is, SkyWalking Satellite is based on pure modularization design. End user can switch or assemble the collector features by their own requirements.

So, in satellite_config.yaml, there are three parts.
1. [The common configurations](./configuration/common.md).
2. [The sharing plugin configurations](./configuration/sharing-plugins.md).
3. [The pipe plugin configurations](./configuration/pipe-plugins.md).

## Advanced feature document link list
1. [Overriding settings](./configuration/override-settings.md) in satellite_config.yaml is supported
