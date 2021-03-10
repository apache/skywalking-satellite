module github.com/apache/skywalking-satellite

go 1.14

replace skywalking/network v1.0.0 => ./protocol/gen-codes/skywalking/network

require (
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Shopify/sarama v1.27.2
	github.com/census-instrumentation/opencensus-proto v0.3.0
	github.com/enriquebris/goconcurrentqueue v0.6.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/gophercloud/gophercloud v0.16.0 // indirect
	github.com/grandecola/mmap v0.6.0
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/zap v1.13.0
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	k8s.io/api v0.20.4 // indirect
	k8s.io/apimachinery v0.20.4 // indirect
	skywalking/network v1.0.0
)
