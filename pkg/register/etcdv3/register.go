package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"rpcsf/pkg/db/etcdv3"
	"rpcsf/pkg/server"
)

type etcdv3Register struct {
	client *etcdv3.Client
	lease  *etcdv3.Lease
}

func newEtcdv3Register(client *etcdv3.Client, ttl int64) *etcdv3Register {
	lease, err := etcdv3.NewLease(client, ttl)
	if err != nil {
		fmt.Println(err)
	}
	return &etcdv3Register{
		client: client,
		lease:  lease,
	}
}

func (e *etcdv3Register) RegisterService(ctx context.Context, serverInfo *server.ServerInfo) error {
	opOptions := make([]clientv3.OpOption, 0)
	key := fmt.Sprintf("/%s/%s/%s/%s", serverInfo.RegisterDir, serverInfo.Scheme, serverInfo.ServerName, serverInfo.ServerVersion)
	val, err := json.Marshal(serverInfo)
	if err != nil {
		return err
	}
	if e.lease != nil {
		opOptions = append(opOptions, clientv3.WithLease(e.lease.LeaseId()), clientv3.WithSerializable())
		e.lease.SetKv(key, string(val))
		go e.lease.KeepAlive()
	}
	_, err = e.client.Put(ctx, key, string(val), opOptions...)
	return err
}

func (e *etcdv3Register) DeregisterService(ctx context.Context, serverInfo *server.ServerInfo) error {
	key := fmt.Sprintf("/%s/%s/%s/%s", serverInfo.RegisterDir, serverInfo.Scheme, serverInfo.ServerName, serverInfo.ServerVersion)
	e.client.Delete(ctx, key)
	if e.lease != nil {
		e.client.Revoke(ctx, e.lease.LeaseId())
	}
	e.lease.Close()
	return nil
}
