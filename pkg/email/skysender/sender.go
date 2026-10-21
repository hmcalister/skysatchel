package skysender

import (
	"context"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

type SendParams struct {
	skysatchel.Message
	skysatchel.Recipients
}

// Sender sends a message to a list of recipients.
// Implementations must be safe for concurrent use.
type Sender interface {
	Send(ctx context.Context, params SendParams) error
}
