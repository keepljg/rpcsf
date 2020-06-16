package register

import (
	"context"
	"rpcsf/pkg/server"
)

type Registry interface {
	RegisterService(context.Context, *server.ServerInfo) error
	DeregisterService(context.Context, *server.ServerInfo) error
}

type NoR struct {
}

func (*NoR) RegisterService(context.Context, *server.ServerInfo) error {
	return nil
}

func (*NoR) DeregisterService(context.Context, *server.ServerInfo) error {
	return nil
}
