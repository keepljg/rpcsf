package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"io"
	"os"
	"rpcsf/pkg/db/etcdv3"
	"rpcsf/pkg/util/helper"
)

type DiscoveryService struct {
	client          *etcdv3.Client
	kv              clientv3.KV
	addrs           map[string][]string
	watcher         *etcdv3.Etcdv3Watch
	targetsFilePath string
	ctx             context.Context
	cancelFunc      context.CancelFunc
}

type renameFile struct {
	*os.File
	filename string
}

func (f *renameFile) Close() error {
	if err := f.File.Sync(); err != nil {
		fmt.Println("service.Close.File.Sync err", err)
		return err
	}

	if err := f.File.Close(); err != nil {
		fmt.Println("service.Close.File.Close err", err)
		return err
	}
	return os.Rename(f.File.Name(), f.filename)
}

func create(filename string) (io.WriteCloser, error) {
	tmpFilename := filename + ".tmp"

	f, err := os.Create(tmpFilename)
	if err != nil {
		return nil, err
	}

	rf := &renameFile{
		File:     f,
		filename: filename,
	}
	return rf, nil
}

func NewDiscoveryService(targetsFilePath string, etcdName ...string) *DiscoveryService {
	ctx, cancelFunc := context.WithCancel(context.Background())
	d := &DiscoveryService{
		targetsFilePath: targetsFilePath,
		client:          etcdv3.GetClient(etcdName...),
		watcher:         etcdv3.NewEtcdv3Watch(preKey, etcdv3.GetClient(etcdName...), true),
		ctx:             ctx,
		cancelFunc:      cancelFunc,
	}
	d.targetsReset()
	return d
}

func (d *DiscoveryService) getALL() ([]TargetGroup, error) {
	getResp, err := d.kv.Get(context.Background(), preKey, clientv3.WithPrefix())
	if err != nil {
		fmt.Println("service.GetAll err", err)
		return nil, err
	}
	targetGroups := make([]TargetGroup, 0, len(getResp.Kvs))
	addrs := make(map[string][]string)
	for _, kv := range getResp.Kvs {
		var targetGroup TargetGroup
		if err := json.Unmarshal(kv.Value, &targetGroup); err == nil {
			if targetGroup.Targets != nil {
				targetGroups = append(targetGroups, targetGroup)
				if _, ok := addrs[targetGroup.Labels["jobName"]]; !ok {
					addrs[targetGroup.Labels["jobName"]] = make([]string, 0)
				}
				addrs[targetGroup.Labels["jobName"]] = append(addrs[targetGroup.Labels["jobName"]], targetGroup.Targets[0])
			}
		}
	}
	d.addrs = addrs
	return targetGroups, nil
}

func (d *DiscoveryService) DiscoveryWatcher() {
	go func() {
		changed := d.watcher.Changed()
		for {
			select {
			case c := <-changed:
				var targetGroup *TargetGroup
				var isNeedUpdate bool
				if err := json.Unmarshal(c.Value, targetGroup); err == nil {
					if c.Typ == mvccpb.PUT {
						isNeedUpdate = true
						if addrs, ok := d.addrs[targetGroup.Labels["jobName"]]; ok {
							for _, addr := range addrs {
								if addr == targetGroup.Labels["ip"] {
									isNeedUpdate = false
								}
							}
						}
					}

					if c.Typ == mvccpb.DELETE {
						isNeedUpdate = false
						if c.Typ == mvccpb.PUT {
							if addrs, ok := d.addrs[targetGroup.Labels["jobName"]]; ok {
								for _, addr := range addrs {
									if addr == targetGroup.Labels["ip"] {
										isNeedUpdate = true
									}
								}
							}
						}
					}

					if isNeedUpdate {
						d.targetsReset()
					}
				}
			case <-d.ctx.Done():
				return
			}
		}
	}()
}

func (d *DiscoveryService) targetsReset() {
	fmt.Println("target reset ", helper.Date("Y-m-d H:i:s"))
	targetGroups, err := d.getALL()
	if len(targetGroups) == 0 {
		return
	}
	content, err := json.Marshal(targetGroups)
	if err != nil {
		fmt.Println("service.DiscoveryWatcher err", err)
		return
	}
	f, err := create(d.targetsFilePath)
	if err != nil {
		fmt.Println("service.DiscoveryWatcher err", err)
		return
	}

	_, err = f.Write(content)
	if err != nil {
		fmt.Println("service.DiscoveryWatcher err", err)
		return
	}
	defer f.Close()
}

func (d *DiscoveryService) Close() {
	d.cancelFunc()
}
