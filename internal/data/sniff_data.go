package data

import (
	"google.golang.org/grpc"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type SniffData struct {
	*v1.SniffData
	grpc.ClientStream
}
