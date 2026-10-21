package skymime_test

import (
	"bytes"
	"fmt"
	"net/mail"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
	skymime "github.com/hmcalister/skysatchel/pkg/email/mime"
)

func ExampleBuildMIMEMessage() {
	msg := skysatchel.Message{
		Subject: "Weekly report",
		Body:    "See attached.",
		Attachments: []skysatchel.Attachment{
			{Filename: "report.txt", Data: []byte("hello")},
		},
	}
	recipients := skysatchel.Recipients{ToRecipients: []string{"alice@example.com"}}

	rendered, err := skymime.BuildMIMEMessage(mail.Address{
		Name:    "Reports",
		Address: "reports@example.com",
	}, recipients, msg)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(rendered))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(parsed.Header.Get("From"))
	fmt.Println(parsed.Header.Get("To"))
	// Output:
	// "Reports" <reports@example.com>
	// alice@example.com
}
