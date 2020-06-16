package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"runtime"
	"time"
)

func (c *Client) recovery() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				pl := fmt.Sprintf("grpc client panic: %v\n%v\n%v\n%s\n", req, reply, rerr, buf)
				fmt.Fprintf(os.Stderr, pl)
			}
		}()
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

func (c *Client) Logger() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		startTime := time.Now()
		err = invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(startTime).Seconds()
		fmt.Println(fmt.Sprintf("[RPCSF] CALL %v | %s Spend %f s", startTime.Format("2006/01/02 - 15:04:05"), method, duration))
		return err
	}
}
