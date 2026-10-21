package skysender

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func TestNewSMTPRejectsBadFrom(t *testing.T) {
	_, err := NewSMTPSender(
		SMTPConfig{
			Host:        "smtp.example.com",
			Port:        "465",
			Username:    "user",
			Password:    "secret",
			FromAddress: "not-an-address",
			Timeout:     time.Second,
		})
	if err == nil {
		t.Fatal("got nil error, expecting error on invalid address")
	}
}

func TestNewSMTPRejectsEmptyFrom(t *testing.T) {
	_, err := NewSMTPSender(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "465",
		Username: "user",
		Password: "secret",
	})
	if err == nil {
		t.Fatal("got nil error, expecting error on empty from address")
	}
}

func TestNewSMTPRejectsMissingServerAuth(t *testing.T) {
	tests := []struct {
		name string
		cfg  SMTPConfig
	}{
		{"empty host", SMTPConfig{Port: "465", Username: "user", Password: "secret", FromAddress: "noreply@example.com"}},
		{"empty port", SMTPConfig{Host: "smtp.example.com", Username: "user", Password: "secret", FromAddress: "noreply@example.com"}},
		{"empty username", SMTPConfig{Host: "smtp.example.com", Port: "465", Password: "secret", FromAddress: "noreply@example.com"}},
		{"empty password", SMTPConfig{Host: "smtp.example.com", Port: "465", Username: "user", FromAddress: "noreply@example.com"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewSMTPSender(tc.cfg); err == nil {
				t.Fatal("got nil error, expecting error on missing server auth field")
			}
		})
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
