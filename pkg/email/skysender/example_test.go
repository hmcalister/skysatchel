package skysender_test

import (
	"fmt"

	"github.com/hmcalister/skysatchel/pkg/email/skysender"
)

func ExampleNewSMTPSender() {
	sender, err := skysender.NewSMTPSender(skysender.SMTPConfig{
		Host:               "smtp.example.com",
		Port:               "465",
		Username:           "user",
		Password:           "secret",
		FromAddress:        "reports@example.com",
		FromName:           "Reports",
		MaxAttachmentBytes: 10 << 20,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	_ = sender
	fmt.Println("sender ready")
	// Output: sender ready
}
