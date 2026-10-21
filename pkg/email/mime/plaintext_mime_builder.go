package skymime

import (
	"strings"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

var _ mimeBuilder = plaintextMimeBuilder{}

type plaintextMimeBuilder struct{}

func (builder plaintextMimeBuilder) buildEmail(commonHeaders []string, msg skysatchel.Message) ([]byte, error) {
	body := normalizeCRLF(msg.Body)
	return []byte(strings.Join(commonHeaders, "\r\n") + "\r\n\r\n" + body), nil
}
