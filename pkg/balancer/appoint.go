package balancer

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync"
)
const (
	AppointBalanceName = "appoint"
	addrKey = "addr"
)


func init() {
	// 注册到grpc 到balance map中
	balancer.Register(newAppointPickBuilder())
}

func newAppointPickBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(AppointBalanceName, &AppointPickBuilder{}, base.Config{HealthCheck: true})
}

type AppointPickBuilder struct {
}

func (*AppointPickBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var subConns = make(map[string]balancer.SubConn)
	for address, subConn := range readySCs {
		subConns[address.Addr] = subConn
	}
	return &appointPicker{
		subConns: subConns,
		mu:       sync.Mutex{},
	}
}


type appointPicker struct {
	subConns map[string]balancer.SubConn
	mu sync.Mutex
}

func (a *appointPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	addr := ctx.Value(addrKey)
	addrStr, ok := addr.(string)
	if !ok {
		return nil, nil, ErrAppointAddrError
	}
	a.mu.Lock()
	subConn, ok := a.subConns[addrStr]
	a.mu.Unlock()
	if !ok {
		return nil, nil, ErrNoSubConnSelect
	}
	return subConn, nil, nil
}



