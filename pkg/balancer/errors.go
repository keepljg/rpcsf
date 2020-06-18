package balancer

import "errors"

var (
	ErrNoSubConnSelect  = errors.New("no sub conn select")
	ErrAppointAddrError = errors.New("appoint addr error")

	ErrConsistentHashKeyError = errors.New("consistent hash key error")
)
