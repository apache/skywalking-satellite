# Receiver/http-nativelog-receiver
## Description
This is a receiver for SkyWalking http logging format, which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto.
## Support Forwarders
 - [nativelog-grpc-forwarder](forwarder_nativelog-grpc-forwarder.md)
## DefaultConfig
```yaml
# The native log request URI.
uri: "/logging"
# The request timeout seconds.
timeout: 5
```
