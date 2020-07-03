package xgrpc

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

func (s *Server) Tracer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx) // 提取md
		if !ok {
			md = metadata.Pairs()
		}
		var serverSpan opentracing.Span
		clientContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, trace.MetadataReaderWriter{md}) // 注入md
		if err == nil {
			serverSpan = opentracing.StartSpan(
				"RPC Server "+info.FullMethod, ext.RPCServerOption(clientContext)) // 生成server span
		} else {
			serverSpan = opentracing.StartSpan("RPC Server " + info.FullMethod)
		}
		defer serverSpan.Finish()
		ctx = trace.ContextWithSpan(ctx, serverSpan) // 将span 写入ctx中
		resp, err = handler(ctx, req)
		if err != nil {
			code := codes.Unknown
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			}
			serverSpan.SetTag("code", code)
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}
		return resp, err
	}
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func (s *Server) Metries() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()
		resp, err = handler(ctx, req)
		gst, _ := status.FromError(err)
		server, method := splitMethodName(info.FullMethod)
		metric.ServerMetricHandler.Incr("Unary", server, method, gst.Message())
		metric.ServerMetricHandler.Timing("Unary", time.Now().Sub(startTime).Seconds(), server, method)
		return resp, err
	}
}
