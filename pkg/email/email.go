package skysatchel

import (
	"errors"
	"mime"
	"net/mail"
	"path/filepath"
	"strings"
)

// Recipients of all forms
type Recipients struct {
	ToRecipients  []string
	CCRecipients  []string
	BCCRecipients []string
}

// Message contains the content of an email.
type Message struct {
	// Subject line of the email.
	Subject string
	// Body of the email as a raw string. This may be empty.
	// New lines are converted to CRLF.
	Body string
	// Attachments is an optional list of files to attach.
	// A nil or empty slice sends a plain text message.
	Attachments []Attachment
}

// Attachment is a single file attached to a Message.
type Attachment struct {
	// Filename is the name shown to the recipient (e.g. "report.pdf").
	Filename string
	// ContentType is the MIME type of the data (e.g. "application/pdf").
	// When empty, buildMIME infers it from Filename.
	// Falls back to application/octet-stream.
	ContentType string
	// Data is the raw (unencoded) content of the file.
	Data []byte
}

// Remove tags like recipient+tag@email.com
func NormalizeEmail(address string) string {
	address = strings.ToLower(strings.TrimSpace(address))
	if at := strings.LastIndexByte(address, '@'); at > 0 {
		if plus := strings.IndexByte(address[:at], '+'); plus > 0 {
			address = address[:plus] + address[at:]
		}
	}
	return address
}

// Parse a single address and return the bare address.
func ParseAddress(address string) (string, error) {
	// Line breaks in a header value allow header injection.
	if strings.ContainsAny(address, "\r\n") {
		return "", errors.New("email address contains a line break")
	}
	a, err := mail.ParseAddress(address)
	if err != nil {
		return "", err
	}

	return a.Address, nil
}

// Returns att.ContentType, inferring it from the filename
// (or falling back to application/octet-stream) when it is empty.
func AttachmentContentType(att Attachment) string {
	if strings.TrimSpace(att.ContentType) != "" {
		return att.ContentType
	}
	if ct := mime.TypeByExtension(filepath.Ext(att.Filename)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
