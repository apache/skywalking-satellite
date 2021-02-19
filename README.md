Apache SkyWalking Satellite
==========

<img src="http://skywalking.apache.org/assets/logo.svg" alt="Sky Walking logo" height="90px" align="right" />

**SkyWalking Satellite**: A lightweight collector/sidecar could be deployed closing to the target monitored system, to collect metrics, traces, and logs. Also, it provides advanced features, such as, local cache, format transform, sampling.

[![GitHub stars](https://img.shields.io/github/stars/apache/skywalking.svg?style=for-the-badge&label=Stars&logo=github)](https://github.com/apache/skywalking)
[![Twitter Follow](https://img.shields.io/twitter/follow/asfskywalking.svg?style=for-the-badge&label=Follow&logo=twitter)](https://twitter.com/AsfSkyWalking)

# Documentation
- [Official documentation](https://skywalking.apache.org/docs/)
- [Blog](https://skywalking.apache.org/blog/2020-11-25-skywalking-satellite-0.1.0-design/) about the design of Satellite 0.1.0.

NOTICE, SkyWalking Satellite uses [v3 protocols](https://github.com/apache/skywalking/blob/master/docs/en/protocols/README.md). They are incompatible with previous SkyWalking releases before SkyWalking 8.0.

# Performance
- 0.5 core supported 3000 ops throughput with memory queue.
- 0.5 core supported 1500 ops throughput with the memory mapped queue(Ensure data stability).

Please read [the doc](./docs/en/FAQ/performance.md) to get more details.

# Download
Go to the [download page](https://skywalking.apache.org/downloads/) to download all available binaries, including MacOS, Linux and Windows. Due to system compatibility problems, some plugins of SkyWalking Satellite cannot be used in Windows system. Check [the corresponding documentation](./docs/en/setup/plugins) to see whether the plugin is available on Windows.

# Compile
As SkyWalking Satellite is using `Makefile`, compiling the project is as easy as executing a command in the root directory of the project.
```shell script
git clone https://github.com/apache/skywalking-satellite
cd skywalking-satellite
git submodule init
git submodule update
make build
```
If you want to know more details about compiling, please read [the doc](./docs/en/guides/compile/compile.md).


# Commands
|  Commands| Flags   | Description  |
|  ----  | ----  |----  |
| start  | --config FILE | Start Satellite with the configuration FILE. (default: "configs/satellite_config.yaml" or read value from *SATELLITE_CONFIG* env).|
| docs  | --output value | Generate Satellite plugin documentations to the output path. (default: "docs" or read value from *SATELLITE_DOC_PATH* env) |


# Contact Us
* Mail list: **dev@skywalking.apache.org**. Mail to `dev-subscribe@skywalking.apache.org`, follow the reply to subscribe the mail list.
* Join `skywalking` channel at [Apache Slack](http://s.apache.org/slack-invite). If the link is not working, find the latest one at [Apache INFRA WIKI](https://cwiki.apache.org/confluence/display/INFRA/Slack+Guest+Invites).
* Twitter, [ASFSkyWalking](https://twitter.com/ASFSkyWalking)
* QQ Group: 901167865(Recommended), 392443393
* [bilibili B站 视频](https://space.bilibili.com/390683219)

# License
[Apache 2.0 License.](/LICENSE)
