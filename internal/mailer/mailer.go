package mailer

import "context"

type Mailer interface {
	SendHTML(ctx context.Context, toEmail, toName, subject, htmlContent string) error
	SendVerification(ctx context.Context, toEmail, toName, token string) error
}
