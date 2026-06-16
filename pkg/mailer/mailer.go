// Package mailer abstracts outbound email. The dev driver logs messages; a real
// SMTP driver can be swapped in without touching callers.
package mailer

import "github.com/collegeassess/backend/pkg/logger"

// Mailer sends transactional email.
type Mailer interface {
	Send(to, subject, body string) error
}

// LogMailer writes emails to the log — useful for local/dev and tests.
type LogMailer struct{ From string }

func NewLogMailer(from string) *LogMailer { return &LogMailer{From: from} }

func (m *LogMailer) Send(to, subject, body string) error {
	logger.Info("email", "from", m.From, "to", to, "subject", subject, "body", body)
	return nil
}
