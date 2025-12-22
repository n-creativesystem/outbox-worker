package outbox

import "time"

type ClientOption func(opt *Client)

func WithTimeNow(timeNow func() time.Time) ClientOption {
	return func(opt *Client) {
		opt.timeNow = timeNow
	}
}
