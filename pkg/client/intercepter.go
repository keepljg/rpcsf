package client

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"os"
	"rpcsf/pkg/metric"
	"rpcsf/pkg/trace"
	"runtime"
	"strings"
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

func (c *Client) Tracer() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, ctx := trace.StartSpanFromContext(ctx, "RPC Client"+method, opentracing.Tag{Key: string(ext.Component), Value: "grpc"}, ext.SpanKindRPCClient) //创建一个span
		defer span.Finish()
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, trace.MetadataReaderWriter{md}) // 将md注入
		if err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}
		newCtx := metadata.NewOutgoingContext(ctx, md) // 创建新的ctx
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			code := codes.Unknown
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			}
			span.SetTag("response_code", code)
			ext.Error.Set(span, true)

			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		return err
	}
}


func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func (c *Client) Metric() grpc.UnaryClientInterceptor{
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		gst, _ := status.FromError(err)
		server, method := splitMethodName(method)
		metric.ClientMetricHandler.Incr("Unary", server, method, gst.Message())
		metric.ClientMetricHandler.Timing("Unary", time.Now().Sub(startTime).Seconds(),server,  method, gst.Message())
		return err
	}
}
