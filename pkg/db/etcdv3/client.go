package etcdv3

import (
	"go.etcd.io/etcd/clientv3"
	"sync"
)

var (
	clients = make(map[string]*Client)
	mu      sync.Mutex
)

type Client struct {
	*clientv3.Client
}

func GetClient(name ...string) *Client {
	config := ReadConfig(name...)
	mu.Lock()
	defer mu.Unlock()
	if c, ok := clients[config.Name]; ok {
		return c
	}
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
	c := &Client{client}
	clients[config.Name] = c
	return c
}
