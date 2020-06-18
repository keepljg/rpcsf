package balancer

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync"
)

const (
	RoundRobinBalanceName = "RoundRobinBalance"
)

func init() {
	balancer.Register(newRandomBuilder())
}

func newRoundRobinBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(RoundRobinBalanceName, &RoundRobinPickerBuilder{}, base.Config{HealthCheck: true})
}

type RoundRobinPickerBuilder struct {
}

func (*RoundRobinPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	subconns := make([]balancer.SubConn, 0, len(readySCs))
	for _, subconn := range readySCs {
		subconns = append(subconns, subconn)
	}
	return &roundRobinPicker{
		subconns: subconns,
		mu:       sync.Mutex{},
	}
}

type roundRobinPicker struct {
	subconns []balancer.SubConn
	next     int
	mu       sync.Mutex
}

func (r *roundRobinPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	l := len(r.subconns)
	if l == 0 {
		return nil, nil, ErrNoSubConnSelect
	}
	if l == 1 {
		return r.subconns[0], nil, nil
	}
	r.mu.Lock()
	sc := r.subconns[r.next]
	r.next += 1
	if r.next >= l {
		r.next = 0
	}
	r.mu.Unlock()
	return sc, nil, nil
}
