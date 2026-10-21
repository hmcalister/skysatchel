package skysender

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/hmcalister/skysatchel/internal/skymailer"
	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
	skymime "github.com/hmcalister/skysatchel/pkg/email/mime"
)

var _ Sender = SMTPSender{}

type SMTPConfig struct {
	// SMTP server hostname
	Host string
	// SMTP server port
	Port string
	// SMTP server username
	Username string
	// SMTP server password
	Password string
	// Address to appear as sender of emails
	FromAddress string
	// Human-readable name to appear as sender of emails
	FromName string
	// Maximum time to wait for the mailing server
	Timeout time.Duration
	// Maximum size, in bytes, of a single attachment. Zero means unlimited
	MaxAttachmentBytes int64
}

type SMTPSender struct {
	skymailer.ServerAuth
	fromAddress        mail.Address
	timeout            time.Duration
	maxAttachmentBytes int64
	mailer             skymailer.Mailer
}

// Create a new SMTPSender from cfg. No connection is made until Send is called.
//
// Host, Port, Username, Password, and FromAddress are required and must be non-empty.
func NewSMTPSender(cfg SMTPConfig) (*SMTPSender, error) {
	auth, err := skymailer.NewServerAuth(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}

	if cfg.FromAddress == "" {
		return nil, errors.New("email: SMTPConfig.FromAddress must not be empty")
	}

	addr, err := skysatchel.ParseAddress(cfg.FromAddress)
	if err != nil {
		return nil, err
	}

	return &SMTPSender{
		ServerAuth: auth,
		fromAddress: mail.Address{
			Name:    cfg.FromName,
			Address: addr,
		},
		timeout:            cfg.Timeout,
		maxAttachmentBytes: cfg.MaxAttachmentBytes,
		mailer:             skymailer.SMTPMailer{},
	}, nil
}

// Send implements Sender.
//
// Invalid input is rejected before any network activity.
// A cancelled or expired context is reported as the context's error.
func (s SMTPSender) Send(ctx context.Context, sendParams SendParams) error {
	err := skysatchel.ValidateMessage(sendParams.Message)
	if err != nil {
		return err
	}

	err = skysatchel.ValidateRecipients(sendParams.Recipients)
	if err != nil {
		return err
	}

	if s.maxAttachmentBytes > 0 {
		for i, att := range sendParams.Attachments {
			if int64(len(att.Data)) > s.maxAttachmentBytes {
				return fmt.Errorf("%w: attachment %d is %d bytes, limit is %d", skysatchel.ErrAttachmentTooLarge, i, len(att.Data), s.maxAttachmentBytes)
			}
		}
	}

	body, err := skymime.BuildMIMEMessage(s.fromAddress, sendParams.Recipients, sendParams.Message)
	if err != nil {
		return err
	}

	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	return s.mailer.Deliver(ctx, skymailer.MailerDeliverParams{
		ServerAuth:    s.ServerAuth,
		From:          s.fromAddress.Address,
		AllRecipients: skymailer.EnvelopeRecipients(sendParams.Recipients),
		Body:          body,
	})
}
