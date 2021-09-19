Changes by Version
==================
Release Notes.

0.2.0
------------------
#### Features
* Set MAXPROCS according to real cpu quota.
* Update golangci-lint version to 1.39.0.
* Update protoc-gen-go version to 1.26.0.
* Add prometheus-metrics-fetcher plugin.
* Add grpc client plugin.
* Add nativelog-grpc-forwarder plugin.
* Add meter-grpc-forwarder plugin.
* Support native management protocol.
* Support native tracing protocol.
* Support native profile protocol.
* Support native CDS protocol.
* Support native JVM protocol.
* Support native Meter protocol.
* Support native Event protocol.
* Support native protocols E2E testing.
* Add Prometheus service discovery in Kubernetes.

#### Bug Fixes
* Fix the data race in mmap queue.
* Fix channel blocking in sender module.
* Fix `pipes.sender.min_flush_events` config could not support min number.
* Remove service name and instance name labels from Prometheus fetcher.

#### Issues and PR
- All issues are [here](https://github.com/apache/skywalking/milestone/80?closed=1)
- All and pull requests are [here](https://github.com/apache/skywalking-satellite/pulls?q=is%3Apr+milestone%3A0.2.0+is%3Aclosed)