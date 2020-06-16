package main

import (
	"google.golang.org/grpc"
	"rpcsf/pkg/server/xgrpc"
	"time"
)

func main() {
	//wg := new(sync.WaitGroup)
	//wg.Add(2)
	//go defalutServer()
	//go defineServer()
	//wg.Wait()
	defineServer()
}

func defalutServer() {
	server := xgrpc.DefaultServer(nil)
	server.Serve()
}

func defineServer() {
	server := xgrpc.NewServer(&xgrpc.GrpcServerConfigs{
		Address: "127.0.0.1:9998",
		Network: "tcp4",
		Timeout: time.Second * 5,
	}, grpc.InitialWindowSize(100))
	server.Use(server.Logger())
	server.Serve()
}
