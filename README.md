# Sky Satchel

An email library for Go: build, validate, and send MIME email over SMTP, with optional per-address and global rate limiting.

[![CI](https://github.com/hmcalister/skysatchel/actions/workflows/cicd.yml/badge.svg)](https://github.com/hmcalister/skysatchel/actions/workflows/cicd.yml)
[![codecov](https://codecov.io/gh/hmcalister/skysatchel/branch/main/graph/badge.svg)](https://codecov.io/gh/hmcalister/skysatchel)
[![Go Reference](https://pkg.go.dev/badge/github.com/hmcalister/skysatchel.svg)](https://pkg.go.dev/github.com/hmcalister/skysatchel)

## Features

- Simple email interface in pure Go.
- Support for SMTP sending.
- Extensible interfaces for other providers.
- Plain-text and `multipart/mixed` with base64-encoded attachments.
- Optional per-address and global rate limiting in application level logic.

## Install

```sh
go get github.com/hmcalister/skysatchel
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
	"github.com/hmcalister/skysatchel/pkg/email/skysender"
)

func main() {
	sender, err := skysender.NewSMTPSender(skysender.SMTPConfig{
		Host:               "smtp.example.com",
		Port:               "465",
		Username:           "user",
		Password:           "secret",
		FromAddress:        "noreply@example.com",
		FromName:           "Example",
		Timeout:            10 * time.Second,
		MaxAttachmentBytes: 10 << 20,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = sender.Send(context.Background(), skysender.SendParams{
		Message: skysatchel.Message{
			Subject: "Hello",
			Body:    "Sent with Sky Satchel.",
			Attachments: []skysatchel.Attachment{
				{Filename: "report.pdf", Data: pdfBytes},
			},
		},
		Recipients: skysatchel.Recipients{
			ToRecipients: []string{"alice@example.com"},
			CCRecipients: []string{"Bob <bob@example.com>"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent")
}
```


## Rate limiting

Wrap any `skysender.Sender` to cap how often a single address and the sender as a whole can send:

```go
import "github.com/hmcalister/skysatchel/pkg/email/skysender/ratelimit"

limited, err := ratelimit.NewRateLimitedSender(ratelimit.RateLimitedSenderConfig{
	Next:                 sender,
	AddressBurst:         5,
	AddressInterval:      time.Minute,
	GlobalBurst:          100,
	GlobalInterval:       time.Second,
	SweepCadenceDuration: 10 * time.Minute,
})
```

## Error handling

Errors wrap sentinels, so they can be classified with `errors.Is`:

| Sentinel | Meaning |
| --- | --- |
| `skysatchel.ErrInvalidMessage` | The message is unusable (no subject, body, or bad attachment). |
| `skysatchel.ErrAttachmentTooLarge` | An attachment exceeds `MaxAttachmentBytes`. |
| `skysatchel.ErrInvalidRecipient` | A recipient is not a valid address. |
| `skysatchel.ErrNoToRecipients` | The recipient list has no To recipients. |
| `skysatchel.ErrConnection` | The connection failed or was lost. |
| `skysatchel.ErrAuthentication` | The server rejected the credentials. |
| `skysatchel.ErrDelivery` | The server refused the message. |

When only some recipients are rejected, the failure includes a `*skysatchel.RecipientError` per rejected address, and the message is still delivered to the rest.

## Security notes

- Field values containing line breaks (subject, filenames, content types, addresses) are rejected to prevent header injection.
- Delivery uses implicit TLS only; plaintext and STARTTLS are not supported.
- BCC recipients never appear in message headers.
