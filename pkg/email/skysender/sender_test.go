package skysender

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/hmcalister/skysatchel/internal/skymailer"
	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

// Records the arguments of the last deliver call.
type mockMailer struct {
	from          string
	allRecipients []string
	body          []byte
	err           error
	calls         int
}

func (m *mockMailer) Deliver(_ context.Context, params skymailer.MailerDeliverParams) error {
	m.calls++
	m.from, m.allRecipients, m.body = params.From, params.AllRecipients, params.Body
	return m.err
}

func newTestSender(t *testing.T, m *mockMailer) *SMTPSender {
	t.Helper()
	sender, _ := NewSMTPSender(SMTPConfig{
		Host:        "smtp.example.com",
		Port:        "465",
		Username:    "user",
		Password:    "secret",
		FromAddress: "noreply@example.com",
		FromName:    "Cloud",
		Timeout:     time.Second,
	})
	sender.mailer = m
	return sender
}

func TestSend(t *testing.T) {
	mock := &mockMailer{}
	sender := newTestSender(t, mock)

	err := sender.Send(context.Background(), SendParams{
		Message: skysatchel.Message{Subject: "Hello, World!", Body: "Hello cloud"},
		Recipients: skysatchel.Recipients{
			ToRecipients:  []string{"recipient1@example.com", "Bob <recipient2@example.com>"},
			CCRecipients:  []string{"recipient3@example.com"},
			BCCRecipients: []string{"recipient4@example.com"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if mock.from != "noreply@example.com" {
		t.Errorf("from = %q", mock.from)
	}
	// The envelope is a single, de-duplicated, bare-address list.
	if got := strings.Join(mock.allRecipients, ","); got != "recipient1@example.com,recipient2@example.com,recipient3@example.com,recipient4@example.com" {
		t.Errorf("allRecipients = %q", got)
	}

	parsed, err := mail.ReadMessage(strings.NewReader(string(mock.body)))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Header.Get("From") != `"Cloud" <noreply@example.com>` {
		t.Errorf("From header = %q", parsed.Header.Get("From"))
	}
	// To and Cc are written to headers; display names are preserved.
	if to := parsed.Header.Get("To"); to != "recipient1@example.com, \"Bob\" <recipient2@example.com>" {
		t.Errorf("To header = %q", to)
	}
	if cc := parsed.Header.Get("Cc"); cc != "recipient3@example.com" {
		t.Errorf("Cc header = %q", cc)
	}
	// BCC recipients must never appear in a header.
	if parsed.Header.Get("Bcc") != "" {
		t.Errorf("unexpected Bcc header: %v", parsed.Header)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(parsed.Header.Get("Subject"))
	if err != nil || subject != "Hello, World!" {
		t.Errorf("Subject = %q (%v)", subject, err)
	}
	if parsed.Header.Get("Message-ID") == "" || parsed.Header.Get("Date") == "" {
		t.Errorf("missing headers: %v", parsed.Header)
	}
	body, _ := io.ReadAll(parsed.Body)
	if strings.TrimSpace(string(body)) != "Hello cloud" {
		t.Errorf("body = %q", body)
	}
	for _, addr := range []string{"recipient4@example.com"} {
		if strings.Contains(string(mock.body), addr) {
			t.Errorf("BCC recipient %q leaked into message", addr)
		}
	}
}

func TestSendWithAttachments(t *testing.T) {
	mock := &mockMailer{}
	sender := newTestSender(t, mock)

	msg := skysatchel.Message{
		Subject: "Report",
		Body:    "See attached.",
		Attachments: []skysatchel.Attachment{
			{Filename: "hello.txt", Data: []byte("hello world")},
			{Filename: "data.bin", ContentType: "application/x-custom", Data: []byte{0x00, 0x01, 0x02, 0xff}},
		},
	}
	if err := sender.Send(context.Background(), SendParams{
		Message:    msg,
		Recipients: skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}},
	}); err != nil {
		t.Fatal(err)
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(mock.body))
	if err != nil {
		t.Fatal(err)
	}

	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/mixed" {
		t.Fatalf("Content-Type = %q, want multipart/mixed", mediaType)
	}

	mr := multipart.NewReader(parsed.Body, params["boundary"])

	// First part is the text body.
	textPart, err := mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if ct := textPart.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("first part Content-Type = %q, want text/plain", ct)
	}
	textBody, _ := io.ReadAll(textPart)
	if strings.TrimSpace(string(textBody)) != "See attached." {
		t.Errorf("text body = %q", textBody)
	}

	want := []struct {
		filename string
		ct       string
		data     []byte
	}{
		{"hello.txt", "text/plain; charset=utf-8", []byte("hello world")},
		{"data.bin", "application/x-custom", []byte{0x00, 0x01, 0x02, 0xff}},
	}
	for i, w := range want {
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("attachment %d: %v", i, err)
		}
		if part.FileName() != w.filename {
			t.Errorf("attachment %d filename = %q, want %q", i, part.FileName(), w.filename)
		}
		if got := part.Header.Get("Content-Disposition"); !strings.HasPrefix(got, "attachment") {
			t.Errorf("attachment %d Content-Disposition = %q", i, got)
		}
		if got := part.Header.Get("Content-Transfer-Encoding"); got != "base64" {
			t.Errorf("attachment %d encoding = %q, want base64", i, got)
		}
		if got := part.Header.Get("Content-Type"); got != w.ct {
			t.Errorf("attachment %d Content-Type = %q, want %q", i, got, w.ct)
		}
		encoded, _ := io.ReadAll(part)
		data, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(encoded)))
		if err != nil {
			t.Fatalf("attachment %d decode: %v", i, err)
		}
		if !bytes.Equal(data, w.data) {
			t.Errorf("attachment %d data = %q, want %q", i, data, w.data)
		}
	}
}

func TestSendInvalidInput(t *testing.T) {
	validTo := skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}}
	tests := []struct {
		name       string
		msg        skysatchel.Message
		recipients skysatchel.Recipients
		want       error
	}{
		{"empty subject", skysatchel.Message{Body: "x"}, validTo, skysatchel.ErrInvalidMessage},
		{"subject injection", skysatchel.Message{Subject: "a\r\nBcc: x@example.com", Body: "x"}, validTo, skysatchel.ErrInvalidMessage},
		{"empty body", skysatchel.Message{Subject: "s"}, validTo, skysatchel.ErrInvalidMessage},
		{"empty attachment filename", skysatchel.Message{Subject: "s", Body: "x", Attachments: []skysatchel.Attachment{{Data: []byte("d")}}}, validTo, skysatchel.ErrInvalidMessage},
		{"attachment filename injection", skysatchel.Message{Subject: "s", Body: "x", Attachments: []skysatchel.Attachment{{Filename: "a\r\nX: y", Data: []byte("d")}}}, validTo, skysatchel.ErrInvalidMessage},
		{"attachment content type injection", skysatchel.Message{Subject: "s", Body: "x", Attachments: []skysatchel.Attachment{{Filename: "a.txt", ContentType: "text/plain\r\nX: y", Data: []byte("d")}}}, validTo, skysatchel.ErrInvalidMessage},
		{"empty attachment data", skysatchel.Message{Subject: "s", Body: "x", Attachments: []skysatchel.Attachment{{Filename: "a.txt"}}}, validTo, skysatchel.ErrInvalidMessage},
		{"no recipients", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{}, skysatchel.ErrNoToRecipients},
		{"no to recipient", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{CCRecipients: []string{"recipient1@example.com"}}, skysatchel.ErrNoToRecipients},
		{"bad recipient", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{ToRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
		{"bad cc recipient", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}, CCRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
		{"bad bcc recipient", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}, BCCRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
		{"recipient injection", skysatchel.Message{Subject: "s", Body: "x"}, skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com\r\nBcc: x@example.com"}}, skysatchel.ErrInvalidRecipient},
	}
	for _, tc := range tests {
		mock := &mockMailer{}
		t.Run(tc.name, func(t *testing.T) {
			sender := newTestSender(t, mock)
			if err := sender.Send(context.Background(), SendParams{Message: tc.msg, Recipients: tc.recipients}); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
		if mock.calls != 0 {
			t.Errorf("mailer called %d times for invalid input", mock.calls)
		}
	}
}

func TestSendPropagatesMailerError(t *testing.T) {
	mock := &mockMailer{err: skysatchel.ErrDelivery}
	sender := newTestSender(t, mock)

	err := sender.Send(context.Background(), SendParams{
		Message:    skysatchel.Message{Subject: "s", Body: "x"},
		Recipients: skysatchel.Recipients{ToRecipients: []string{"recipient1@example.com"}},
	})
	if !errors.Is(err, skysatchel.ErrDelivery) {
		t.Errorf("err = %v, want ErrDelivery", err)
	}
}
