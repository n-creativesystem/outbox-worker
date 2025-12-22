package aws

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Endpoint struct {
	endpoint map[string]aws.Endpoint
	mu       sync.RWMutex
}

var (
	_ aws.EndpointResolverWithOptions = (*Endpoint)(nil)
)

func NewEndpoint(opts ...Option) *Endpoint {
	e := &Endpoint{
		endpoint: make(map[string]aws.Endpoint),
		mu:       sync.RWMutex{},
	}
	o := &option{}
	for _, opt := range opts {
		opt.apply(o)
	}
	e.SNSEndpoint(o.snsEndpoint)
	e.SQSEndpoint(o.sqsEndpoint)
	return e
}

func (e *Endpoint) AddEndpoint(service string, endpoint aws.Endpoint) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.endpoint[service] = endpoint
}

func (e *Endpoint) ResolveEndpoint(service, region string, options ...any) (aws.Endpoint, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	v, ok := e.endpoint[service]
	if ok {
		v.SigningRegion = region
		return v, nil
	}
	return aws.Endpoint{}, &aws.EndpointNotFoundError{}
}

func (e *Endpoint) EndpointResolver() config.LoadOptionsFunc {
	return config.WithEndpointResolverWithOptions(e)
}

func (e *Endpoint) SNSEndpoint(endpoint string) {
	if endpoint == "" {
		// awsのデフォルトを使うために何もしない
		return
	}
	e.AddEndpoint(sns.ServiceID, aws.Endpoint{URL: endpoint})
}

func (e *Endpoint) SQSEndpoint(endpoint string) {
	if endpoint == "" {
		// awsのデフォルトを使うために何もしない
		return
	}
	e.AddEndpoint(sqs.ServiceID, aws.Endpoint{URL: endpoint})
}
