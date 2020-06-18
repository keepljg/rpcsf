package balancer

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"rpcsf/pkg/util/swr"
	"strconv"
	"sync"
)

const (
	SwrBuilderName = "SwrBuilderName"
)

func init() {
	balancer.Register(newSwrBuilder())
}

func newSwrBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(SwrBuilderName, &SwrPickerBuilder{}, base.Config{
		HealthCheck: true,
	})
}

type SwrPickerBuilder struct {
}

func (s *SwrPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	sw := swr.NewSw()
	for address, subconn := range readySCs {
		var weight = 1
		metaData, ok := address.Metadata.(map[string]string)
		if ok {
			w := metaData["weight"]
			weight, _ = strconv.Atoi(w)
			if weight == 0 {
				weight = 1
			}
		}
		sw.Add(subconn, weight)
	}
	return &swrPicker{
		mu: sync.Mutex{},
		sw: sw,
	}
}

type swrPicker struct {
	mu sync.Mutex
	sw *swr.Sw
}

func (s *swrPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	s.mu.Lock()
	val := s.sw.Get()
	s.mu.Unlock()
	subconn, ok := val.(balancer.SubConn)
	if !ok {
		return nil, nil, ErrNoSubConnSelect
	}
	return subconn, nil, nil
}
