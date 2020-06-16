package etcdv3

import (
	"go.etcd.io/etcd/clientv3"
)

type Client struct {
	*clientv3.Client
}

func newClient(config *Config) *Client {
	conf := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.DialTimeout,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
	}

	client, err := clientv3.New(conf)
	if err != nil {
		panic(err)
	}
	return &Client{client}
}
