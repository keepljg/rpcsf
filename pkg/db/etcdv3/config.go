package etcdv3

import "time"

type Config struct {
	Endpoints []string `json:"endpoints"`
	// 连接超时时间
	DialTimeout          time.Duration `json:"dialTimeout"`
	DialKeepAliveTime    time.Duration `json:"dialKeepAliveTime"`
	DialKeepAliveTimeout time.Duration `json:"dialKeepAliveTimeout"`
}

func ReadConfig(name string) *Config {
	// todo
	return &Config{}
}
