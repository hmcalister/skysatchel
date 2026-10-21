package skysender

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hmcalister/skysatchel/internal/skymailer"
	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func TestNewSMTPRejectsBadFrom(t *testing.T) {
	mailerServerAuth, _ := skymailer.NewServerAuth(
		"smtp.example.com",
		"465",
		"user",
		"secret",
	)
	_, err := NewSMTPSender(
		SMTPConfig{
			ServerAuth:  mailerServerAuth,
			FromAddress: "not-an-address",
			Timeout:     time.Second,
		})
	if err == nil {
		t.Fatal("got nil error, expecting error on invalid address")
	}
}

func TestNewSMTPRejectsEmptyFrom(t *testing.T) {
	mailerServerAuth, _ := skymailer.NewServerAuth(
		"smtp.example.com",
		"465",
		"user",
		"secret",
	)
	_, err := NewSMTPSender(SMTPConfig{ServerAuth: mailerServerAuth})
	if err == nil {
		t.Fatal("got nil error, expecting error on empty from address")
	}
}

func TestSendRejectsOversizedAttachment(t *testing.T) {
	mock := &mockMailer{}
	sender := newTestSender(t, mock)
	sender.maxAttachmentBytes = 4

	err := sender.Send(context.Background(), SendParams{
		Message: skysatchel.Message{
			Subject:     "s",
			Body:        "x",
			Attachments: []skysatchel.Attachment{{Filename: "a.bin", Data: []byte("too long")}},
		},
		Recipients: skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}},
	})
	if !errors.Is(err, skysatchel.ErrAttachmentTooLarge) {
		t.Errorf("err = %v, want ErrAttachmentTooLarge", err)
	}
	if mock.calls != 0 {
		t.Errorf("mailer called %d times for oversized attachment", mock.calls)
	}
}

func TestSendAllowsAttachmentAtLimit(t *testing.T) {
	mock := &mockMailer{}
	sender := newTestSender(t, mock)
	sender.maxAttachmentBytes = 4

	err := sender.Send(context.Background(), SendParams{
		Message: skysatchel.Message{
			Subject:     "s",
			Body:        "x",
			Attachments: []skysatchel.Attachment{{Filename: "a.bin", Data: []byte("four")}},
		},
		Recipients: skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}},
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if mock.calls != 1 {
		t.Errorf("mailer called %d times, want 1", mock.calls)
	}
}

func TestSendUnlimitedAttachmentByDefault(t *testing.T) {
	mock := &mockMailer{}
	sender := newTestSender(t, mock)

	err := sender.Send(context.Background(), SendParams{
		Message: skysatchel.Message{
			Subject:     "s",
			Body:        "x",
			Attachments: []skysatchel.Attachment{{Filename: "a.bin", Data: bytes.Repeat([]byte("x"), 1024)}},
		},
		Recipients: skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}},
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}
