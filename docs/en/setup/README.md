# Setup
First and most important thing is, SkyWalking Satellite startup behaviours are driven by configs/satellite_config.yaml. Understood the setting file will help you to read this document.

## Startup script
The startup script is /bin/skywalking-satellite-{version}-{plateform}-amd64. 
1. Generate docs.
```shell script
./bin/skywalking-satellite-{version}-{plateform}-amd64 docs --output=./docs
```
2. Start SkyWalking Satellite.
```shell script
./bin/skywalking-satellite-{version}-{plateform}-amd64 start --config=./configs/satellite_config.yaml
```
## satellite_config.yaml
The core concept behind this setting file is, SkyWalking Satellite is based on pure modularization design. End user can switch or assemble the collector features by their own requirements.

So, in satellite_config.yaml, there are three parts.
1. [The common configurations](./configuration/common.md).
2. [The sharing plugin configurations](./configuration/sharing-plugins.md).
3. [The pipe plugin configurations](./configuration/pipe-plugins.md).

## Advanced feature document link list
1. [Overriding settings](./configuration/override-settings.md) in satellite_config.yaml is supported