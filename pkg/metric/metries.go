package metric

import "rpcsf/pkg/metric/prometheus"

var (
	ServerMetricHandler = prometheus.NewProm().RegisterCounter("grpc_server_handled_total", "", []string{"type", "service", "method", "code"}).RegisterTimer("grpc_server_handled_seconds", "", []string{"type", "service", "method"})

	ClientMetricHandler = prometheus.NewProm().RegisterCounter("grpc_client_handled_total", "", []string{"type", "server", "method", "code"}).RegisterTimer("grpc_client_handled_seconds", "", []string{"type", "server", "method"})
)
