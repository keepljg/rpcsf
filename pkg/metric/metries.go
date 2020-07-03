package metric

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"rpcsf/pkg/metric/prometheus"
	"rpcsf/pkg/monitor"
)

var (
	ServerMetricHandler = prometheus.NewProm().RegisterCounter("grpc_server_handled_total", "", []string{"type", "service", "method", "code"}).RegisterTimer("grpc_server_handled_seconds", "", []string{"type", "service", "method"})

	ClientMetricHandler = prometheus.NewProm().RegisterCounter("grpc_client_handled_total", "", []string{"type", "server", "method", "code"}).RegisterTimer("grpc_client_handled_seconds", "", []string{"type", "server", "method"})
)

func RegisterMontior() {
	monitor.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
}
