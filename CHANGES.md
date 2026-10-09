Changes by Version
==================
Release Notes.

1.4.0
------------------
#### Features
* Support pprof protocol.
* Update Go to `1.26`.

#### Bug Fixes
* Upgrade the base image to `alpine:3.23` and pin openssl/musl packages to fix CVEs.
* Bump Go toolchain, gRPC, `golang.org/x/*` and Prometheus libraries to fix CVEs.
  (e.g. CVE-2026-33186, CVE-2026-42505, CVE-2026-39822, CVE-2026-33814).

#### Issues and PR
- All and pull requests are [here](https://github.com/apache/skywalking-satellite/pulls?q=is%3Apr+milestone%3A1.4.0+is%3Aclosed)
