# Overview
SkyWalking Satellite: an open-source agent designed for the cloud-native infrastructures, which provides a low-cost, high-efficient, and more secure way to collect telemetry data, such that Trace Segments, Logs, or Metrics.
 

## Why use SkyWalking Satellite?
Observability is the solution to the complex scenario of cloud-native services. However, we may encounter different telemetry data scenarios, different language services, big data analysis, etc. Satellite provides a unified data collection layer for cloud-native services. 
You can easily use it to connect to the SkyWalking ecosystem and enhance the capacity of SkyWalking. 
There are some enhance features on the following when using Satellite.

1. Provide a unified data collection layer to collect logs, traces, and metrics.
2. Provide a safer local cache to reduce the memory cost of the service.
3. Provide the unified transfer way shields the functional differences in the different language libs, such as MQ.
4. Provides the preprocessing functions to ensure accuracy of the metrics, such as sampling.

## Architecture
SkyWalking Satellite is logically split into three parts: Gatherer, Processor, and Sender.

<img src="https://skywalking.apache.org/blog/2020-11-25-skywalking-satellite-0.1.0-design/Satellite.png"/>

- Gatherer collect data and reformat them for SkyWalking requirements.
- Processor processes the input data to generate the new data for  Observability.
- Sender would transfer the downstream data to the SkyWalking OAP with different protocols.