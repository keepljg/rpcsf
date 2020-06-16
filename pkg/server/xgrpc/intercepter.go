package xgrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"runtime"
	"time"
)

// recovery
func (s *Server) Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				pl := fmt.Sprintf("xgrpc server panic: %v\n%v\n%s\n", req, rerr, buf)
				fmt.Fprintf(os.Stderr, pl)
				fmt.Println(pl)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

// 统一调用超时
func (s *Server) Timeout() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			cancel context.CancelFunc
		)
		timeOut := s.config.Timeout
		if dl, ok := ctx.Deadline(); ok {
			ctimeout := time.Until(dl)
			if timeOut < ctimeout { // 传入为准
				timeOut = ctimeout
			}
		}
		ctx, cancel = context.WithTimeout(ctx, timeOut)
		defer cancel()
		return handler(ctx, req)
	}
}

func (s *Server) Logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		duration := time.Since(startTime).Seconds()
		resp, err = handler(ctx, req)
		fmt.Println(fmt.Sprintf("[RPCSF] %v | %s Spend %f s", startTime.Format("2006/01/02 - 15:04:05"), info.FullMethod, duration))
		return
	}
}
