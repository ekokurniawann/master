package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"backend-skripsi/internal/config"
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

	htmlContent := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <title>Verifikasi Akun FortisFit</title>
        </head>
        <body style="margin: 0; padding: 20px; font-family: sans-serif; background-color: #fafafa; color: #333;">
            <div style="max-width: 600px; margin: 0 auto; background: #ffffff; padding: 30px; border-radius: 8px; box-shadow: 0 4px 10px rgba(0,0,0,0.05);">
                <h2 style="color: #000; margin-bottom: 20px;">Halo %s,</h2>
                <p style="font-size: 16px; line-height: 1.6;">Terima kasih telah melakukan registrasi di <strong>FORTISFIT</strong>.</p>
                <p style="font-size: 16px; line-height: 1.6;">Langkah terakhir sebelum kamu bisa menggunakan aplikasi, silakan klik tombol di bawah ini untuk memverifikasi email kamu:</p>
                
                <div style="text-align: center; margin: 30px 0;">
                    <a href="%s" target="_blank" style="display: inline-block; padding: 14px 28px; background-color: #00FFFF; color: #000000; text-decoration: none; font-weight: bold; font-size: 16px; border-radius: 6px; box-shadow: 0 4px 6px rgba(0,255,255,0.2);">
                        Verifikasi Akun Saya
                    </a>
                </div>
                
                <p style="font-size: 14px; color: #666;">Tautan verifikasi ini hanya berlaku selama <strong>15 menit</strong>.</p>
                <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
                <p style="font-size: 14px; color: #999; line-height: 1.5;">Jika kamu tidak merasa melakukan pendaftaran ini, silakan abaikan email ini dengan aman.<br>Salam hangat,<br><strong>FortisFit Team</strong></p>
            </div>
		</body>
        </html>
    `, toName, verificationURL)

	return m.SendHTML(ctx, toEmail, toName, subject, htmlContent)
}
