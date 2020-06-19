package jaeger

import (
	"github.com/uber/jaeger-client-go/config"
	"rpcsf/pkg/trace"
	"time"
)

type Config struct {
	ServiceName     string
	SamplerType     string
	Param           float64
	JaegerAgentAddr string
}

func ReadConfig(name string) *Config {
	// todo 远程读取配置 或本地
	return &Config{}
}

func initJaeger() {
	conf := ReadConfig("jaeger")
	jaegerConf := config.Configuration{
		ServiceName: conf.ServiceName,
		Disabled:    true,
		RPCMetrics:  true,
		Sampler: &config.SamplerConfig{
			Type:  conf.SamplerType,
			Param: conf.Param,
		},
		Reporter: &config.ReporterConfig{
			BufferFlushInterval: time.Second * 1,
			LogSpans:            false,
			LocalAgentHostPort:  conf.JaegerAgentAddr,
		},
	}
	tracer, _, err := jaegerConf.NewTracer()
	if err != nil {
		panic(err)
	}
	trace.SetGlobalTracer(tracer)
}
