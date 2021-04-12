package publicapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	v2 "github.com/exoscale/egoscale/v2/api"
)

const (
	operationStatePending = "pending"
	operationStateSuccess = "success"
	operationStateFailure = "failure"
	operationStateTimeout = "timeout"

	defaultPollingInterval = 3 * time.Second
)

// PollFunc represents a function invoked periodically in a polling loop. It returns a boolean flag
// true if the job is completed or false if polling must continue, and any error that occurred
// during the polling (which interrupts the polling regardless of the boolean flag value).
// Upon successful completion, an interface descring an opaque operation can be returned to the
// caller, which will have to perform type assertion depending on the PollFunc implementation.
type PollFunc func(ctx context.Context) (bool, interface{}, error)

// Poller represents a poller instance.
type Poller struct {
	interval time.Duration
	timeout  time.Duration
}

// NewPoller returns a Poller instance.
func NewPoller() *Poller {
	return &Poller{
		interval: defaultPollingInterval,
	}
}

// WithInterval sets the interval at which the polling function will be executed (default: 3s).
func (p *Poller) WithInterval(interval time.Duration) *Poller {
	if interval > 0 {
		p.interval = interval
	}

	return p
}

// WithTimeout sets the time out value after which the polling routine will be cancelled
// (default: no time out).
func (p *Poller) WithTimeout(timeout time.Duration) *Poller {
	if timeout > 0 {
		p.timeout = timeout
	}

	return p
}

// Poll starts the polling routine, executing the provided polling function at the configured
// polling interval. Upon successful polling, an opaque operation is returned to the caller, which
// actual type has to asserted depending on the PollFunc executed.
func (p *Poller) Poll(ctx context.Context, pf PollFunc) (interface{}, error) {
	if p.timeout > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()
		ctx = ctxWithTimeout
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			done, res, err := pf(ctx)
			if err != nil {
				return nil, err
			}
			if !done {
				continue
			}

			return res, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// OperationPoller returns a PollFunc function which queries the state of the specified job.
// Upon successful job completion, the type of the interface{} returned by the PollFunc is a
// pointer to a Resource object (*Resource).
func (c *ClientWithResponses) OperationPoller(zone string, jobID string) PollFunc {
	return func(ctx context.Context) (bool, interface{}, error) {
		resp, err := c.GetOperationWithResponse(v2.WithZone(ctx, zone), jobID)
		if err != nil {
			return true, nil, err
		}
		if resp.StatusCode() != http.StatusOK {
			return true, nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
		}

		switch *resp.JSON200.State {
		case operationStatePending:
			return false, nil, nil

		case operationStateSuccess:
			return true, resp.JSON200.Reference, nil

		case operationStateFailure:
			return true, nil, errors.New("job failed")

		case operationStateTimeout:
			return true, nil, errors.New("job timed out")

		default:
			return true, nil, fmt.Errorf("unknown job state: %s", *resp.JSON200.State)
		}
	}
}
