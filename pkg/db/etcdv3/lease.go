package etcdv3

import (
	"context"
	"go.etcd.io/etcd/clientv3"
)

// 创建一个租约 并提供keepalive
type Lease struct {
	client  *Client
	leaseId clientv3.LeaseID
	ttl     int64
	key     string
	value   string
	done    chan struct{}
}

func NewLease(client *Client, ttl int64, extra ...string) (*Lease, error) {
	var key, value string
	if len(extra) == 2 {
		key, value = extra[0], extra[1]
	}
	l := &Lease{
		client:  client,
		leaseId: 0,
		ttl:     ttl,
		key:     key,
		value:   value,
		done:    make(chan struct{}),
	}
	err := l.genLeaseId()
	return l, err
}

func (l *Lease) SetKv(key, val string) {
	l.key = key
	l.value = val
}

func (l *Lease) LeaseId() clientv3.LeaseID {
	return l.leaseId
}

func (l *Lease) Close() {
	close(l.done)
}

func (l *Lease) genLeaseId() error {
	leaseGrantResponse, err := l.client.Grant(context.Background(), l.ttl)
	if err != nil {
		return err
	}
	return err
	l.leaseId = leaseGrantResponse.ID
	return nil
}

func (l *Lease) KeepAlive() {
	var leaseGrantResponse *clientv3.LeaseGrantResponse
	for {
		ch, err := l.client.KeepAlive(context.Background(), l.leaseId)
		if err == nil {
			select {
			case ka := <-ch:
				// 续租前过期，重新生成新的租约，重新put
				if ka == nil {
					if l.key != "" && l.value != "" {
						leaseGrantResponse, err = l.client.Grant(context.Background(), l.ttl)
						if err == nil {
							l.leaseId = leaseGrantResponse.ID
							l.client.Put(context.Background(), l.key, l.value, clientv3.WithLease(l.leaseId))
						}
					}
				}
			case <-l.done:
				return
			}
		}
	}
}
