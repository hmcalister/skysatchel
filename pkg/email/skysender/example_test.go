package skysender_test

import (
	"fmt"

	"github.com/hmcalister/skysatchel/internal/skymailer"
	"github.com/hmcalister/skysatchel/pkg/email/skysender"
)

func ExampleNewSMTPSender() {
	auth, err := skymailer.NewServerAuth("smtp.example.com", "465", "user", "secret")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	sender, err := skysender.NewSMTPSender(skysender.SMTPConfig{
		ServerAuth:         auth,
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
