package dialers

import (
	"context"
	"time"
)

const (
	DefaultTimeout         = 20 * time.Second
	DefaultCleanupDNSEvery = 5 * time.Minute
)

type Opts struct {
	Context         context.Context
	Timeout         time.Duration
	CleanupDNSEvery time.Duration
	TLSSkipVerify   bool
}

func (o *Opts) GetContext() context.Context {
	if o.Context == nil {
		return context.Background()
	}

	return o.Context
}

func (o *Opts) GetTimeout() time.Duration {
	if o.Timeout == 0 {
		return DefaultTimeout
	}

	return o.Timeout
}

func (o *Opts) GetCleanupDNSEvery() time.Duration {
	if o.CleanupDNSEvery == 0 {
		return DefaultCleanupDNSEvery
	}

	return o.CleanupDNSEvery
}

func (o *Opts) GetTLSSkipVerify() bool {
	return o.TLSSkipVerify
}
