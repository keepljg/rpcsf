package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	*grpc.ClientConn
	handlers []grpc.UnaryClientInterceptor // 中间件 截断器
	config   *GrpcClientConfig
	opt      []grpc.DialOption
}

func NewClient(config *GrpcClientConfig) *Client {
	c := new(Client)
	c.config = config
	if config.Block {
		c.UseOpt(grpc.WithBlock())
	}

	c.UseOpt(grpc.WithInsecure(), grpc.WithKeepaliveParams(
		keepalive.ClientParameters{
			Time:                config.Time,
			Timeout:             config.Timeout,
			PermitWithoutStream: config.PermitWithoutStream,
		}))

	if config.BalanceName != "" {
		c.UseOpt(grpc.WithBalancerName(config.BalanceName))
	}
	return c
}

func DefaultClient(config *GrpcClientConfig) *Client {
	c := NewClient(config)
	c.handlers = append(c.handlers, c.recovery(), c.Logger())
	return c
}

func (c *Client) Use(handlers ...grpc.UnaryClientInterceptor) {
	c.handlers = append(c.handlers, handlers...)
}

func (c *Client) UseOpt(opts ...grpc.DialOption) {
	c.opt = append(c.opt, opts...)
}

func (c *Client) Dial() *grpc.ClientConn {
	c.UseOpt(grpc.WithChainUnaryInterceptor(c.handlers...))
	var (
		ctx        context.Context
		cancelFunc context.CancelFunc
	)
	if c.config.Block {
		if c.config.DialTimeOut > 0 {
			ctx, cancelFunc = context.WithTimeout(context.Background(), c.config.DialTimeOut)
			defer cancelFunc()
		}
	}

	cc, err := grpc.DialContext(ctx, c.config.Address, c.opt...)
	if err != nil {
		panic(err)
	}
	return cc
}
