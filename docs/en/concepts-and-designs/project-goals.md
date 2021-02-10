# Design Goals
The document outlines the core design goals for SkyWalking Satellite project.

- **Light Weight**. SkyWalking Satellite has a limited cost for resources and high-performance because of the requirements of the sidecar deployment model.

- **Pluggable**. SkyWalking Satellite core team provides many default implementations, but definitely it is not enough,
and also don't fit every scenario. So, we provide a lot of features for being pluggable. 

- **Portability**.  SkyWalking Satellite can run in multiple environments, including: 
    - Use traditional deployment as a demon process to collect data.
    - Use cloud services as a sidecar, such as in the kubernetes platform.

- **Interop**.  Observability is a big landscape, SkyWalking is impossible to support all, even by its community. So SkyWalking Satellite is compatible with many protocols, including: 
    - SkyWalking protocol
    - (WIP) Prometheus protocol.
