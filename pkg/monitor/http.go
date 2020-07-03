package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"rpcsf/pkg/db/etcdv3"
)

var c Config

type Config struct {
	Ip          string
	Port        int
	Ttl         int64
	IsKeepAlive bool
	EtcdName    string
	Runmode     string
	ServerName  string
}

func ReadConfig(name ...string) Config {
	// todo read remote config
	return Config{}
}

var (
	DefaultServeMux = http.NewServeMux()
)

func RegisterPprof() {
	HandleFunc("/debug/pprof/", pprof.Index)
	HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	HandleFunc("/debug/pprof/profile", pprof.Profile)
	HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}

func Handle(pattern string, handler http.Handler) {
	DefaultServeMux.Handle(pattern, handler)
}

func Serve() func() {
	s := &http.Server{Handler: DefaultServeMux, Addr: fmt.Sprintf(":%d", c.Port)}
	go s.ListenAndServe()
	var (
		lease *etcdv3.Lease
		err   error
	)
	if c.IsKeepAlive {
		lease, err = etcdv3.NewLease(etcdv3.GetClient(c.EtcdName), c.Ttl)
		if err != nil {
			panic(err)
		}
		target := fmt.Sprintf("%s:%d", c.Ip, c.Port)
		targetGroup := TargetGroup{
			Targets: []string{target},
			Labels: map[string]string{
				"jobName": c.ServerName,
				"env":     c.Runmode,
				"ip":      c.Ip,
			},
		}
		targetsBytes, err := json.Marshal(targetGroup)
		if err == nil {
			lease.SetKv(preKey+target, string(targetsBytes))
			go lease.KeepAlive()
		}
	}
	return func() {
		s.Close()
		if lease != nil {
			lease.Close()
		}
		return
	}
}
