module github.com/apache/skywalking-satellite

go 1.14

require (
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/grandecola/mmap v0.6.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	google.golang.org/protobuf v1.25.0
	skywalking/network v1.0.0
)

replace skywalking/network v1.0.0 => ./protocol/gen-codes/skywalking/network
