package xgrpc

import "time"

type GrpcServerConfigs struct {
	Address           string
	Network           string
	RegisterDir       string
	ServerName        string
	ServerVersion     string
	Timeout           time.Duration // rcp call 公共超时配置
	IdleTimeout       time.Duration
	MaxLifeTime       time.Duration
	ForceCloseWait    time.Duration
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
}

func ReadConfig(name string) *GrpcServerConfigs {
	// todo 远程读取配置
	return &GrpcServerConfigs{}
}

func DefaultConfig() *GrpcServerConfigs {
	return &GrpcServerConfigs{
		Address:           "127.0.0.1:9999",
		Network:           "tcp",
		Timeout:           time.Second * 2,
		IdleTimeout:       time.Second * 60,
		MaxLifeTime:       time.Hour * 2,
		ForceCloseWait:    time.Second * 20,
		KeepAliveInterval: time.Second * 60,
		KeepAliveTimeout:  time.Second * 20,
	}
}

func SetConfig(config *GrpcServerConfigs) {
	dc := DefaultConfig()
	if config == nil {
		panic("xgrpc server config is nil")
	}
	if config.Address == "" {
		panic("xgrpc server config.Address is empty")
	}
	if config.Network == "" {
		config.Network = dc.Network
	}
	if config.Timeout == 0 {
		config.Timeout = dc.Timeout
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = dc.IdleTimeout
	}
	if config.MaxLifeTime == 0 {
		config.MaxLifeTime = dc.MaxLifeTime
	}
	if config.ForceCloseWait == 0 {
		config.ForceCloseWait = dc.ForceCloseWait
	}
	if config.KeepAliveInterval == 0 {
		config.KeepAliveInterval = dc.KeepAliveInterval
	}
	if config.KeepAliveTimeout == 0 {
		config.KeepAliveTimeout = dc.KeepAliveTimeout
	}
}
