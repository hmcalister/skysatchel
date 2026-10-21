// Package skysender sends email messages.
//
// SMTPSender validates a message, renders it as MIME, and delivers it over SMTP.
// Implementations of Sender must be safe for concurrent use.
package skysender
