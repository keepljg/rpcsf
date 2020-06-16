package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"rpcsf/pkg/db/etcdv3"
	"rpcsf/pkg/server"
	"sync"
)

type etcdv3Resovler struct {
	client     *etcdv3.Client
	watch      *etcdv3.Etcdv3Watch
	addrs      []resolver.Address
	preKey     string
	wg         *sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func newetcdv3Resovler(client *etcdv3.Client, preKey string) *etcdv3Resovler {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &etcdv3Resovler{
		client:     client,
		watch:      etcdv3.NewEtcdv3Watch(preKey, client, true),
		preKey:     preKey,
		wg:         new(sync.WaitGroup),
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

func (r *etcdv3Resovler) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	panic("implement me")
}

func (r *etcdv3Resovler) Scheme() string {
	return "etcd3"
}

func (r *etcdv3Resovler) ResolveNow(resolver.ResolveNowOptions) {
	return
}

func (r *etcdv3Resovler) Close() {
	r.watch.Close()
	r.cancelFunc()
	r.wg.Wait()
}

func (r *etcdv3Resovler) GetAllAddresses() []resolver.Address {
	ret := make([]resolver.Address, 0)

	resp, err := r.client.Get(r.ctx, r.preKey, clientv3.WithPrefix())
	if err == nil {
		addrs := extractAddrs(resp)
		if len(addrs) > 0 {
			for _, addr := range addrs {
				v := addr
				ret = append(ret, resolver.Address{
					Addr: v.Addr,
					//Metadata: &v.Metadata, // todo
				})
			}
		}
	}
	return ret
}

func extractAddrs(resp *clientv3.GetResponse) []server.ServerInfo {
	addrs := make([]server.ServerInfo, 0)

	if resp == nil || resp.Kvs == nil {
		return addrs
	}

	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			nodeData := server.ServerInfo{}
			err := json.Unmarshal(v, &nodeData)
			if err != nil {
				grpclog.Info("Parse node data error:", err)
				continue
			}
			addrs = append(addrs, nodeData)
		}
	}
	return addrs
}

func (r *etcdv3Resovler) copyAddress(in []resolver.Address) []resolver.Address {
	out := make([]resolver.Address, len(in))
	copy(out, in)
	return out
}

func (r *etcdv3Resovler) start(cc resolver.ClientConn) {
	out := r.watcher()
	r.wg.Add(1)
	for address := range out {
		cc.UpdateState(resolver.State{Addresses: address})
	}
}

func (r *etcdv3Resovler) watcher() chan []resolver.Address {
	out := make(chan []resolver.Address, 10)
	changed := r.watch.Changed()
	r.wg.Add(1)
	go func() {
		defer func() {
			close(out)
			close(changed)
		}()
		for {
			select {
			case change, ok := <-changed:
				if !ok {
					return
				}
				switch change.Typ {
				case mvccpb.PUT:
					serverInfo := server.ServerInfo{}
					err := json.Unmarshal([]byte(change.Value), &serverInfo)
					if err != nil {
						fmt.Println(err)
					}
					if r.addAddr(resolver.Address{Addr: serverInfo.Addr}) {
						out <- r.copyAddress(r.addrs)
					}

				case mvccpb.DELETE:
					serverInfo := server.ServerInfo{}
					err := json.Unmarshal([]byte(change.Value), &serverInfo)
					if err != nil {
						fmt.Println(err)
					}
					if r.removeAddr(resolver.Address{Addr: serverInfo.Addr}) {
						out <- r.copyAddress(r.addrs)
					}
				}

			default:

			}
		}
	}()
	return out
}

func (r *etcdv3Resovler) addAddr(addr resolver.Address) bool {
	for _, v := range r.addrs {
		if v == addr {
			return false
		}
	}
	r.addrs = append(r.addrs, addr)
	return true
}

func (r *etcdv3Resovler) removeAddr(addr resolver.Address) bool {
	for index, v := range r.addrs {
		if v == addr {
			r.addrs = append(r.addrs[:index], r.addrs[:index+1]...)
			return true
		}
	}
	return false
}
