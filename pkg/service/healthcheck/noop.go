package healthcheck

import "context"

type noop struct {
	err error
}

var (
	_ Ping = (*noop)(nil)
)

func NewNoop() Ping {
	return &noop{}
}

func (n *noop) PingContext(_ context.Context) error {
	return n.err
}

func newNoopWithErr(err error) Ping {
	return &noop{
		err: err,
	}
}
