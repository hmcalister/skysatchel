package skysatchel_test

import (
	"errors"
	"testing"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func TestValidateMessage(t *testing.T) {
	validAtt := skysatchel.Attachment{Filename: "a.txt", Data: []byte("d")}
	tests := []struct {
		name string
		msg  skysatchel.Message
		want error
	}{
		{"valid", skysatchel.Message{Subject: "s", Body: "b"}, nil},
		{"valid with attachment", skysatchel.Message{Subject: "s", Body: "b", Attachments: []skysatchel.Attachment{validAtt}}, nil},
		{"empty subject", skysatchel.Message{Subject: " ", Body: "b"}, skysatchel.ErrInvalidMessage},
		{"subject line break", skysatchel.Message{Subject: "a\r\nb", Body: "b"}, skysatchel.ErrInvalidMessage},
		{"empty body", skysatchel.Message{Subject: "s"}, skysatchel.ErrInvalidMessage},
		{"invalid attachment", skysatchel.Message{Subject: "s", Body: "b", Attachments: []skysatchel.Attachment{{}}}, skysatchel.ErrInvalidMessage},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := skysatchel.ValidateMessage(tc.msg)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestValidateAttachment(t *testing.T) {
	tests := []struct {
		name    string
		att     skysatchel.Attachment
		wantErr bool
	}{
		{"valid", skysatchel.Attachment{Filename: "a.txt", Data: []byte("d")}, false},
		{"empty filename", skysatchel.Attachment{Data: []byte("d")}, true},
		{"filename line break", skysatchel.Attachment{Filename: "a\nb", Data: []byte("d")}, true},
		{"content type line break", skysatchel.Attachment{Filename: "a.txt", ContentType: "t\r\nX: y", Data: []byte("d")}, true},
		{"empty data", skysatchel.Attachment{Filename: "a.txt"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := skysatchel.ValidateAttachment(tc.att)
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateRecipients(t *testing.T) {
	tests := []struct {
		name       string
		recipients skysatchel.Recipients
		want       error
	}{
		{"valid", skysatchel.Recipients{ToRecipients: []string{"a@example.com"}}, nil},
		{"valid with display name", skysatchel.Recipients{ToRecipients: []string{"Bob <b@example.com>"}}, nil},
		{"no to recipients", skysatchel.Recipients{CCRecipients: []string{"a@example.com"}}, skysatchel.ErrNoToRecipients},
		{"bad to recipient", skysatchel.Recipients{ToRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
		{"bad cc recipient", skysatchel.Recipients{ToRecipients: []string{"a@example.com"}, CCRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
		{"bad bcc recipient", skysatchel.Recipients{ToRecipients: []string{"a@example.com"}, BCCRecipients: []string{"nope"}}, skysatchel.ErrInvalidRecipient},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := skysatchel.ValidateRecipients(tc.recipients)
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"bare", "a@example.com", "a@example.com", false},
		{"display name", "Bob <b@example.com>", "b@example.com", false},
		{"empty", "", "", true},
		{"no at sign", "nope", "", true},
		{"line break", "a@example.com\r\nBcc: x@example.com", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := skysatchel.ParseAddress(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercases", "Alice@Example.COM", "alice@example.com"},
		{"trims space", "  a@example.com  ", "a@example.com"},
		{"removes tag", "a+tag@example.com", "a@example.com"},
		{"removes tag and normalizes", " A+Tag@Example.COM ", "a@example.com"},
		{"spaces inside kept", "a b@example.com", "a b@example.com"},
		{"no tag", "a@example.com", "a@example.com"},
		{"plus in domain kept", "a@ex+ample.com", "a@ex+ample.com"},
		{"no at sign", "nope", "nope"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := skysatchel.NormalizeEmail(tc.input); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Testskysatchel.AttachmentContentTypeInference checks the fallback chain for the MIME
// type of an attachment.
func TestAttachmentContentTypeInference(t *testing.T) {
	tests := []struct {
		name string
		att  skysatchel.Attachment
		want string
	}{
		{"explicit", skysatchel.Attachment{ContentType: "application/pdf"}, "application/pdf"},
		{"from extension", skysatchel.Attachment{Filename: "report.pdf"}, "application/pdf"},
		{"unknown extension", skysatchel.Attachment{Filename: "file.unknownext"}, "application/octet-stream"},
		{"no extension", skysatchel.Attachment{Filename: "Makefile"}, "application/octet-stream"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := skysatchel.AttachmentContentType(tc.att); got != tc.want {
				t.Errorf("attachmentContentType = %q, want %q", got, tc.want)
			}
		})
	}
}
