package aws

type option struct {
	snsEndpoint string
	sqsEndpoint string
}

type Option interface {
	apply(o *option)
}

type optionFn func(o *option)

func (fn optionFn) apply(o *option) {
	fn(o)
}

func WithSNSEndpoint(endpoint string) Option {
	return optionFn(func(o *option) {
		o.snsEndpoint = endpoint
	})
}

func WithSQSEndpoint(endpoint string) Option {
	return optionFn(func(o *option) {
		o.sqsEndpoint = endpoint
	})
}
