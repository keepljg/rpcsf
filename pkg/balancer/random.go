package balancer

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math/rand"
)

const (
	RandomPickerBuilderName = "RandomPickerBuilder"
)

func init() {
	balancer.Register(newRandomBuilder())
}

func newRandomBuilder() balancer.Builder{
	return base.NewBalancerBuilderWithConfig(RandomPickerBuilderName, &RandomPickerBuilder{}, base.Config{HealthCheck:true})
}

type RandomPickerBuilder struct {
}

func (*RandomPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	subconns := make([]balancer.SubConn, 0, len(readySCs))
	for _, subconn := range readySCs {
		subconns = append(subconns, subconn)
	}
	return &randomPicker{
		subconns: subconns,
	}
}


type randomPicker struct {
	subconns []balancer.SubConn
}

func (r *randomPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	if len(r.subconns) == 0 {
		return nil, nil, ErrNoSubConnSelect
	}
	if len(r.subconns) == 1 {
		return r.subconns[0], nil, nil
	}
	sc := r.subconns[rand.Intn(len(r.subconns))]
	return sc, nil, nil
}

