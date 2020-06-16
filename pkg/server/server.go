package server

import (
	"context"
)

type ServerInfo struct {
	Addr          string
	RegisterDir   string
	Scheme        string
	ServerName    string
	ServerVersion string
}

type Server interface {
	Serve()
	Stop()
	GracefulStop(ctx context.Context) error
	ServerInfo() *ServerInfo
}
