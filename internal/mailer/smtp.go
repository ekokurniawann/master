package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"backend-skripsi/internal/config"
	"backend-skripsi/internal/view"
)

type smtpMailer struct{}

func NewSMTPMailer() Mailer {
	return &smtpMailer{}
}

func (m *smtpMailer) SendHTML(ctx context.Context, toEmail, toName, subject, htmlContent string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("mailer.smtp.SendHTML: context cancelled before sending: %w", err)
	}

	mailerCfg := config.Get().Mailer
	auth := smtp.PlainAuth("", mailerCfg.Username, mailerCfg.Password, mailerCfg.Host)

	message := fmt.Sprintf("From: %s <%s>\r\n", mailerCfg.SenderName, mailerCfg.SenderEmail) +
		fmt.Sprintf("To: %s <%s>\r\n", toName, toEmail) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n" +
		"\r\n" +
		htmlContent

	addr := fmt.Sprintf("%s:%d", mailerCfg.Host, mailerCfg.Port)

	err := smtp.SendMail(addr, auth, mailerCfg.SenderEmail, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("mailer.smtp.SendHTML: failed to send email via SMTP: %w", err)
	}

	return nil
}

func (m *smtpMailer) SendVerification(ctx context.Context, toEmail, toName, token string) error {
	subject := "Verifikasi Akun FortisFit Kamu"

	appCfg := config.Get().App
	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify?token=%s&email=%s", appCfg.BaseURL, token, toEmail)

	htmlContent, err := view.RenderEmail("verification.html", view.EmailData{
		FullName: toName,
		URL:      verificationURL,
	})
	if err != nil {
		return fmt.Errorf("mailer.smtp.SendVerification: failed to render template: %w", err)
	}

	return m.SendHTML(ctx, toEmail, toName, subject, htmlContent)
}

func (m *smtpMailer) SendPasswordReset(ctx context.Context, toEmail, toName, token string) error {
	subject := "Reset Password Akun FortisFit Kamu"

	appCfg := config.Get().App
	resetURL := fmt.Sprintf("%s/api/v1/auth/reset-password?token=%s&email=%s", appCfg.BaseURL, token, toEmail)

	htmlContent, err := view.RenderEmail("password_reset.html", view.EmailData{
		FullName: toName,
		URL:      resetURL,
	})
	if err != nil {
		return fmt.Errorf("mailer.smtp.SendPasswordReset: failed to render template: %w", err)
	}

	return m.SendHTML(ctx, toEmail, toName, subject, htmlContent)
}
