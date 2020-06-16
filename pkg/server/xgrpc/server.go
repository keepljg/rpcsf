package xgrpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"rpcsf/pkg/server"
)

type Server struct {
	*grpc.Server
	listener net.Listener
	handlers []grpc.UnaryServerInterceptor // 中间件 截断器
	config   *GrpcServerConfigs
}

type ServerOptions func(s *Server)

func (s *Server) Serve() {
	s.Server.Serve(s.listener)
	return
}

func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

func (s *Server) ServerInfo() *server.ServerInfo {
	s.Server.GracefulStop()
	return nil
}

func NewServer(config *GrpcServerConfigs, options ...grpc.ServerOption) *Server {
	var err error
	if config == nil {
		config = DefaultConfig()
	}
	SetConfig(config)
	keepParam := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     config.IdleTimeout,
		MaxConnectionAgeGrace: config.ForceCloseWait,
		Time:                  config.KeepAliveInterval,
		Timeout:               config.KeepAliveTimeout,
		MaxConnectionAge:      config.MaxLifeTime,
	})
	s := new(Server)
	s.config = config
	options = append(options, keepParam, grpc.UnaryInterceptor(s.interceptor))
	s.Server = grpc.NewServer(options...)
	s.listener, err = net.Listen(config.Network, config.Address)
	if err != nil {
		panic(err)
	}
	return s
}

func DefaultServer(config *GrpcServerConfigs, options ...grpc.ServerOption) *Server {
	server := NewServer(config, options...)
	server.Use(server.Recovery(), server.Logger(), server.Timeout())
	return server
}

func (s *Server) WithUnaryServerInterceptor(handlers ...grpc.UnaryServerInterceptor) {
	s.handlers = handlers
}

func (s *Server) Use(handlers ...grpc.UnaryServerInterceptor) { // 添加截断器
	s.handlers = append(s.handlers, handlers...)
}

// xgrpc 只能使用一个截断器， 使用递归的方法 可以传入多个中间件
func (s *Server) interceptor(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		i     int
		chain grpc.UnaryHandler
	)

	n := len(s.handlers)
	if n == 0 {
		return handler(ctx, req)
	}

	chain = func(ic context.Context, ir interface{}) (interface{}, error) {
		if i == n-1 {
			return handler(ic, ir)
		}
		i++
		return s.handlers[i](ic, ir, args, chain)
	}

	return s.handlers[0](ctx, req, args, chain) // 进行递归调用 中间截断器
}
