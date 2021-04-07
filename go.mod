module github.com/apache/skywalking-satellite

go 1.14

replace skywalking/network v1.0.0 => ./protocol/gen-codes/skywalking/network

require (
	github.com/Shopify/sarama v1.27.2
	github.com/enriquebris/goconcurrentqueue v0.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.5
	github.com/grandecola/mmap v0.6.0
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	skywalking/network v1.0.0
)
