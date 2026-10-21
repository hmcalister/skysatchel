package skysatchel_test

import (
	"fmt"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func ExampleValidateMessage() {
	msg := skysatchel.Message{
		Subject: "Weekly report",
		Body:    "Please find the report attached.",
	}

	if err := skysatchel.ValidateMessage(msg); err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("valid")
	// Output: valid
}

func ExampleValidateRecipients() {
	recipients := skysatchel.Recipients{
		ToRecipients:  []string{"alice@example.com", "Bob <bob@example.com>"},
		CCRecipients:  []string{"carol@example.com"},
		BCCRecipients: []string{"audit@example.com"},
	}

	if err := skysatchel.ValidateRecipients(recipients); err != nil {
		fmt.Println("invalid:", err)
		return
	}
	fmt.Println("valid")
	// Output: valid
}

func ExampleNormalizeEmail() {
	fmt.Println(skysatchel.NormalizeEmail("  Alice+newsletter@Example.COM "))
	// Output: alice@example.com
}

func ExampleParseAddress() {
	addr, err := skysatchel.ParseAddress("Bob <bob@example.com>")
	fmt.Println(addr, err)
	// Output: bob@example.com <nil>
}

func ExampleAttachmentContentType() {
	fmt.Println(skysatchel.AttachmentContentType(skysatchel.Attachment{Filename: "report.pdf"}))
	fmt.Println(skysatchel.AttachmentContentType(skysatchel.Attachment{Filename: "data.unknownext"}))
	// Output:
	// application/pdf
	// application/octet-stream
}
