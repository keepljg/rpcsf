package client

import "time"

type GrpcClientConfig struct {
	Address     string
	BalanceName string
	DialTimeOut time.Duration
	Block       bool

	Time                time.Duration
	Timeout             time.Duration
	PermitWithoutStream bool
}

func ReadConfig(name string) *GrpcClientConfig {
	// todo 远程读取配置
	return &GrpcClientConfig{}
}

func DefaultConfig() *GrpcClientConfig {
	return &GrpcClientConfig{
		Address:             "etcd",
		BalanceName:         "",
		DialTimeOut:         time.Second * 10,
		Block:               true,
		Time:                0,
		Timeout:             0,
		PermitWithoutStream: false,
	}
}

