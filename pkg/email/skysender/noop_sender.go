package skysender

import "context"

var _ Sender = NoOpSender{}

// Perform no delivery. Use in tests only, never in production.
//
// Messages passed to Send are silently discarded.
type NoOpSender struct{}

// Discard the message and report success.
func (NoOpSender) Send(_ context.Context, _ SendParams) error {
	return nil
}
