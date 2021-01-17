module github.com/apache/skywalking-satellite

go 1.14

replace skywalking/network v1.0.0 => ./protocol/gen-codes/skywalking/network

require (
	github.com/Shopify/sarama v1.27.2
	github.com/gobuffalo/packr/v2 v2.8.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/grandecola/mmap v0.6.0
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/sys v0.0.0-20210113181707-4bcb84eeeb78 // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/tools v0.0.0-20210115202250-e0d201561e39 // indirect
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	skywalking/network v1.0.0
)
