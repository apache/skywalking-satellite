module github.com/apache/skywalking-satellite

go 1.14

replace skywalking/network v1.0.0 => ./protocol/gen-codes/skywalking/network

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/grandecola/mmap v0.6.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20201221093633-bc327ba9c2f0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.2.8 // indirect
	skywalking/network v1.0.0
)
