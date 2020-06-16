package etcdv3

import (
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type Change struct {
	Value []byte
	Typ   mvccpb.Event_EventType
}

type Etcdv3Watch struct {
	preKey              string
	isPre               bool
	lastUpdatedRevision int64
	client              *Client
	cancel              context.CancelFunc
	changed             chan Change
}

func NewEtcdv3Watch(preKey string, client *Client, isPre bool) *Etcdv3Watch {
	e := &Etcdv3Watch{
		preKey:  preKey,
		client:  client,
		changed: make(chan Change),
		isPre:   isPre,
	}
	go e.watch()
	return e
}

func (e *Etcdv3Watch) watch() {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	opts := make([]clientv3.OpOption, 0)
	if e.isPre {
		opts = append(opts, clientv3.WithPrefix())
	}
	ch := e.client.Watch(ctx, e.preKey, append(opts, clientv3.WithCreatedNotify(), clientv3.WithRev(e.lastUpdatedRevision))...)
	for {
		for resp := range ch {
			e.handle(&resp)
		}
		time.Sleep(time.Second)
		if e.lastUpdatedRevision > 0 {
			ch = e.client.Watch(ctx, e.preKey, append(opts, clientv3.WithCreatedNotify(), clientv3.WithRev(e.lastUpdatedRevision))...)
		} else {
			ch = e.client.Watch(ctx, e.preKey, append(opts, clientv3.WithCreatedNotify())...)
		}
	}
}

func (e *Etcdv3Watch) Changed() chan Change {
	return e.changed
}

func (e *Etcdv3Watch) Close() {
	close(e.changed)
	e.cancel()
}

func (e *Etcdv3Watch) handle(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > e.lastUpdatedRevision {
		e.lastUpdatedRevision = resp.CompactRevision
	}
	if resp.Header.GetRevision() > e.lastUpdatedRevision {
		e.lastUpdatedRevision = resp.Header.GetRevision()
	}
	if resp.Err() != nil {
		return
	}
	for _, event := range resp.Events {
		if event.Type == mvccpb.PUT || event.Type == mvccpb.DELETE {
			select {
			case e.changed <- Change{Value: event.Kv.Value, Typ: event.Type}:
			default:
			}
		}
	}
}
