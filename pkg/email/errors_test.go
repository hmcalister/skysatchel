package skysatchel_test

import (
	"errors"
	"fmt"
	"net/textproto"
	"strings"
	"testing"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func TestRecipientError(t *testing.T) {
	err := &skysatchel.RecipientError{Recipient: "bob@example.com", Err: skysatchel.ErrDelivery}

	if msg := err.Error(); !strings.Contains(msg, "bob@example.com") {
		t.Errorf("error message = %q, want it to mention the recipient", msg)
	}
	if !errors.Is(err, skysatchel.ErrDelivery) {
		t.Error("RecipientError should unwrap to its cause")
	}
}

func TestRecipientErrorsInJoin(t *testing.T) {
	joined := errors.Join(
		&skysatchel.RecipientError{Recipient: "a@example.com", Err: skysatchel.ErrDelivery},
		&skysatchel.RecipientError{Recipient: "b@example.com", Err: skysatchel.ErrConnection},
	)

	if !errors.Is(joined, skysatchel.ErrDelivery) {
		t.Error("joined error should match ErrDelivery")
	}
	if !errors.Is(joined, skysatchel.ErrConnection) {
		t.Error("joined error should match ErrConnection")
	}
	var recErr *skysatchel.RecipientError
	if !errors.As(joined, &recErr) {
		t.Fatal("joined error should contain a RecipientError")
	}
}

func TestWrapSMTPError(t *testing.T) {
	serverResponse := &textproto.Error{Code: 550, Msg: "mailbox unavailable"}
	if err := skysatchel.WrapSMTPError(serverResponse); !errors.Is(err, skysatchel.ErrDelivery) {
		t.Errorf("server response: err = %v, want ErrDelivery", err)
	}
	if err := skysatchel.WrapSMTPError(fmt.Errorf("broken pipe")); !errors.Is(err, skysatchel.ErrConnection) {
		t.Errorf("non-protocol error: err = %v, want ErrConnection", err)
	}

	err := skysatchel.WrapSMTPError(serverResponse)
	var smtpErr *textproto.Error
	if !errors.As(err, &smtpErr) || smtpErr.Code != 550 {
		t.Errorf("wrapped error should preserve the *textproto.Error, got %v", err)
	}
}
