# Setup
First and most important thing is, SkyWalking Satellite startup behaviours are driven by configs/satellite_config.yaml. Understanding the setting file will help you to read this document.

## Requirements and default settings

Before you start, you should know that the main purpose of quickstart is to help you obtain a basic configuration for previews/demo. Performance and long-term running are not our goals.

You can use `bin/startup.sh` (or cmd) to start up the satellite with their [default settings](../../../configs/satellite_config.yaml), set out as follows:

- Receive SkyWalking related protocols through grpc(listens on `0.0.0.0/11800`) and transmit them to SkyWalking backend(to `0.0.0.0/11800`).
- Expose Self-Observability telemetry data to Prometheus(listens on `0.0.0.0/1234`)

## Startup script
Startup Script
```shell script
bin/startup.sh 
```

## Examples
You can quickly build your satellite according to the following examples:

1. [Forward Agent Data to OAP Server](examples/forward-agent-data-to-oap-server/README.md)
2. [Transmit Log to Kafka](examples/transmit-log-to-kafka/README.md)
3. [GRPC Load Balance Client](examples/grpc-load-balance-client/README.md)
3. [High Throughput Queue](examples/high-throughput-queue/README.md)

## satellite_config.yaml
The core concept behind this setting file is, SkyWalking Satellite is based on pure modularization design. End user can switch or assemble the collector features by their own requirements.

So, in satellite_config.yaml, there are three parts.
1. [The common configurations](./configuration/common.md).
2. [The sharing plugin configurations](./configuration/sharing-plugins.md).
3. [The pipe plugin configurations](./configuration/pipe-plugins.md).

## Advanced feature document link list
1. [Overriding settings](./configuration/override-settings.md) in satellite_config.yaml is supported
