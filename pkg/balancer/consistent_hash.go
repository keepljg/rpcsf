package balancer

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"rpcsf/pkg/util/consistent_hash"
	"sync"
)

const (
	ConsistentHasBalanceName = "ConsistentHash"
	ConsistentHashKey        = "ConsistentHashKey"
)

func init() {
	balancer.Register(newConsistentHashBalanceBuilder())
}

func newConsistentHashBalanceBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(ConsistentHasBalanceName, &ConsistentHashPickerBuilder{}, base.Config{
		HealthCheck: true,
	})
}

type ConsistentHashPickerBuilder struct {
}

func (*ConsistentHashPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	hash := consistent_hash.NewHash(-1, nil)
	subConns := make(map[string]balancer.SubConn)
	for address, subConn := range readySCs {
		hash.Add(address.Addr)
		subConns[address.Addr] = subConn
	}
	return &ConsistentHashPicker{
		hash:     hash,
		subConns: subConns,
		mu:       sync.Mutex{},
	}
}

type ConsistentHashPicker struct {
	hash     *consistent_hash.Hash
	subConns map[string]balancer.SubConn
	mu       sync.Mutex
}

func (c *ConsistentHashPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	consistentHashKey := ctx.Value(ConsistentHashKey)
	consistentHashKeyStr, ok := consistentHashKey.(string)
	if !ok {
		return nil, nil, ErrConsistentHashKeyError
	}
	c.mu.Lock()
	addr := c.hash.Get(consistentHashKeyStr)
	if subConn, ok := c.subConns[addr]; ok {
		return subConn, nil, nil
	}
	return nil, nil, ErrNoSubConnSelect
}
