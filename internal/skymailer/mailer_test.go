package skymailer

import (
	"reflect"
	"testing"

	skysatchel "github.com/hmcalister/skysatchel/pkg/email"
)

func TestNewServerAuthRejectsMissingFields(t *testing.T) {
	type cfgData struct {
		host     string
		port     string
		username string
		password string
	}
	tests := []struct {
		name string
		cfg  cfgData
	}{
		{
			"missing host",
			cfgData{
				host:     "",
				port:     "465",
				username: "username",
				password: "secret",
			},
		},
		{
			"missing port",
			cfgData{
				host:     "hostname",
				port:     "",
				username: "username",
				password: "secret",
			},
		},
		{
			"missing username",
			cfgData{
				host:     "hostname",
				port:     "465",
				username: "",
				password: "secret",
			},
		},
		{
			"missing password",
			cfgData{
				host:     "hostname",
				port:     "465",
				username: "username",
				password: "",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewServerAuth(tc.cfg.host, tc.cfg.port, tc.cfg.username, tc.cfg.password)
			if err == nil {
				t.Errorf("expected an error for missing required config %v", tc.name)
			}
		})
	}
}

func TestNewServerAuthStoresFields(t *testing.T) {
	auth, err := NewServerAuth("smtp.example.com", "465", "user", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if auth.host != "smtp.example.com" || auth.port != "465" || auth.username != "user" || auth.password != "secret" {
		t.Errorf("auth = %+v, want the supplied fields", auth)
	}
}

func TestEnvelopeRecipients(t *testing.T) {
	tests := []struct {
		name       string
		recipients skysatchel.Recipients
		want       []string
	}{
		{
			"empty",
			skysatchel.Recipients{},
			nil,
		},
		{
			"all groups",
			skysatchel.Recipients{
				ToRecipients:  []string{"to@example.com"},
				CCRecipients:  []string{"cc@example.com"},
				BCCRecipients: []string{"bcc@example.com"},
			},
			[]string{"to@example.com", "cc@example.com", "bcc@example.com"},
		},
		{
			"strips display names",
			skysatchel.Recipients{ToRecipients: []string{"Bob <bob@example.com>"}},
			[]string{"bob@example.com"},
		},
		{
			"de-duplicates",
			skysatchel.Recipients{
				ToRecipients:  []string{"a@example.com", "a@example.com"},
				CCRecipients:  []string{"A <a@example.com>"},
				BCCRecipients: []string{"b@example.com"},
			},
			[]string{"a@example.com", "b@example.com"},
		},
		{
			"skips invalid addresses",
			skysatchel.Recipients{ToRecipients: []string{"nope", "a@example.com"}},
			[]string{"a@example.com"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EnvelopeRecipients(tc.recipients)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("EnvelopeRecipients = %v, want %v", got, tc.want)
			}
		})
	}
}
